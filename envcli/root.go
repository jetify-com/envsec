// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func EnvCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "envsec",
		Short: "Manage environment variables and secrets",
		Long: heredoc.Doc(`
			Manage environment variables and secrets

			Securely stores and retrieves environment variables on the cloud.
			Environment variables are always encrypted, which makes it possible to
			store values that contain passwords and other secrets.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		// we're manually showing usage
		SilenceUsage: true,
	}

	command.AddCommand(downloadCmd())
	command.AddCommand(execCmd())
	command.AddCommand(listCmd())
	command.AddCommand(removeCmd())
	command.AddCommand(setCmd())
	command.AddCommand(uploadCmd())
	command.AddCommand(authCmd())
	command.SetUsageFunc(UsageFunc)
	return command
}

func Execute(ctx context.Context) {
	cmd := EnvCmd()
	_ = cmd.ExecuteContext(ctx)
}
