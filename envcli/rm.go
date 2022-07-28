package envcli

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
	"go.jetpack.io/envsec/tux"
)

func RemoveCmd(cmdCfg *CmdConfig) *cobra.Command {
	command := &cobra.Command{
		Use:   "rm <NAME1> [<NAME2>]...",
		Short: "Delete one or more environment variables",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, envNames []string) error {
			err := cmdCfg.Store.DeleteAll(cmd.Context(), cmdCfg.EnvId, envNames)
			if err == nil {
				err = tux.WriteHeader(cmd.OutOrStdout(),
					"[DONE] Deleted environment %s %v in environment: %s\n",
					tux.Plural(envNames, "variable", "variables"),
					strings.Join(tux.QuotedTerms(envNames), ", "),
					strings.ToLower(cmdCfg.EnvId.EnvName),
				)
			}
			if errors.Is(err, envsec.FaultyParamError) {
				err = tux.WriteHeader(cmd.OutOrStdout(),
					"[CANCELLED] Could not delete variable '%v' in environment: %s.\n"+
						"Please make sure all listed variables exist and you have proper permission to remove them.\n",
					strings.Split(err.Error(), ":")[0],
					strings.ToLower(cmdCfg.EnvId.EnvName),
				)
			}
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		},
	}

	return command
}
