package envcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/tux"
)

func DownloadCmd(cmdCfg *CmdConfig) *cobra.Command {
	command := &cobra.Command{
		Use:   "download <file1>",
		Short: "Download environment variables into the specified .env file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			orderedEnvVars, err := listEnv(cmd, cmdCfg.Store, cmdCfg.EnvId)
			if err != nil {
				return errors.WithStack(err)
			}

			if len(orderedEnvVars) == 0 {
				err = tux.WriteHeader(cmd.OutOrStdout(),
					"[DONE] There are no environment variables to download for environment: %s\n",
					strings.ToLower(cmdCfg.EnvId.EnvName),
				)
				return errors.WithStack(err)
			}

			wd, err := os.Getwd()
			if err != nil {
				return errors.WithStack(err)
			}
			// A single relativeFilePath is guaranteed to be there.
			filePath := filepath.Join(wd, args[0] /* relativeFilePath */)

			// .env file contents
			lines := []string{}
			for _, envVar := range orderedEnvVars {
				// name=value
				lines = append(lines, fmt.Sprintf("%s=%s", envVar[0], envVar[1]))
			}
			contents := strings.Join(lines, "\n")

			err = os.WriteFile(filePath, []byte(contents), 0644)
			if err != nil {
				return errors.WithStack(err)
			}
			err = tux.WriteHeader(cmd.OutOrStdout(),
				"[DONE] Downloaded environment variables to %v for environment: %s\n",
				strings.Join(tux.QuotedTerms(args), ", "),
				strings.ToLower(cmdCfg.EnvId.EnvName),
			)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		},
	}
	return command
}
