// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
)

type execCmdFlags struct {
	configFlags
}

func execCmd() *cobra.Command {
	flags := &execCmdFlags{}
	command := &cobra.Command{
		Use:   "exec <command>",
		Short: "Execute a command with Jetpack-stored environment variables",
		Long:  "Execute a specified command with remote environment variables being present for the duration of the command. If an environment variable exists both locally and in remote storage, the remotely stored one is prioritized.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCfg, err := flags.genConfig(cmd.Context())
			if err != nil {
				return err
			}
			commandString := strings.Join(args, " ")
			commandToRun := exec.Command("/bin/sh", "-c", commandString)

			// Default environment to fetch values from is DEV
			envNames := []string{"dev"}
			// If a specific environment was set by the user, then just use that one.
			if cmd.Flags().Changed(environmentFlagName) {
				envNames = []string{strings.ToLower(cmdCfg.EnvID.EnvName)}
			}
			envID := envsec.EnvID{
				OrgID:     cmdCfg.EnvID.OrgID,
				ProjectID: cmdCfg.EnvID.ProjectID,
				EnvName:   envNames[0],
			}
			// Get list of stored env variables
			envVars, err := cmdCfg.Store.List(cmd.Context(), envID)
			if err != nil {
				return errors.WithStack(err)
			}
			// Attach stored env variables to the command environment
			commandToRun.Env = os.Environ()
			for _, envVar := range envVars {
				commandToRun.Env = append(commandToRun.Env, fmt.Sprintf("%s=%s", envVar.Name, envVar.Value))
			}
			commandToRun.Stdin = cmd.InOrStdin()
			commandToRun.Stdout = cmd.OutOrStdout()
			commandToRun.Stderr = cmd.ErrOrStderr()
			return commandToRun.Run()
		},
	}
	flags.configFlags.register(command)
	return command
}
