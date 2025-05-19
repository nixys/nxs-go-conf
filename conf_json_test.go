package conf

import (
	"encoding/json"
	"os"
	"testing"
)

const testJSONTmpConfPath = "/tmp/nxs-go-conf_test_json.conf"

func TestJSONFormatPath(t *testing.T) {

	var c tConfOut

	// Prepare test config file and fill it with testing data
	testPrepareJSONConfig(t)
	defer os.Remove(testJSONTmpConfPath)

	// Load data
	if err := Load(&c, Settings{
		ConfPath:    testJSONTmpConfPath,
		ConfType:    ConfigTypeJSON,
		WeaklyTypes: false,
		UnknownDeny: true,
	}); err != nil {
		t.Fatal("Config load error:", err)
	}

	// Check data
	testConfCheck(t, c)
}

func TestJSONFormatBytes(t *testing.T) {

	var c tConfOut

	// Prepare test config file and fill it with testing data
	d := testPrepareJSONConfig(t)
	defer os.Remove(testJSONTmpConfPath)

	// Load data
	if err := LoadBytes(&c, SettingsBytes{
		Data:        d,
		ConfType:    ConfigTypeJSON,
		WeaklyTypes: false,
		UnknownDeny: true,
	}); err != nil {
		t.Fatal("Config load error:", err)
	}

	// Check data
	testConfCheck(t, c)
}

func testPrepareJSONConfig(t *testing.T) []byte {

	c := tConfIn{
		StringTest: testValString,
		StructsTest: tStructTestIn{
			StringTest: testValString,
		},
		StructsSliceTest: []tStructTestIn{
			{
				StringTest: testValString1,
			},
			{
				StringTest: testValString2,
			},
			{
				StringTest: testValString3,
			},
		},
		StructsMapTest: map[string]tStructTestIn{
			testValMapKey1: {
				StringTest: testValString1,
			},
			testValMapKey2: {
				StringTest: "ENV:" + testValStringEnvVar,
			},
			testValMapKey3: {},
		},
		StringsSliceTest: []string{
			testValString1,
			testValString2,
			testValString3,
		},
	}

	s, err := json.Marshal(&c)
	if err != nil {
		t.Fatal("json encode error:", err)
	}

	if err := os.WriteFile(testJSONTmpConfPath, s, 0644); err != nil {
		t.Fatal("Config file prepare error:", err)
	}

	// Set ENV variables
	os.Setenv(testValStringEnvVar, testValString2)

	return s
}
