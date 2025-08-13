// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"github.com/spf13/cobra"
	"go.jetify.com/envsec/pkg/envsec"
)

const environmentFlagName = "environment"

type listCmdFlags struct {
	configFlags
	ShowValues bool
	Format     string
}

func ListCmd() *cobra.Command {
	flags := &listCmdFlags{}

	command := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List all stored environment variables",
		Long:    "List all stored environment variables. If no environment flag is provided, variables in all environments will be listed.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdCfg, err := flags.genConfig(cmd)
			if err != nil {
				return err
			}

			secrets, err := cmdCfg.envsec.List(cmd.Context())
			if err != nil {
				return err
			}

			return envsec.PrintEnvVar(
				cmd.OutOrStdout(), cmdCfg.envsec.EnvID, secrets, flags.ShowValues, flags.Format)
		},
	}

	command.Flags().BoolVarP(
		&flags.ShowValues,
		"show",
		"s",
		false,
		"display the value of each environment variable (secrets included)",
	)
	command.Flags().StringVarP(
		&flags.Format,
		"format",
		"f",
		"table",
		"format to use for displaying keys and values, one of: table, dotenv, json",
	)
	flags.register(command)

	return command
}
