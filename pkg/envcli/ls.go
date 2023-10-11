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

			// TODO: parallelize
			for _, envName := range cmdCfg.EnvNames {
				envID := envsec.EnvID{
					OrgID:     cmdCfg.EnvID.OrgID,
					ProjectID: cmdCfg.EnvID.ProjectID,
					EnvName:   envName,
				}
				envVars, err := cmdCfg.Store.List(cmd.Context(), envID)
				if err != nil {
					return errors.WithStack(err)
				}

				err = printEnv(cmd, envID, envVars, flags.ShowValues, flags.Format)
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
