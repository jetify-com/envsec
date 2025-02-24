package envsec

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"go.jetify.com/envsec/internal/tux"
)

// Download downloads the environment variables for the environment specified.
// If format is empty, we default to dotenv format unless path ends in .json
func (e *Envsec) Download(ctx context.Context, path, format string) error {
	if err := ValidateFormat(format); err != nil {
		return err
	}

	envVars, err := e.List(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(envVars) == 0 {
		err = tux.WriteHeader(e.Stderr,
			"[DONE] There are no environment variables to download for environment: %s\n",
			strings.ToLower(e.EnvID.EnvName),
		)
		return errors.WithStack(err)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(e.WorkingDir, path)
	}

	envVarMap := map[string]string{}
	for _, envVar := range envVars {
		envVarMap[envVar.Name] = envVar.Value
	}

	if format == "" && filepath.Ext(path) == ".json" {
		format = "json"
	}

	var contents []byte
	if format == "json" {
		contents, err = encodeToJSON(envVarMap)
	} else {
		contents, err = encodeToDotEnv(envVarMap)
	}

	if err != nil {
		return errors.WithStack(err)
	}

	err = os.WriteFile(path, contents, 0o644)
	if err != nil {
		return errors.WithStack(err)
	}
	err = tux.WriteHeader(e.Stderr,
		"[DONE] Downloaded environment variables to %q for environment: %s\n",
		path,
		strings.ToLower(e.EnvID.EnvName),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func encodeToJSON(m map[string]string) ([]byte, error) {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m); err != nil {
		return nil, errors.WithStack(err)
	}
	return b.Bytes(), nil
}

func encodeToDotEnv(m map[string]string) ([]byte, error) {
	envContents, err := godotenv.Marshal(m)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return []byte(envContents), nil
}
