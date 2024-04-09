// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type rootCmdFlags struct {
	jsonErrors bool
}

func RootCmd(flags *rootCmdFlags) *cobra.Command {
	command := &cobra.Command{
		Use:   "envsec",
		Short: "Manage environment variables and secrets",
		Long: heredoc.Doc(`
			Manage environment variables and secrets

			Securely stores and retrieves environment variables on the cloud.
			Environment variables are always encrypted, which makes it possible to
			store values that contain passwords and other secrets.
		`),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.jsonErrors {
				// Don't print anything to stderr so we can print the error in json
				cmd.SetErr(io.Discard)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		// we're manually showing usage
		SilenceUsage: true,
		// We manually capture errors so we can print different formats
		SilenceErrors: true,
	}

	command.PersistentFlags().BoolVar(
		&flags.jsonErrors,
		"json-errors", false, "Print errors in json format",
	)
	command.Flag("json-errors").Hidden = true

	command.AddCommand(authCmd())
	command.AddCommand(DownloadCmd())
	command.AddCommand(ExecCmd())
	command.AddCommand(genDocsCmd())
	command.AddCommand(initCmd())
	command.AddCommand(ListCmd())
	command.AddCommand(infoCmd())
	command.AddCommand(RemoveCmd())
	command.AddCommand(SetCmd())
	command.AddCommand(UploadCmd())
	command.AddCommand(versionCmd())
	command.SetUsageFunc(UsageFunc)
	return command
}

func Execute(ctx context.Context) int {
	flags := &rootCmdFlags{}
	cmd := RootCmd(flags)
	err := cmd.ExecuteContext(ctx)
	if err == nil {
		return 0
	}
	if flags.jsonErrors {
		var jsonErr struct {
			Error string `json:"error"`
		}
		jsonErr.Error = err.Error()
		b, err := json.Marshal(jsonErr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			fmt.Fprintln(os.Stderr, string(b))
		}
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
	return 1
}
