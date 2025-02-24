// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetify.com/envsec/pkg/envsec"
)

type setCmdFlags struct {
	configFlags
}

func SetCmd() *cobra.Command {
	flags := &setCmdFlags{}
	command := &cobra.Command{
		Use:   "set <NAME1>=<value1> [<NAME2>=<value2>]...",
		Short: "Securely store one or more environment variables",
		Long:  "Securely store one or more environment variables. To test contents of a file as a secret use set=@<file>",
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return envsec.ValidateSetArgs(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cmdCfg, err := flags.genConfig(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			return cmdCfg.envsec.SetFromArgs(ctx, args)
		},
	}
	flags.configFlags.register(command)
	return command
}
