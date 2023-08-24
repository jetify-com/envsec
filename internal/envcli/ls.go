// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
)

const environmentFlagName = "environment"

type listCmdFlags struct {
	configFlags
	ShowValues bool
	Format     string
}

func listCmd() *cobra.Command {
	flags := &listCmdFlags{}

	command := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List all stored environment variables",
		Long:    "List all stored environment variables. If no environment flag is provided, variables in all environments will be listed.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdCfg, err := flags.genConfig(cmd.Context())
			if err != nil {
				return err
			}
			// Populate the valid Environments
			envNames := []string{"DEV", "PROD", "STAGING"}
			// If a specific environment was set by the user, then just use that one.
			if cmd.Flags().Changed(environmentFlagName) {
				envNames = []string{cmdCfg.EnvId.EnvName}
			}

			// TODO: parallelize
			for _, envName := range envNames {
				envId := envsec.EnvId{
					OrgId:     cmdCfg.EnvId.OrgId,
					ProjectId: cmdCfg.EnvId.ProjectId,
					EnvName:   envName,
				}
				envVars, err := cmdCfg.Store.List(cmd.Context(), envId)
				if err != nil {
					return errors.WithStack(err)
				}

				err = printEnv(cmd, envId, envVars, flags.ShowValues, flags.Format)
				if err != nil {
					return errors.WithStack(err)
				}
			}
			return nil
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
		"Display the key values in key=value format",
	)
	flags.configFlags.register(command)

	return command
}
