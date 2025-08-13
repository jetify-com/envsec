// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type removeCmdFlags struct {
	configFlags
}

func RemoveCmd() *cobra.Command {
	flags := &removeCmdFlags{}
	command := &cobra.Command{
		Use:   "rm <NAME1> [<NAME2>]...",
		Short: "Delete one or more environment variables",
		Long:  "Delete one or more environment variables that are stored.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, envNames []string) error {
			cmdCfg, err := flags.genConfig(cmd)
			if err != nil {
				return errors.WithStack(err)
			}
			return cmdCfg.envsec.DeleteAll(cmd.Context(), envNames...)
		},
	}
	flags.register(command)

	return command
}
