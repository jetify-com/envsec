// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/tux"
)

func UploadCmd(cmdCfg *CmdConfig) *cobra.Command {

	command := &cobra.Command{
		Use:   "upload <file1> [<fileN>]...",
		Short: "Upload variables defined in a .env file",
		Long:  "Upload variables defined in one or more .env files. The files should have one NAME=VALUE per line.",
		Args:  cobra.MinimumNArgs(1),
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

			envMap, err := godotenv.Read(filePaths...)
			if err != nil {
				return errors.WithStack(err)
			}

			err = SetEnvMap(cmd.Context(), cmdCfg.Store, cmdCfg.EnvId, envMap)
			if err != nil {
				return errors.WithStack(err)
			}

			err = tux.WriteHeader(cmd.OutOrStdout(),
				"[DONE] Uploaded environment variables from %s %v to environment: %s\n",
				tux.Plural(relativeFilePaths, "file", "files"),
				strings.Join(tux.QuotedTerms(relativeFilePaths), ", "),
				strings.ToLower(cmdCfg.EnvId.EnvName),
			)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		},
	}
	return command
}
