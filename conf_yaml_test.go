package conf

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	testYAMLTmpConfPath     = "/tmp/nxs-go-conf_test_yaml.conf"
	testYAMLValString       = "Test String"
	testYAMLValString1      = "Test String1"
	testYAMLValString2      = "Test String2"
	testYAMLValString3      = "Test String3"
	testYAMLValInt          = 123
	testYAMLValMapKey1      = "map_key1"
	testYAMLValMapKey2      = "map_key2"
	testYAMLValMapKey3      = "map_key3"
	testYAMLValStringEnvVar = "TEST_YAML_CONF_STRING"
)

type tConfYAMLIn struct {
	StringTest       string                    `yaml:"string_test,omitempty"`
	IntTest          int                       `yaml:"int_test,omitempty"`
	StructsTest      StructYAMLTest            `yaml:"struct_test,omitempty"`
	StructsSliceTest []StructYAMLTest          `yaml:"struct_slice_test,omitempty"`
	StructsMapTest   map[string]StructYAMLTest `yaml:"struct_map_test,omitempty"`
	StringsSliceTest []string                  `yaml:"strings_slice_test"`
}

type StructYAMLTest struct {
	StringTest string `yaml:"string_test,omitempty"`
}

func TestYAMLFormat(t *testing.T) {

	type tConfOut struct {
		StringTest  string `conf:"string_test" conf_extraopts:"required"`
		IntTest     int    `conf:"int_test" conf_extraopts:"default=18"`
		StructsTest struct {
			StringTest string `conf:"string_test" conf_extraopts:"required"`
		} `conf:"struct_test" conf_extraopts:"required"`
		StructsSliceTest []struct {
			StringTest string `conf:"string_test" conf_extraopts:"default=Test String"`
		} `conf:"struct_slice_test" conf_extraopts:"required"`
		StructsMapTest map[string]struct {
			StringTest string `conf:"string_test" conf_extraopts:"default=Test String"`
		} `conf:"struct_map_test" conf_extraopts:"required"`
		StringsSliceTest []string `conf:"strings_slice_test"`
	}

	var c tConfOut

	// Prepare test config file and fill it with testing data
	testPrepareYAMLConfig(t)

	if err := Load(&c, Settings{
		ConfPath:    testYAMLTmpConfPath,
		ConfType:    ConfigTypeYAML,
		WeaklyTypes: false,
		UnknownDeny: true,
	}); err != nil {
		t.Fatal("Config load error:", err)
	}

	// Remove test config file
	os.Remove(testYAMLTmpConfPath)

	// Check loaded data

	// Check specified string data
	if c.StringTest != testYAMLValString {
		t.Fatal("Incorrect loaded data: StringTest")
	}

	// Check default int value
	if c.IntTest != testYAMLValInt {
		t.Fatal("Incorrect loaded data: IntTest")
	}

	// Check substruct field
	if c.StructsTest.StringTest != testYAMLValString {
		t.Fatal("Incorrect loaded data: StructsTest.StringTest")
	}

	// Check substructs slice size
	if len(c.StructsSliceTest) != 3 {
		t.Fatal("Incorrect loaded data: StructsSliceTest")
	}

	// Check substruct map string field
	if c.StructsMapTest[testYAMLValMapKey1].StringTest != testYAMLValString1 {
		t.Fatal("Incorrect loaded data: StructsMapTest[map_key1].StringTest")
	}

	// Check substruct map string field ENV data
	if c.StructsMapTest[testYAMLValMapKey2].StringTest != testYAMLValString2 {
		t.Fatal("Incorrect loaded data: StructsMapTest[map_key2].StringTest")
	}

	// Check substruct map string field default data
	if c.StructsMapTest[testYAMLValMapKey3].StringTest != testYAMLValString {
		t.Fatal("Incorrect loaded data: StructsMapTest[map_key3].StringTest")
	}

	// Check string slice size
	if len(c.StringsSliceTest) != 3 {
		t.Fatal("Incorrect loaded data: StringsSliceTest")
	}
}

func testPrepareYAMLConfig(t *testing.T) {

	c := tConfYAMLIn{
		StringTest: testYAMLValString,
		IntTest:    testYAMLValInt,
		StructsTest: StructYAMLTest{
			StringTest: testYAMLValString,
		},
		StructsSliceTest: []StructYAMLTest{
			{
				StringTest: testYAMLValString1,
			},
			{
				StringTest: testYAMLValString2,
			},
			{
				StringTest: testYAMLValString3,
			},
		},
		StructsMapTest: map[string]StructYAMLTest{
			testYAMLValMapKey1: {
				StringTest: testYAMLValString1,
			},
			testYAMLValMapKey2: {
				StringTest: "ENV:" + testYAMLValStringEnvVar,
			},
			testYAMLValMapKey3: {},
		},
		StringsSliceTest: []string{
			testYAMLValString1,
			testYAMLValString2,
			testYAMLValString3,
		},
	}

	s, err := yaml.Marshal(&c)
	if err != nil {
		t.Fatal("Yaml encode error:", err)
	}

	if err := os.WriteFile(testYAMLTmpConfPath, s, 0644); err != nil {
		t.Fatal("Config file prepare error:", err)
	}

	// Set ENV variables
	os.Setenv(testYAMLValStringEnvVar, testYAMLValString2)
}
