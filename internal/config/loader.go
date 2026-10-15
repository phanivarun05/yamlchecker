package config

import (
	"fmt"
	"log/slog"
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
		slog.Error("unmarshalling failed", "path", path)
		return nil, err
	}
	return result, nil
}
