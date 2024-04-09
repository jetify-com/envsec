// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/pkg/envsec"
)

type downloadCmdFlags struct {
	configFlags
	format string
}

func DownloadCmd() *cobra.Command {
	flags := &downloadCmdFlags{}
	command := &cobra.Command{
		Use:   "download <file1>",
		Short: "Download environment variables into the specified file",
		Long:  "Download environment variables stored into the specified file (most commonly a .env file). The format of the file is one NAME=VALUE per line.",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return envsec.ValidateFormat(flags.format)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCfg, err := flags.genConfig(cmd)
			if err != nil {
				return errors.WithStack(err)
			}
			return cmdCfg.envsec.Download(cmd.Context(), args[0], flags.format)
		},
	}

	flags.configFlags.register(command)
	command.Flags().StringVarP(
		&flags.format, "format", "f", "env", "file format: dotenv or json")

	return command
}
