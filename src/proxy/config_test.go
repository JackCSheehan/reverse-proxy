package proxy

import (
	"errors"
	"io/fs"
	"reflect"
	"testing"
)

// Test constructing a new config object from valid configs
func TestNewConfigGood(t *testing.T) {
	actual, err := NewConfig("test_good.yaml")
	expected := Config{
		Port: 9001,
		Endpoints: []Endpoint{
			{
				From: "/home",
				Pool: []string{
					"http://localhost:1001/about",
					"http://localhost:1002/info",
					"http://localhost:1003/contact",
				},
			},
			{
				From: "/",
				Pool: []string{
					"http://localhost:1001",
					"http://localhost:1002",
				},
			},
			{
				From: "/post",
				Pool: []string{
					"http://localhost:1001/comments",
				},
			},
		},
	}

	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if reflect.DeepEqual(*actual, expected) {
		t.Fatalf("Expected %+v but got %+v", expected, actual)
	}
}

// Tests calling NewConfig with a missing file
func TestNewConfigMissingFile(t *testing.T) {
	config, err := NewConfig("this_config_does_not_exist.yaml")

	var expectedError *fs.PathError
	if !errors.As(err, &expectedError) {
		t.Fatalf("Expected %T error, but got %T instead", expectedError, err)
	}

	if config != nil {
		t.Fatalf("Expected to get a nil config, but got %v instead", config)
	}
}

// Tests calling NewConfig with an invalid YAML
func TestNewConfigInvalidSyntax(t *testing.T) {
	config, err := NewConfig("test_bad.yaml")

	expectedErrorMessage := "yaml: line 2: could not find expected ':'"

	if err.Error() != expectedErrorMessage {
		t.Fatalf("Expected '%s' error, but got '%s' instead", expectedErrorMessage, err)
	}

	if config != nil {
		t.Fatalf("Expected to get a nil config, but got %v instead", config)
	}
}

// Tests calling NewConfig on a config file with missing keys
func TestNewConfigMissingKeys(t *testing.T) {
	// Test a missing top-level config key
	config, err := NewConfig("test_missing_port.yaml")

	var expectedError *MissingRequiredFieldError
	if !errors.As(err, &expectedError) {
		t.Fatalf("Expected %T error, but got %T instead", expectedError, err)
	}

	if config != nil {
		t.Fatalf("Expected to get a nil config, but got %v instead", config)
	}

	if expectedError.fieldName != "port" && expectedError.endpointId != -1 {
		t.Fatalf("Missing port error was not created with the expected field values")
	}

	// Test a missing key inside an endpoint definition
	config, err = NewConfig("test_missing_from.yaml")

	if !errors.As(err, &expectedError) {
		t.Fatalf("Expected %T error, but got %T instead", expectedError, err)
	}

	if config != nil {
		t.Fatalf("Expected to get a nil config, but got %v instead", config)
	}

	if expectedError.fieldName != "from" && expectedError.endpointId != 1 {
		t.Fatalf("Missing port error was not created with the expected field values")
	}

	// Test an endpoint missing a pool
	config, err = NewConfig("test_missing_pool.yaml")

	if !errors.As(err, &expectedError) {
		t.Fatalf("Expected %T error, but got %T instead", expectedError, err)
	}

	if config != nil {
		t.Fatalf("Expected to get a nil config, but got %v instead", config)
	}

	if expectedError.fieldName != "pool" && expectedError.endpointId != 1 {
		t.Fatalf("Missing port error was not created with the expected field values")
	}
}
