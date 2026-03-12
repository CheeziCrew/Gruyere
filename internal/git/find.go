package git

import (
	"os"
	"path/filepath"
	"strings"
)

// FindOpenAPIFiles walks root looking for openapi.yaml files, skipping target/ directories.
func FindOpenAPIFiles(root string) ([]string, error) {
	var found []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() && strings.Contains(path, "target") {
			return filepath.SkipDir
		}
		if !info.IsDir() && info.Name() == "openapi.yaml" {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			found = append(found, filepath.ToSlash(rel))
		}
		return nil
	})
	return found, err
}
