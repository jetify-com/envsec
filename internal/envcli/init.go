package envcli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/internal/jetcloud"
)

func initCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "init",
		Short: "initialize directory and envsec project",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := newAuthenticator().GetUser()
			if err != nil {
				return err
			}

			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			projectID, err := jetcloud.InitProject(cmd.Context(), user, wd)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Initialized project ID=%s\n", projectID)
			return nil
		},
	}
	return command
}
