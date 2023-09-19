// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/internal/tux"
)

var errUnsupportedFormat = errors.New("unsupported format")

type uploadCmdFlags struct {
	configFlags
	format string
}

func uploadCmd() *cobra.Command {
	flags := &uploadCmdFlags{}
	command := &cobra.Command{
		Use:   "upload <file1> [<fileN>]...",
		Short: "Upload variables defined in a .env file",
		Long: "Upload variables defined in one or more .env files. The files " +
			"should have one NAME=VALUE per line.",
		Args: cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.format == "json" || flags.format == "env" {
				return nil
			}
			return errors.Wrapf(errUnsupportedFormat, "format: %s", flags.format)
		},
		RunE: func(cmd *cobra.Command, relativeFilePaths []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return errors.WithStack(err)
			}

			filePaths := []string{}
			for _, relFilePath := range relativeFilePaths {
				// get an absolute path from the relative path
				absPath := filepath.Join(wd, relFilePath)

				exists, err := fileExists(absPath)
				if err != nil {
					return errors.WithStack(err)
				}
				if !exists {
					return errors.Errorf("could not find file at path: %s", relFilePath)
				}
				filePaths = append(filePaths, absPath)
			}

			var envMap map[string]string
			if flags.format == "json" {
				envMap, err = loadFromJSON(filePaths)
				if err != nil {
					return errors.Wrap(
						err,
						"failed to load from JSON. Ensure the file is a flat key-value "+
							"JSON formatted file",
					)
				}
			} else {
				envMap, err = godotenv.Read(filePaths...)
				if err != nil {
					return errors.WithStack(err)
				}
			}

			cmdCfg, err := flags.genConfig(cmd.Context())
			if err != nil {
				return err
			}
			err = SetEnvMap(cmd.Context(), cmdCfg.Store, cmdCfg.EnvID, envMap)
			if err != nil {
				return errors.WithStack(err)
			}

			err = tux.WriteHeader(cmd.OutOrStdout(),
				"[DONE] Uploaded %d environment variable(s) from %s %v to environment: %s\n",
				len(envMap),
				tux.Plural(relativeFilePaths, "file", "files"),
				strings.Join(tux.QuotedTerms(relativeFilePaths), ", "),
				strings.ToLower(cmdCfg.EnvID.EnvName),
			)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		},
	}

	command.Flags().StringVarP(
		&flags.format, "format", "f", "env", "File format: env or json")
	flags.configFlags.register(command)

	return command
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
