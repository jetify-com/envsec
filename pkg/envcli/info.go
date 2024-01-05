package envcli

import (
	"os"

	"github.com/spf13/cobra"
)

func infoCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "info",
		Short: "show info about the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			workingDir, err := os.Getwd()
			if err != nil {
				return err
			}
			return defaultEnvsec(cmd, workingDir).
				DescribeCurrentProject(cmd.Context(), cmd.OutOrStdout())
		},
	}

	return command
}
