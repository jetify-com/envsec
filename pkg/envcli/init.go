package envcli

import (
	"os"

	"github.com/spf13/cobra"
	"go.jetify.com/envsec/internal/build"
	"go.jetify.com/envsec/pkg/envsec"
	"go.jetify.com/pkg/envvar"
)

type initCmdFlags struct {
	force bool
}

func initCmd() *cobra.Command {
	flags := &initCmdFlags{}
	command := &cobra.Command{
		Use:   "init",
		Short: "Initialize directory and envsec project",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			workingDir, err := os.Getwd()
			if err != nil {
				return err
			}

			e := defaultEnvsec(cmd, workingDir)
			return e.NewProject(cmd.Context(), flags.force)
		},
	}

	command.Flags().BoolVarP(
		&flags.force,
		"force",
		"f",
		false,
		"force initialization even if already initialized",
	)

	return command
}

func defaultEnvsec(cmd *cobra.Command, workingDir string) *envsec.Envsec {
	return &envsec.Envsec{
		APIHost: build.JetpackAPIHost(),
		Auth: envsec.AuthConfig{
			Audience:        []string{envvar.Get("ENVSEC_AUDIENCE", build.Audience())},
			ClientID:        envvar.Get("ENVSEC_CLIENT_ID", build.ClientID()),
			Issuer:          envvar.Get("ENVSEC_ISSUER", build.Issuer()),
			SuccessRedirect: envvar.Get("ENVSEC_SUCCESS_REDIRECT", build.SuccessRedirect()),
		},
		IsDev:      build.IsDev,
		Stderr:     cmd.ErrOrStderr(),
		WorkingDir: workingDir,
	}
}
