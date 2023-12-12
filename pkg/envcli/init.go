package envcli

import (
	"os"

	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/internal/build"
	"go.jetpack.io/envsec/pkg/envsec"
	"go.jetpack.io/pkg/envvar"
)

type initCmdFlags struct {
	force bool
}

func initCmd() *cobra.Command {
	flags := &initCmdFlags{}
	command := &cobra.Command{
		Use:   "init",
		Short: "initialize directory and envsec project",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			workingDir, err := os.Getwd()
			if err != nil {
				return err
			}

			return (&envsec.Envsec{
				APIHost: build.JetpackAPIHost(),
				Auth: envsec.AuthConfig{
					ClientID: envvar.Get("ENVSEC_CLIENT_ID", build.ClientID()),
					Issuer:   envvar.Get("ENVSEC_ISSUER", build.Issuer()),
				},
				IsDev:      build.IsDev,
				Stderr:     cmd.ErrOrStderr(),
				WorkingDir: workingDir,
			}).NewProject(cmd.Context(), flags.force)
		},
	}

	command.Flags().BoolVarP(
		&flags.force,
		"force",
		"f",
		false,
		"Force initialization even if already initialized",
	)

	return command
}
