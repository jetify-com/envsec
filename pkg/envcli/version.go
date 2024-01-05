// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/internal/build"
)

type versionFlags struct {
	verbose bool
}

func versionCmd() *cobra.Command {
	flags := versionFlags{}
	command := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return versionCmdFunc(cmd, args, flags)
		},
	}

	command.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, // value
		"displays additional version information",
	)
	return command
}

func versionCmdFunc(cmd *cobra.Command, _ []string, flags versionFlags) error {
	w := cmd.OutOrStdout()
	if flags.verbose {
		fmt.Fprintf(w, "Version:     %v\n", build.Version)
		fmt.Fprintf(w, "Build Env:   %v\n", build.BuildEnv())
		fmt.Fprintf(w, "Platform:    %v\n", fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
		fmt.Fprintf(w, "Commit:      %v\n", build.Commit)
		fmt.Fprintf(w, "Commit Time: %v\n", build.CommitDate)
		fmt.Fprintf(w, "Go Version:  %v\n", runtime.Version())
	} else {
		fmt.Fprintf(w, "%v\n", build.Version)
	}
	return nil
}
