package envcli

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/internal/jetcloud"
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
			tok, ok := client.GetSession()
			if !ok {
				return errors.New("not logged in, run `envsec auth login`")
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
