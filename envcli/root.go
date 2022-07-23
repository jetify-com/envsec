package envcli

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
)

type CmdConfig struct {
	Store envsec.Store
	EnvId envsec.EnvId
}

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
	}
	cmdConfig := &CmdConfig{}
	command.AddCommand(ListCmd(cmdConfig))
	command.SetUsageFunc(UsageFunc)
	return command
}

func Execute(ctx context.Context) {
	cmd := EnvCmd()
	_ = cmd.ExecuteContext(ctx)
}
