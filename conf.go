package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

// Available types for loadable config
const (
	ConfigTypeYAML = 0
	ConfigTypeJSON = 1
)

const (
	tagConfName          = "conf"
	tagConfExtraOptsName = "conf_extraopts"
	tagConfRequiredName  = "required"
	tagConfDefaultName   = "default"
)

const (
	regexpEnv = "ENV:(.*)"
)

// ConfigType is a loadable config type
type ConfigType int

// Settings struct contains settings config load
type Settings struct {

	// ConfPath contains the path to config file
	ConfPath string

	// ConfType contains config file type (see `ConfigType` constants)
	ConfType ConfigType

	// WeaklyTypes if true makes "weak" conversions while config file decoding
	// (see: https://godoc.org/github.com/mitchellh/mapstructure#DecoderConfig `WeaklyTypedInput` option)
	WeaklyTypes bool

	// UnknownDeny if true fails with an error if config file contains fields that no matching in the result interface
	UnknownDeny bool

	md mapstructure.Metadata
}

type defaultValue struct {
	value string
	isSet bool
}

// Load reads config
func Load(conf interface{}, s Settings) error {

	// Check `conf` is a pointer
	if reflect.TypeOf(conf).Kind() != reflect.Ptr {
		return fmt.Errorf("config load internal error: `conf` must be a pointer")
	}

	cfgFile, err := os.ReadFile(s.ConfPath)
	if err != nil {
		return fmt.Errorf("config error: %s", err)
	}

	rawConf := make(map[string]any)

	switch s.ConfType {
	case ConfigTypeYAML:
		if err := yaml.Unmarshal(cfgFile, &rawConf); err != nil {
			return fmt.Errorf("config error: %s", err)
		}
	case ConfigTypeJSON:
		if err := json.Unmarshal(cfgFile, &rawConf); err != nil {
			return fmt.Errorf("config error: %s", err)
		}
	default:
		return fmt.Errorf("config error: unknown config type")
	}

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: s.WeaklyTypes,
		Metadata:         &s.md,
		DecodeHook:       s.decodeFromString,
		Result:           conf,
		TagName:          tagConfName,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("config error: %v", err)
	}

	err = decoder.Decode(rawConf)
	if err != nil {
		return fmt.Errorf("config error: %v", err)
	}

	// Set options default values
	if err := s.setDefaults(reflect.ValueOf(conf), "", defaultValue{"", false}); err != nil {
		return fmt.Errorf("config error: %v", err)
	}

	if err := s.checkUsedRequredOpts(reflect.ValueOf(conf), ""); err != nil {
		return fmt.Errorf("config error: %v", err)
	}

	if err := s.checkUnknownOpts(); err != nil {
		return fmt.Errorf("config error: %v", err)
	}

	return nil
}

// setDefaults sets the default values from tags.
func (s *Settings) setDefaults(val reflect.Value, parentName string, dv defaultValue) error {

	if val.Kind() == reflect.Ptr && val.IsNil() == true {
		return nil
	}

	// Check val is pointer
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check val is writable
	if val.CanSet() == false {
		return fmt.Errorf("internal error, object is not writable")
	}

	switch val.Type().Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			vf := val.Field(i)
			tf := val.Type().Field(i)

			elName := parentName
			if elName != "" {
				elName = strings.Join([]string{elName, s.fieldNameNormalize(tf)}, ".")
			} else {
				elName = s.fieldNameNormalize(tf)
			}

			v, isSet := s.tagValGet(tf.Tag.Get(tagConfExtraOptsName), tagConfDefaultName)

			if err := s.setDefaults(vf, elName, defaultValue{v, isSet}); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			vf := val.Index(i)

			elName := fmt.Sprintf("%s[%d]", parentName, i)

			if err := s.setDefaults(vf, elName, defaultValue{"", false}); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, k := range val.MapKeys() {
			vf := val.MapIndex(k)

			// Create copy of element to make it writable
			t := reflect.Indirect(reflect.New(vf.Type()))
			t.Set(reflect.ValueOf(vf.Interface()))

			elName := fmt.Sprintf("%s[%s]", parentName, k)

			if err := s.setDefaults(t, elName, defaultValue{"", false}); err != nil {
				return err
			}

			val.SetMapIndex(k, t)
		}

	default:

		// If default value set for this element and this option not used in conf file, fill it with default value
		if dv.isSet == true && s.optIsUsed(parentName, s.md.Keys) == false {

			d, err := s.convFromString(dv.value, val.Type())
			if err != nil {
				return err
			}

			switch val.Type().Kind() {
			case reflect.Bool:
				val.SetBool(d.(bool))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				val.SetInt(d.(int64))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				val.SetUint(d.(uint64))
			case reflect.Float32, reflect.Float64:
				val.SetFloat(d.(float64))
			case reflect.String:
				val.SetString(d.(string))
			default:
				return fmt.Errorf("internal error, default value not available for this field type `%s`", parentName)
			}
		}
	}

	return nil
}

// checkUsedRequredOpts checks that config file contains all requirement options
func (s *Settings) checkUsedRequredOpts(val reflect.Value, parentName string) error {

	if val.Kind() == reflect.Ptr && val.IsNil() == true {
		return nil
	}

	// Check val is pointer
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Type().Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			vf := val.Field(i)
			tf := val.Type().Field(i)

			elName := parentName
			if elName != "" {
				elName = strings.Join([]string{elName, s.fieldNameNormalize(tf)}, ".")
			} else {
				elName = s.fieldNameNormalize(tf)
			}

			tag := tf.Tag.Get(tagConfExtraOptsName)

			if s.tagKeyCheck(tag, tagConfRequiredName) == true && s.optIsUsed(elName, s.md.Keys) == false {
				return fmt.Errorf("required option '%s' is not specified", elName)
			}

			if err := s.checkUsedRequredOpts(vf, elName); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			vf := val.Index(i)

			elName := fmt.Sprintf("%s[%d]", parentName, i)

			if err := s.checkUsedRequredOpts(vf, elName); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, k := range val.MapKeys() {
			vf := val.MapIndex(k)

			elName := fmt.Sprintf("%s[%s]", parentName, k)

			if err := s.checkUsedRequredOpts(vf, elName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Settings) checkUnknownOpts() error {
	if s.UnknownDeny == true && len(s.md.Unused) > 0 {
		return fmt.Errorf("unknown option '%s'", s.md.Unused[0])
	}
	return nil
}

// decodeFromString decodes values from string to other types.
// Able to use field values in format `ENV:VARIABLE_NAME` to get values from ENV variables.
func (s *Settings) decodeFromString(f reflect.Type, t reflect.Type, v interface{}) (interface{}, error) {

	var str string

	if f.Kind() != reflect.String {
		return v, nil
	}

	var r = regexp.MustCompile(regexpEnv)

	result := r.FindStringSubmatch(v.(string))

	if result != nil {
		str = os.Getenv(result[1])
		if str == "" {
			return v, fmt.Errorf("empty ENV variable '%s'", result[1])
		}
	} else {
		str = v.(string)
	}

	return s.convFromString(str, t)
}

// convFromString converts string value to other type in accordance to `t`
func (s *Settings) convFromString(str string, t reflect.Type) (interface{}, error) {

	switch t.Kind() {
	case reflect.Bool:
		return strconv.ParseBool(str)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(str, 0, t.Bits())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.ParseUint(str, 0, t.Bits())
	case reflect.Float32:
		return strconv.ParseFloat(str, 32)
	case reflect.Float64:
		return strconv.ParseFloat(str, 64)
	}

	return str, nil
}

// fieldNameNormalize returns either name from tag if specified, or struct field name as is
func (s *Settings) fieldNameNormalize(tf reflect.StructField) string {

	tag := tf.Tag.Get(tagConfName)

	str := s.tagValIndexGet(tag, 0)
	if str != "" {
		return str
	}

	return tf.Name
}

// optIsUsed checks that string slice `usedOpts` contains `opt`
func (s *Settings) optIsUsed(opt string, usedOpts []string) bool {

	for _, v := range usedOpts {
		if v == opt {
			return true
		}
	}

	return false
}

// tagPartsMakeMap prepairs map for tag pairs
func (s *Settings) tagPartsMakeMap(tag string) map[string]string {

	tm := make(map[string]string)

	p := strings.Split(tag, ",")

	for _, e := range p {
		s := strings.Split(e, "=")
		if len(s) > 1 {
			tm[strings.Trim(s[0], " \t")] = s[1]
		} else {
			tm[strings.Trim(s[0], " \t")] = ""
		}
	}

	return tm
}

// tagKeyCheck cheks that `tag` contains `key`
func (s *Settings) tagKeyCheck(tag string, key string) bool {

	tm := s.tagPartsMakeMap(tag)

	if _, ok := tm[key]; ok {
		return true
	}

	return false
}

// tagValGet gets from `tag` value for `key`
func (s *Settings) tagValGet(tag string, key string) (string, bool) {

	tm := s.tagPartsMakeMap(tag)

	v, ok := tm[key]
	return v, ok
}

// tagConfGetName gets raw value (without splitting by '=') from tag by index
func (s *Settings) tagValIndexGet(tag string, i int) string {

	p := strings.Split(tag, ",")

	if i >= len(p) {

		return ""
	}

	return p[i]
}
