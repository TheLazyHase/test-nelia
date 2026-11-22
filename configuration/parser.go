package configuration

import (
	"encoding/json"
	"fmt"
	"os"
)

func ParseFile(file string) (Configuration, error) {
	fileContent, err := os.ReadFile(file)
	var conf Configuration
	if err != nil {
		return conf, fmt.Errorf("Error reading configuration file: %w", err)
	}

	err = json.Unmarshal(fileContent, &conf)
	if err != nil {
		return conf, fmt.Errorf("Error parsing json configuration file: %w", err)
	}
	return conf, nil
}
