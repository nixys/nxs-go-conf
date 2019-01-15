package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

const (
	testJSONTmpConfPath        = "/tmp/nxs-go-conf_test_json.conf"
	testJSONValName            = "Test JSON Name"
	testJSONValAge             = 18
	testJSONValJobAddress      = "Test JSON Address"
	testJSONValJobNameEnvVar   = "TEST_JSON_CONF_JOB_NAME"
	testJSONValJobNameEnvVal   = "Test JSON Job name"
	testJSONValJobSalaryEnvVar = "TEST_JSON_CONF_JOB_SALARY"
	testJSONValJobSalaryEnvVal = 1.2
)

type tConfJSONIn struct {
	Name           string         `json:"name,omitempty"`
	Age            int            `json:"age,omitempty"`
	Job            tConfJSONInJob `json:"job,omitempty"`
	FavoriteDishes []string       `json:"favorite_dishes"`
}

type tConfJSONInJob struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
	Salary  string `json:"salary,omitempty"`
}

func TestJSONFormat(t *testing.T) {

	type tConfOut struct {
		Name string `conf:"name" conf_extraopts:"required"`
		Age  int    `conf_extraopts:"default=18"`
		Job  struct {
			Name    string  `conf:"name" conf_extraopts:"required"`
			Address string  `conf:"address" conf_extraopts:"default=Test JSON Address"`
			Salary  float64 `conf:"salary" conf_extraopts:"default=1.1"`
		} `conf:"job" conf_extraopts:"required"`
		FavoriteDishes []string `conf:"favorite_dishes"`
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
	if c.Name != testJSONValName {
		t.Fatal("Incorrect loaded data: Name")
	}

	// Check default int value
	if c.Age != testJSONValAge {
		t.Fatal("Incorrect loaded data: Age")
	}

	// Check specified string ENV data
	if c.Job.Name != testJSONValJobNameEnvVal {
		t.Fatal("Incorrect loaded data: Job.Name")
	}

	// Check default string value
	if c.Job.Address != testJSONValJobAddress {
		t.Fatal("Incorrect loaded data: Job.Address")
	}

	// Check specified float ENV data
	if c.Job.Salary != testJSONValJobSalaryEnvVal {
		t.Fatal("Incorrect loaded data: Job.Salary")
	}

	// Check string slices
	if len(c.FavoriteDishes) != 2 {
		t.Fatal("Incorrect loaded data: FavoriteDishes")
	}
}

func testPrepareJSONConfig(t *testing.T) {

	c := tConfJSONIn{
		Name: testJSONValName,
		Job: tConfJSONInJob{
			Name:   "ENV:" + testJSONValJobNameEnvVar,
			Salary: "ENV:" + testJSONValJobSalaryEnvVar,
		},
		FavoriteDishes: []string{"apples", "ice cream"},
	}

	s, err := json.Marshal(&c)
	if err != nil {
		t.Fatal("Json encode error:", err)
	}

	if err := ioutil.WriteFile(testJSONTmpConfPath, s, 0644); err != nil {
		t.Fatal("Config file prepare error:", err)
	}

	// Set ENV variables
	os.Setenv(testJSONValJobNameEnvVar, testJSONValJobNameEnvVal)
	os.Setenv(testJSONValJobSalaryEnvVar, strconv.FormatFloat(testJSONValJobSalaryEnvVal, 'f', 3, 64))
}
