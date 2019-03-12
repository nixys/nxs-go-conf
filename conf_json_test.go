package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

const (
	testJSONTmpConfPath     = "/tmp/nxs-go-conf_test_json.conf"
	testJSONValString       = "Test String"
	testJSONValString1      = "Test String1"
	testJSONValString2      = "Test String2"
	testJSONValString3      = "Test String3"
	testJSONValInt          = 123
	testJSONValMapKey1      = "map_key1"
	testJSONValMapKey2      = "map_key2"
	testJSONValMapKey3      = "map_key3"
	testJSONValStringEnvVar = "TEST_JSON_CONF_STRING"
)

type tConfJSONIn struct {
	StringTest       string                    `json:"string_test,omitempty"`
	IntTest          int                       `json:"int_test,omitempty"`
	StructsTest      StructJSONTest            `json:"struct_test,omitempty"`
	StructsSliceTest []StructJSONTest          `json:"struct_slice_test,omitempty"`
	StructsMapTest   map[string]StructJSONTest `json:"struct_map_test,omitempty"`
	StringsSliceTest []string                  `json:"strings_slice_test"`
}

type StructJSONTest struct {
	StringTest string `json:"string_test,omitempty"`
}

func TestJSONFormat(t *testing.T) {

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
	testPrepareJSONConfig(t)

	if err := Load(&c, Settings{
		ConfPath:    testJSONTmpConfPath,
		ConfType:    ConfigTypeJSON,
		WeaklyTypes: false,
		UnknownDeny: true,
	}); err != nil {
		t.Fatal("Config load error:", err)
	}

	// Remove test config file
	os.Remove(testJSONTmpConfPath)

	// Check loaded data

	// Check specified string data
	if c.StringTest != testJSONValString {
		t.Fatal("Incorrect loaded data: StringTest")
	}

	// Check default int value
	if c.IntTest != testJSONValInt {
		t.Fatal("Incorrect loaded data: IntTest")
	}

	// Check substruct field
	if c.StructsTest.StringTest != testJSONValString {
		t.Fatal("Incorrect loaded data: StructsTest.StringTest")
	}

	// Check substructs slice size
	if len(c.StructsSliceTest) != 3 {
		t.Fatal("Incorrect loaded data: StructsSliceTest")
	}

	// Check substruct map string field
	if c.StructsMapTest[testJSONValMapKey1].StringTest != testJSONValString1 {
		t.Fatal("Incorrect loaded data: StructsMapTest[map_key1].StringTest")
	}

	// Check substruct map string field ENV data
	if c.StructsMapTest[testJSONValMapKey2].StringTest != testJSONValString2 {
		t.Fatal("Incorrect loaded data: StructsMapTest[map_key2].StringTest")
	}

	// Check substruct map string field default data
	if c.StructsMapTest[testJSONValMapKey3].StringTest != testJSONValString {
		t.Fatal("Incorrect loaded data: StructsMapTest[map_key3].StringTest")
	}

	// Check string slice size
	if len(c.StringsSliceTest) != 3 {
		t.Fatal("Incorrect loaded data: StringsSliceTest")
	}
}

func testPrepareJSONConfig(t *testing.T) {

	c := tConfJSONIn{
		StringTest: testJSONValString,
		IntTest:    testJSONValInt,
		StructsTest: StructJSONTest{
			StringTest: testJSONValString,
		},
		StructsSliceTest: []StructJSONTest{
			{
				StringTest: testJSONValString1,
			},
			{
				StringTest: testJSONValString2,
			},
			{
				StringTest: testJSONValString3,
			},
		},
		StructsMapTest: map[string]StructJSONTest{
			testJSONValMapKey1: StructJSONTest{
				StringTest: testJSONValString1,
			},
			testJSONValMapKey2: StructJSONTest{
				StringTest: "ENV:" + testJSONValStringEnvVar,
			},
			testJSONValMapKey3: StructJSONTest{},
		},
		StringsSliceTest: []string{
			testJSONValString1,
			testJSONValString2,
			testJSONValString3,
		},
	}

	s, err := json.Marshal(&c)
	if err != nil {
		t.Fatal("Json encode error:", err)
	}

	if err := ioutil.WriteFile(testJSONTmpConfPath, s, 0644); err != nil {
		t.Fatal("Config file prepare error:", err)
	}

	// Set ENV variables
	os.Setenv(testJSONValStringEnvVar, testJSONValString2)
}
