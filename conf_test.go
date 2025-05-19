package conf

import "testing"

const (
	testValString       = "Test String"
	testValString1      = "Test String1"
	testValString2      = "Test String2"
	testValString3      = "Test String3"
	testValInt          = 18
	testValMapKey1      = "map_key1"
	testValMapKey2      = "map_key2"
	testValMapKey3      = "map_key3"
	testValStringEnvVar = "TEST_CONF_STRING"
)

type tConfIn struct {
	StringTest       string                   `json:"string_test,omitempty" yaml:"string_test,omitempty"`
	IntTest          int                      `json:"int_test,omitempty" yaml:"int_test,omitempty"`
	StructsTest      tStructTestIn            `json:"struct_test,omitempty" yaml:"struct_test,omitempty"`
	StructsSliceTest []tStructTestIn          `json:"struct_slice_test,omitempty" yaml:"struct_slice_test,omitempty"`
	StructsMapTest   map[string]tStructTestIn `json:"struct_map_test,omitempty" yaml:"struct_map_test,omitempty"`
	StringsSliceTest []string                 `json:"strings_slice_test" yaml:"strings_slice_test"`
}

type tStructTestIn struct {
	StringTest string `json:"string_test,omitempty"  yaml:"string_test,omitempty"`
}

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

func testConfCheck(t *testing.T, c tConfOut) {

	// Check loaded data

	// Check specified string data
	if c.StringTest != testValString {
		t.Fatal("Incorrect loaded data: StringTest")
	}

	// Check default int value
	if c.IntTest != testValInt {
		t.Fatal("Incorrect loaded data: IntTest")
	}

	// Check substruct field
	if c.StructsTest.StringTest != testValString {
		t.Fatal("Incorrect loaded data: StructsTest.StringTest")
	}

	// Check substructs slice size
	if len(c.StructsSliceTest) != 3 {
		t.Fatal("Incorrect loaded data: StructsSliceTest")
	}

	// Check substruct map string field
	if c.StructsMapTest[testValMapKey1].StringTest != testValString1 {
		t.Fatal("Incorrect loaded data: StructsMapTest[map_key1].StringTest")
	}

	// Check substruct map string field ENV data
	if c.StructsMapTest[testValMapKey2].StringTest != testValString2 {
		t.Fatal("Incorrect loaded data: StructsMapTest[map_key2].StringTest")
	}

	// Check substruct map string field default data
	if c.StructsMapTest[testValMapKey3].StringTest != testValString {
		t.Fatal("Incorrect loaded data: StructsMapTest[map_key3].StringTest")
	}

	// Check string slice size
	if len(c.StringsSliceTest) != 3 {
		t.Fatal("Incorrect loaded data: StringsSliceTest")
	}
}
