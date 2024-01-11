// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/pkg/envsec"
)

type uploadCmdFlags struct {
	configFlags
	format string
}

func UploadCmd() *cobra.Command {
	flags := &uploadCmdFlags{}
	command := &cobra.Command{
		Use:   "upload <file1> [<fileN>]...",
		Short: "Upload variables defined in a .env file",
		Long: "Upload variables defined in one or more .env files. The files " +
			"should have one NAME=VALUE per line.",
		Args: cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return envsec.ValidateFormat(flags.format)
		},
		RunE: func(cmd *cobra.Command, paths []string) error {
			cmdCfg, err := flags.genConfig(cmd)
			if err != nil {
				return err
			}

			return cmdCfg.envsec.Upload(cmd.Context(), paths, flags.format)
		},
	}

	command.Flags().StringVarP(
		&flags.format, "format", "f", "", "File format: dotenv or json")
	flags.configFlags.register(command)

	return command
}
