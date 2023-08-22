package auth

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

func writeFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(os.WriteFile(path, data, 0644))
}

func parseFile(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(json.Unmarshal(data, value))
}
