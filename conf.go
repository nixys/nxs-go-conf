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
}

// SettingsBytes struct contains settings config load from bytes
type SettingsBytes struct {

	// Data contains config data
	Data []byte

	// ConfType contains config file type (see `ConfigType` constants)
	ConfType ConfigType

	// WeaklyTypes if true makes "weak" conversions while config file decoding
	// (see: https://godoc.org/github.com/mitchellh/mapstructure#DecoderConfig `WeaklyTypedInput` option)
	WeaklyTypes bool

	// UnknownDeny if true fails with an error if config file contains fields that no matching in the result interface
	UnknownDeny bool
}

type conf struct {
	md mapstructure.Metadata
}

type opts struct {
	confType    ConfigType
	weaklyTypes bool
	unknownDeny bool
}

type defaultValue struct {
	value string
	isSet bool
}

// Load reads config from file
func Load(out any, s Settings) error {

	d, err := os.ReadFile(s.ConfPath)
	if err != nil {
		return fmt.Errorf("config error: %s", err)
	}

	if err := confRead(
		out,
		d,
		opts{
			confType:    s.ConfType,
			weaklyTypes: s.WeaklyTypes,
			unknownDeny: s.UnknownDeny,
		},
	); err != nil {
		return fmt.Errorf("config error: %s", err)
	}

	return nil
}

// LoadBytes reads config from bytes
func LoadBytes(out any, s SettingsBytes) error {

	if err := confRead(
		out,
		s.Data,
		opts{
			confType:    s.ConfType,
			weaklyTypes: s.WeaklyTypes,
			unknownDeny: s.UnknownDeny,
		},
	); err != nil {
		return fmt.Errorf("config error: %s", err)
	}

	return nil
}

// confRead reads config
func confRead(out any, d []byte, o opts) error {

	var c conf

	// Check `r` is a pointer
	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		return fmt.Errorf("`out` must be a pointer")
	}

	rawConf := make(map[string]any)

	switch o.confType {
	case ConfigTypeYAML:
		if err := yaml.Unmarshal(d, &rawConf); err != nil {
			return err
		}
	case ConfigTypeJSON:
		if err := json.Unmarshal(d, &rawConf); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown config type")
	}

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: o.weaklyTypes,
		Metadata:         &c.md,
		DecodeHook:       c.decodeFromString,
		Result:           out,
		TagName:          tagConfName,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	err = decoder.Decode(rawConf)
	if err != nil {
		return err
	}

	// Set options default values
	if err := c.setDefaults(reflect.ValueOf(out), "", defaultValue{"", false}); err != nil {
		return err
	}

	if err := c.checkUsedRequredOpts(reflect.ValueOf(out), ""); err != nil {
		return err
	}

	if err := c.checkUnknownOpts(o.unknownDeny); err != nil {
		return err
	}

	return nil
}

// setDefaults sets the default values from tags.
func (cnf *conf) setDefaults(val reflect.Value, parentName string, dv defaultValue) error {

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
				elName = strings.Join([]string{elName, cnf.fieldNameNormalize(tf)}, ".")
			} else {
				elName = cnf.fieldNameNormalize(tf)
			}

			v, isSet := cnf.tagValGet(tf.Tag.Get(tagConfExtraOptsName), tagConfDefaultName)

			if err := cnf.setDefaults(vf, elName, defaultValue{v, isSet}); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			vf := val.Index(i)

			elName := fmt.Sprintf("%s[%d]", parentName, i)

			if err := cnf.setDefaults(vf, elName, defaultValue{"", false}); err != nil {
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

			if err := cnf.setDefaults(t, elName, defaultValue{"", false}); err != nil {
				return err
			}

			val.SetMapIndex(k, t)
		}

	default:

		// If default value set for this element and this option not used in conf file, fill it with default value
		if dv.isSet == true && cnf.optIsUsed(parentName, cnf.md.Keys) == false {

			d, err := cnf.convFromString(dv.value, val.Type())
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
func (cnf *conf) checkUsedRequredOpts(val reflect.Value, parentName string) error {

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
				elName = strings.Join([]string{elName, cnf.fieldNameNormalize(tf)}, ".")
			} else {
				elName = cnf.fieldNameNormalize(tf)
			}

			tag := tf.Tag.Get(tagConfExtraOptsName)

			if cnf.tagKeyCheck(tag, tagConfRequiredName) == true && cnf.optIsUsed(elName, cnf.md.Keys) == false {
				return fmt.Errorf("required option '%s' is not specified", elName)
			}

			if err := cnf.checkUsedRequredOpts(vf, elName); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			vf := val.Index(i)

			elName := fmt.Sprintf("%s[%d]", parentName, i)

			if err := cnf.checkUsedRequredOpts(vf, elName); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, k := range val.MapKeys() {
			vf := val.MapIndex(k)

			elName := fmt.Sprintf("%s[%s]", parentName, k)

			if err := cnf.checkUsedRequredOpts(vf, elName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cnf *conf) checkUnknownOpts(unknownDeny bool) error {
	if unknownDeny == true && len(cnf.md.Unused) > 0 {
		return fmt.Errorf("unknown option '%s'", cnf.md.Unused[0])
	}
	return nil
}

// decodeFromString decodes values from string to other types.
// Able to use field values in format `ENV:VARIABLE_NAME` to get values from ENV variables.
func (cnf *conf) decodeFromString(f reflect.Type, t reflect.Type, v any) (any, error) {

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

	return cnf.convFromString(str, t)
}

// convFromString converts string value to other type in accordance to `t`
func (cnf *conf) convFromString(str string, t reflect.Type) (any, error) {

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
func (cnf *conf) fieldNameNormalize(tf reflect.StructField) string {

	tag := tf.Tag.Get(tagConfName)

	str := cnf.tagValIndexGet(tag, 0)
	if str != "" {
		return str
	}

	return tf.Name
}

// optIsUsed checks that string slice `usedOpts` contains `opt`
func (cnf *conf) optIsUsed(opt string, usedOpts []string) bool {

	for _, v := range usedOpts {
		if v == opt {
			return true
		}
	}

	return false
}

// tagPartsMakeMap prepairs map for tag pairs
func (cnf *conf) tagPartsMakeMap(tag string) map[string]string {

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
func (cnf *conf) tagKeyCheck(tag string, key string) bool {

	tm := cnf.tagPartsMakeMap(tag)

	if _, ok := tm[key]; ok {
		return true
	}

	return false
}

// tagValGet gets from `tag` value for `key`
func (cnf *conf) tagValGet(tag string, key string) (string, bool) {

	tm := cnf.tagPartsMakeMap(tag)

	v, ok := tm[key]
	return v, ok
}

// tagConfGetName gets raw value (without splitting by '=') from tag by index
func (cnf *conf) tagValIndexGet(tag string, i int) string {

	p := strings.Split(tag, ",")

	if i >= len(p) {

		return ""
	}

	return p[i]
}
