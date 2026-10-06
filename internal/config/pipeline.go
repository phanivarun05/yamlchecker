package config

import (
	"fmt"
	"yamlchecker-cli/internal/scanner"
)

type LoadedFile struct {
	Path string
	Data map[string]interface{}
	Err  error
}

func LoadAll(dir string) ([]LoadedFile, error) {
	files, err := scanner.FindYAMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("Error Scanning Directory: %w", err)
	}
	var loaded []LoadedFile
	for _, file := range files {
		data, err := Load(file)
		loaded = append(loaded, LoadedFile{Path: file, Data: data, Err: err})
	}
	return loaded, nil
}
