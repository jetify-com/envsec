// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
	"go.jetpack.io/envsec/internal/tux"
)

type removeCmdFlags struct {
	configFlags
}

func removeCmd() *cobra.Command {
	flags := &removeCmdFlags{}
	command := &cobra.Command{
		Use:   "rm <NAME1> [<NAME2>]...",
		Short: "Delete one or more environment variables",
		Long:  "Delete one or more environment variables that are stored.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, envNames []string) error {
			cmdCfg, err := flags.genConfig(cmd.Context())
			if err != nil {
				return errors.WithStack(err)
			}
			err = cmdCfg.Store.DeleteAll(cmd.Context(), cmdCfg.EnvId, envNames)
			if err == nil {
				err = tux.WriteHeader(cmd.OutOrStdout(),
					"[DONE] Deleted environment %s %v in environment: %s\n",
					tux.Plural(envNames, "variable", "variables"),
					strings.Join(tux.QuotedTerms(envNames), ", "),
					strings.ToLower(cmdCfg.EnvId.EnvName),
				)
			}
			if errors.Is(err, envsec.FaultyParamError) {
				err = tux.WriteHeader(cmd.OutOrStdout(),
					"[CANCELLED] Could not delete variable '%v' in environment: %s.\n"+
						"Please make sure all listed variables exist and you have proper permission to remove them.\n",
					strings.Split(err.Error(), ":")[0],
					strings.ToLower(cmdCfg.EnvId.EnvName),
				)
			}
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		},
	}
	flags.configFlags.register(command)

	return command
}
