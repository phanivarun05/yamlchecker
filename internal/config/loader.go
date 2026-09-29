package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (map[string]interface{}, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Parsing %s: %w", path, err)
	}

	var result map[string]interface{}

	if err := yaml.Unmarshal(bytes, &result); err != nil {
		return nil, fmt.Errorf("unmarshalling %s: %w", path, err)
	}
	return result, nil
}
