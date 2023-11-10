package envcli

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/pkg/jetcloud"
)

func initCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "init",
		Short: "initialize directory and envsec project",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAuthClient()
			if err != nil {
				return err
			}
			tok, err := client.GetSession(cmd.Context())
			if err != nil {
				return fmt.Errorf("error: %w, run `envsec auth login`", err)
			}

			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			projectID, err := jetcloud.InitProject(cmd.Context(), tok, wd)
			if errors.Is(err, jetcloud.ErrProjectAlreadyInitialized) {
				fmt.Fprintf(
					cmd.ErrOrStderr(),
					"Warning: project already initialized ID=%s\n",
					projectID,
				)
			} else if err != nil {
				return err
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "Initialized project ID=%s\n", projectID)
			}
			return nil
		},
	}
	return command
}
