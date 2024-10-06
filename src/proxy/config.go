package proxy

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// Error used to indicate a missing required field
type MissingRequiredFieldError struct {
	// Name of the missing field
	fieldName string

	// ID of the endpoint in which the missing field is located
	endpointId int
}

func (e *MissingRequiredFieldError) Error() string {
	// If no endpoint was defined, return a simple error message
	if e.endpointId == -1 {
		return fmt.Sprintf("%s is a required field", e.fieldName)
	}

	// If endpoint was defined, include it in the error message
	return fmt.Sprintf("%s is a required field, but it is missing from endpoint ID %d", e.fieldName, e.endpointId)
}

func NewMissingRequiredFieldError(fieldName string, endpointId int) *MissingRequiredFieldError {
	return &MissingRequiredFieldError{
		fieldName:  fieldName,
		endpointId: endpointId,
	}
}

// Maps a particular endpoint to a list of other endpoints
type Endpoint struct {
	From string   `yaml:from`
	Pool []string `yaml:Pool`
}

// Reverse proxy configuration
type Config struct {
	Port      int        `yaml:port`
	Endpoints []Endpoint `yaml:endpoints`
}

// Accepts a path to a config file, parses it, and constructs a new Config instance
func NewConfig(configFilePath string) (*Config, error) {
	configFileContents, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	// Default value for port lets us know if the user didn't include a port field
	config := Config{Port: -1}

	err = yaml.Unmarshal(configFileContents, &config)
	if err != nil {
		return nil, err
	}

	// Validate the config
	if config.Port == -1 {
		return nil, NewMissingRequiredFieldError("port", -1)
	}

	for i, endpoint := range config.Endpoints {
		if endpoint.From == "" {
			return nil, NewMissingRequiredFieldError("from", i)
		}

		if len(endpoint.Pool) == 0 {
			return nil, NewMissingRequiredFieldError("pool", i)
		}
	}

	return &config, nil
}
