// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"go.jetify.com/envsec/internal/build"
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
		_, _ = fmt.Fprintf(w, "Version:     %v\n", build.Version)
		_, _ = fmt.Fprintf(w, "Build Env:   %v\n", build.BuildEnv())
		_, _ = fmt.Fprintf(w, "Platform:    %v\n", fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
		_, _ = fmt.Fprintf(w, "Commit:      %v\n", build.Commit)
		_, _ = fmt.Fprintf(w, "Commit Time: %v\n", build.CommitDate)
		_, _ = fmt.Fprintf(w, "Go Version:  %v\n", runtime.Version())
	} else {
		_, _ = fmt.Fprintf(w, "%v\n", build.Version)
	}
	return nil
}
