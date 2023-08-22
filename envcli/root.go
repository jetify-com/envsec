// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
)

type CmdConfig struct {
	Store envsec.Store
	EnvId envsec.EnvId
}

type globalFlags struct {
	projectId string
	orgId     string
	envName   string
}

func EnvCmd() *cobra.Command {
	flags := &globalFlags{}
	cmdConfig := CmdConfig{}

	command := &cobra.Command{
		Use:   "envsec",
		Short: "Manage environment variables and secrets",
		Long: heredoc.Doc(`
			Manage environment variables and secrets

			Securely stores and retrieves environment variables on the cloud.
			Environment variables are always encrypted, which makes it possible to
			store values that contain passwords and other secrets.
		`),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			c, err := newCmdConfig(cmd.Context(), flags)
			if err != nil {
				return errors.WithStack(err)
			}
			cmdConfig = *c
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	registerFlags(command, flags)

	command.AddCommand(DownloadCmd(&cmdConfig))
	command.AddCommand(ExecCmd(&cmdConfig))
	command.AddCommand(ListCmd(&cmdConfig))
	command.AddCommand(RemoveCmd(&cmdConfig))
	command.AddCommand(SetCmd(&cmdConfig))
	command.AddCommand(UploadCmd(&cmdConfig))
	command.AddCommand(authCmd())
	command.SetUsageFunc(UsageFunc)
	return command
}

func registerFlags(cmd *cobra.Command, opts *globalFlags) {
	cmd.PersistentFlags().StringVar(
		&opts.projectId,
		"project-id",
		"",
		"Project id to namespace secrets by",
	)

	cmd.PersistentFlags().StringVar(
		&opts.orgId,
		"org-id",
		"",
		"Organization id to namespace secrets by",
	)

	cmd.PersistentFlags().StringVar(
		&opts.envName,
		"environment",
		"dev",
		"Environment name, such as dev or prod",
	)
}

func Execute(ctx context.Context) {
	cmd := EnvCmd()
	_ = cmd.ExecuteContext(ctx)
}

func newCmdConfig(ctx context.Context, flags *globalFlags) (*CmdConfig, error) {
	s, err := envsec.NewStore(ctx, &envsec.SSMConfig{})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	envid, err := envsec.NewEnvId(flags.projectId, flags.orgId, flags.envName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &CmdConfig{
		Store: s,
		EnvId: envid,
	}, nil
}
