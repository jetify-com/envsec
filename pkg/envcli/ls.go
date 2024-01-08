// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"os"

	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/pkg/envsec"
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

			envIDs := []envsec.EnvID{}
			for _, envName := range cmdCfg.EnvNames {
				envIDs = append(envIDs, envsec.EnvID{
					OrgID:     cmdCfg.EnvID.OrgID,
					ProjectID: cmdCfg.EnvID.ProjectID,
					EnvName:   envName,
				})
			}

			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			vars, err := defaultEnvsec(cmd, wd).List(
				cmd.Context(),
				cmdCfg.Store,
				envIDs...,
			)
			if err != nil {
				return err
			}

			return envsec.PrintEnvVars(
				vars, cmd.OutOrStdout(), flags.ShowValues, flags.Format)
		},
	}

	command.Flags().BoolVarP(
		&flags.ShowValues,
		"show",
		"s",
		false,
		"Display the value of each environment variable (secrets included)",
	)
	command.Flags().StringVarP(
		&flags.Format,
		"format",
		"f",
		"table",
		"Display the key values in key=value format. Must be one of: table | dotenv | json",
	)
	flags.configFlags.register(command)

	return command
}
