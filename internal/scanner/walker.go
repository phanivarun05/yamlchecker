package scanner

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

func FindYAMLFiles(rootDir string) ([]string, error) {
	var results []string
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(d.Name())
		if ext == ".yaml" || ext == ".yml" {
			results = append(results, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", rootDir, err)
	}
	return results, nil
}
