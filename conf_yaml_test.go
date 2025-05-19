package conf

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

const testYAMLTmpConfPath = "/tmp/nxs-go-conf_test_yaml.conf"

func TestYAMLFormatPath(t *testing.T) {

	var c tConfOut

	// Prepare test config file and fill it with testing data
	testPrepareYAMLConfig(t)
	defer os.Remove(testYAMLTmpConfPath)

	// Load data
	if err := Load(&c, Settings{
		ConfPath:    testYAMLTmpConfPath,
		ConfType:    ConfigTypeYAML,
		WeaklyTypes: false,
		UnknownDeny: true,
	}); err != nil {
		t.Fatal("Config load error:", err)
	}

	// Check data
	testConfCheck(t, c)
}

func TestYAMLFormatBytes(t *testing.T) {

	var c tConfOut

	// Prepare test config file and fill it with testing data
	d := testPrepareYAMLConfig(t)
	defer os.Remove(testYAMLTmpConfPath)

	// Load data
	if err := LoadBytes(&c, SettingsBytes{
		Data:        d,
		ConfType:    ConfigTypeYAML,
		WeaklyTypes: false,
		UnknownDeny: true,
	}); err != nil {
		t.Fatal("Config load error:", err)
	}

	// Check data
	testConfCheck(t, c)
}

func testPrepareYAMLConfig(t *testing.T) []byte {

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

	s, err := yaml.Marshal(&c)
	if err != nil {
		t.Fatal("yaml encode error:", err)
	}

	if err := os.WriteFile(testYAMLTmpConfPath, s, 0644); err != nil {
		t.Fatal("Config file prepare error:", err)
	}

	// Set ENV variables
	os.Setenv(testValStringEnvVar, testValString2)

	return s
}
