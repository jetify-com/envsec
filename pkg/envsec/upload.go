package envsec

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"go.jetpack.io/pkg/runx/impl/fileutil"
)

// Upload uploads the environment variables for the environment specified from
// the given paths.
// If format is empty, we default to dotenv format unless path ends in .json
func (e *Envsec) Upload(ctx context.Context, paths []string, format string) error {
	if err := ValidateFormat(format); err != nil {
		return err
	}

	filePaths := []string{}
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(e.WorkingDir, path)
		}

		if !fileutil.Exists(path) {
			return errors.Errorf("could not find file at path: %s", path)
		}
		filePaths = append(filePaths, path)
	}

	envMap := map[string]string{}
	var err error
	for _, path := range filePaths {
		var newVars map[string]string
		if format == "json" || (format == "" && filepath.Ext(path) == ".json") {
			newVars, err = loadFromJSON([]string{path})
			if err != nil {
				return errors.Wrap(
					err,
					"failed to load from JSON. Ensure the file is a flat key-value "+
						"JSON formatted file",
				)
			}
		} else {
			newVars, err = godotenv.Read(path)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		for k, v := range newVars {
			envMap[k] = v
		}
	}

	return e.SetMap(ctx, envMap)
}

func loadFromJSON(filePaths []string) (map[string]string, error) {
	envMap := map[string]string{}
	for _, filePath := range filePaths {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if err = json.Unmarshal(content, &envMap); err != nil {
			return nil, errors.WithStack(err)
		}
		for k, v := range envMap {
			envMap[k] = v
		}
	}
	return envMap, nil
}

func ValidateFormat(format string) error {
	if format != "" && format != "json" && format != "dotenv" {
		return errors.Errorf("incorrect format. Must be one of json|dotenv")
	}
	return nil
}
