package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

// ConfigType is a loadable config type
type ConfigType int

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

// Load reads config
func Load(conf interface{}, s Settings) error {

	var md mapstructure.Metadata

	// Check `conf` is a pointer
	if reflect.TypeOf(conf).Kind() != reflect.Ptr {
		return fmt.Errorf("config load internal error: `conf` must be a pointer")
	}

	cfgFile, err := ioutil.ReadFile(s.ConfPath)
	if err != nil {
		return fmt.Errorf("config error: %s", err)
	}

	rawConf := make(map[string]interface{})

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

	// Set options default values
	if err := setDefaults(reflect.ValueOf(conf)); err != nil {
		return fmt.Errorf("config error: %v", err)
	}

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: s.WeaklyTypes,
		Metadata:         &md,
		DecodeHook:       decodeFromString,
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

	if err := checkUsedRequredOpts(reflect.ValueOf(conf), "", md.Keys); err != nil {
		return fmt.Errorf("config error: %v", err)
	}

	if s.UnknownDeny == true && len(md.Unused) > 0 {
		return fmt.Errorf("config error: unknown option '%s'", md.Unused[0])
	}

	return nil
}

// decodeFromString decodes values from string to other types.
// Able to use field values in format `ENV:VARIABLE_NAME` to get values from ENV variables.
func decodeFromString(f reflect.Type, t reflect.Type, v interface{}) (interface{}, error) {

	var s string

	if f.Kind() != reflect.String {
		return v, nil
	}

	var r = regexp.MustCompile(regexpEnv)

	result := r.FindStringSubmatch(v.(string))

	if result != nil {
		s = os.Getenv(result[1])
		if s == "" {
			return v, fmt.Errorf("empty ENV variable '%s'", result[1])
		}
	} else {
		s = v.(string)
	}

	return convFromString(s, t)
}

// convFromString converts string value to other type in accordance to `t`
func convFromString(s string, t reflect.Type) (interface{}, error) {

	switch t.Kind() {
	case reflect.Bool:
		return strconv.ParseBool(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(s, 0, t.Bits())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.ParseUint(s, 0, t.Bits())
	case reflect.Float32:
		return strconv.ParseFloat(s, 32)
	case reflect.Float64:
		return strconv.ParseFloat(s, 64)
	}

	return s, nil
}

// setDefaults sets the default values from tags.
// Only for _Int*_, _Uint*_, _Bool_ and _String_ (not within the arrays, maps or slices) types default values are available.
func setDefaults(val reflect.Value) error {

	// Check val is pointer
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check val is struct
	if val.Type().Kind() != reflect.Struct {
		return fmt.Errorf("internal error, must be a pointer to struct")
	}

	// Check val is writable
	if val.CanSet() == false {
		return fmt.Errorf("internal error, object is not writable")
	}

	for i := 0; i < val.NumField(); i++ {

		vf := val.Field(i)
		tf := val.Type().Field(i)

		if s, ok := tagValGet(tf.Tag.Get(tagConfExtraOptsName), tagConfDefaultName); ok == true {

			d, err := convFromString(s, tf.Type)
			if err != nil {
				return err
			}

			switch tf.Type.Kind() {
			case reflect.Bool:
				vf.SetBool(d.(bool))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				vf.SetInt(d.(int64))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				vf.SetUint(d.(uint64))
			case reflect.Float32, reflect.Float64:
				vf.SetFloat(d.(float64))
			case reflect.String:
				vf.SetString(d.(string))
			default:
				return fmt.Errorf("internal error, default value not available for this field `%s`", tf.Name)
			}
		}

		if vf.Kind() == reflect.Struct {
			if err := setDefaults(vf); err != nil {
				return err
			}
		}
	}

	return nil
}

// fieldNameNormalize returns either name from tag if specified, or struct field name as is
func fieldNameNormalize(tf reflect.StructField) string {

	tag := tf.Tag.Get(tagConfName)

	s := tagValIndexGet(tag, 0)
	if s != "" {
		return s
	}

	return tf.Name
}

// checkUsedRequredOpts checks that config file contains all requirement options
func checkUsedRequredOpts(val reflect.Value, parentName string, usedOpts []string) error {

	// Check val is pointer
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check val is struct
	if val.Type().Kind() != reflect.Struct {
		return fmt.Errorf("must be a pointer to struct")
	}

	for i := 0; i < val.NumField(); i++ {

		vf := val.Field(i)
		tf := val.Type().Field(i)

		s := parentName

		if s != "" {
			s = strings.Join([]string{s, fieldNameNormalize(tf)}, ".")
		} else {
			s = fieldNameNormalize(tf)
		}

		tag := tf.Tag.Get(tagConfExtraOptsName)
		if tagKeyCheck(tag, tagConfRequiredName) == true && optIsUsed(s, usedOpts) == false {
			return fmt.Errorf("required option '%s' is not specified", s)
		}

		if vf.Kind() == reflect.Struct {
			if err := checkUsedRequredOpts(vf, s, usedOpts); err != nil {
				return err
			}
		}
	}

	return nil
}

// optIsUsed checks that string slice `usedOpts` contains `opt`
func optIsUsed(opt string, usedOpts []string) bool {

	for _, v := range usedOpts {
		if v == opt {
			return true
		}
	}

	return false
}

// tagPartsMakeMap prepairs map for tag pairs
func tagPartsMakeMap(tag string) map[string]string {

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
func tagKeyCheck(tag string, key string) bool {

	tm := tagPartsMakeMap(tag)

	if _, ok := tm[key]; ok {
		return true
	}

	return false
}

// tagValGet gets from `tag` value for `key`
func tagValGet(tag string, key string) (string, bool) {

	tm := tagPartsMakeMap(tag)

	v, ok := tm[key]
	return v, ok
}

// tagConfGetName gets raw value (without splitting by '=') from tag by index
func tagValIndexGet(tag string, i int) string {

	p := strings.Split(tag, ",")

	if i >= len(p) {

		return ""
	}

	return p[i]
}
