package conf

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"gopkg.in/yaml.v2"
)

const (
	testYAMLTmpConfPath        = "/tmp/nxs-go-conf_test_yaml.conf"
	testYAMLValName            = "Test YAML Name"
	testYAMLValAge             = 19
	testYAMLValJobAddress      = "Test YAML Address"
	testYAMLValJobNameEnvVar   = "TEST_YAML_CONF_JOB_NAME"
	testYAMLValJobNameEnvVal   = "Test YAML Job name"
	testYAMLValJobSalaryEnvVar = "TEST_YAML_CONF_JOB_SALARY"
	testYAMLValJobSalaryEnvVal = 1.4
)

type tConfYAMLIn struct {
	Name           string         `yaml:"name,omitempty"`
	Age            int            `yaml:"age,omitempty"`
	Job            tConfYAMLInJob `yaml:"job,omitempty"`
	FavoriteDishes []string       `yaml:"favorite_dishes"`
}

type tConfYAMLInJob struct {
	Name    string `yaml:"name,omitempty"`
	Address string `yaml:"address,omitempty"`
	Salary  string `yaml:"salary,omitempty"`
}

func TestYAMLFormat(t *testing.T) {

	type tConfOut struct {
		Name string `conf:"name" conf_extraopts:"required"`
		Age  int    `conf_extraopts:"default=19"`
		Job  struct {
			Name    string  `conf:"name" conf_extraopts:"required"`
			Address string  `conf:"address" conf_extraopts:"default=Test YAML Address"`
			Salary  float64 `conf:"salary" conf_extraopts:"default=1.3"`
		} `conf:"job" conf_extraopts:"required"`
		FavoriteDishes []string `conf:"favorite_dishes"`
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
	if c.Name != testYAMLValName {
		t.Fatal("Incorrect loaded data: Name")
	}

	// Check default int value
	if c.Age != testYAMLValAge {
		t.Fatal("Incorrect loaded data: Age")
	}

	// Check specified string ENV data
	if c.Job.Name != testYAMLValJobNameEnvVal {
		t.Fatal("Incorrect loaded data: Job.Name")
	}

	// Check default string value
	if c.Job.Address != testYAMLValJobAddress {
		t.Fatal("Incorrect loaded data: Job.Address")
	}

	// Check specified float ENV data
	if c.Job.Salary != testYAMLValJobSalaryEnvVal {
		t.Fatal("Incorrect loaded data: Job.Salary")
	}

	// Check string slices
	if len(c.FavoriteDishes) != 2 {
		t.Fatal("Incorrect loaded data: FavoriteDishes")
	}
}

func testPrepareYAMLConfig(t *testing.T) {

	c := tConfYAMLIn{
		Name: testYAMLValName,
		Job: tConfYAMLInJob{
			Name:   "ENV:" + testYAMLValJobNameEnvVar,
			Salary: "ENV:" + testYAMLValJobSalaryEnvVar,
		},
		FavoriteDishes: []string{"apples", "ice cream"},
	}

	s, err := yaml.Marshal(&c)
	if err != nil {
		t.Fatal("Yaml encode error:", err)
	}

	if err := ioutil.WriteFile(testYAMLTmpConfPath, s, 0644); err != nil {
		t.Fatal("Config file prepare error:", err)
	}

	// Set ENV variables
	os.Setenv(testYAMLValJobNameEnvVar, testYAMLValJobNameEnvVal)
	os.Setenv(testYAMLValJobSalaryEnvVar, strconv.FormatFloat(testYAMLValJobSalaryEnvVal, 'f', 3, 64))
}
