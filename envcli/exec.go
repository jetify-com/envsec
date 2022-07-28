package envcli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
)

func ExecCmd(cmdCfg *CmdConfig) *cobra.Command {
	command := &cobra.Command{
		Use:   "exec <command>",
		Short: "executes a command with Jetpack-stored environment variables",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			commandString := strings.Join(args, " ")
			commandToRun := exec.Command("/bin/sh", "-c", commandString)

			// Default environment to fetch values from is DEV
			envNames := []string{"DEV"}
			// If a specific environment was set by the user, then just use that one.
			if cmd.Flags().Changed(environmentFlagName) {
				envNames = []string{cmdCfg.EnvId.EnvName}
			}
			envId := envsec.EnvId{
				OrgId:     cmdCfg.EnvId.OrgId,
				ProjectId: cmdCfg.EnvId.ProjectId,
				EnvName:   envNames[0],
			}
			// Get list of stored env variables
			envVars, err := cmdCfg.Store.List(cmd.Context(), envId)
			if err != nil {
				return errors.WithStack(err)
			}
			// Attach stored env variables to the command environment
			commandToRun.Env = []string{}
			for _, envVar := range envVars {
				commandToRun.Env = append(commandToRun.Env, fmt.Sprintf("%s=%s", envVar.Name, envVar.Value))
			}
			commandToRun.Stdin = cmd.InOrStdin()
			commandToRun.Stdout = cmd.OutOrStdout()
			commandToRun.Stderr = cmd.ErrOrStderr()
			return commandToRun.Run()
		},
	}
	return command
}
