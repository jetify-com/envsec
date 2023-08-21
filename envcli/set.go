// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/tux"
)

func SetCmd(cmdCfg *CmdConfig) *cobra.Command {
	command := &cobra.Command{
		Use:   "set <NAME1>=<value1> [<NAME2>=<value2>]...",
		Short: "Securely store one or more environment variables",
		Long:  "Securely store one or more environment variables. To test contents of a file as a secret use set=@<file>",
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				k, _, ok := strings.Cut(arg, "=")
				if !ok || k == "" {
					return errors.Errorf(
						"argument %s must have an '=' to be of the form NAME=VALUE",
						arg,
					)
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			envMap, err := parseArgs(args)
			if err != nil {
				return errors.WithStack(err)
			}
			err = SetEnvMap(cmd.Context(), cmdCfg.Store, cmdCfg.EnvId, envMap)
			if err != nil {
				return errors.WithStack(err)
			}

			insertedNames := lo.Keys(envMap)
			err = tux.WriteHeader(cmd.OutOrStdout(),
				"[DONE] Set environment %s %v in environment: %s\n",
				tux.Plural(insertedNames, "variable", "variables"),
				strings.Join(tux.QuotedTerms(insertedNames), ", "),
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

func parseArgs(args []string) (map[string]string, error) {
	envMap := map[string]string{}
	for _, arg := range args {
		k, v, _ := strings.Cut(arg, "=")
		if strings.HasPrefix(v, "\\@") {
			v = strings.TrimPrefix(v, "\\")
		} else if strings.HasPrefix(v, "@") {
			file := strings.TrimPrefix(v, "@")
			if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
				return nil, errors.Errorf(
					"@ syntax is used for setting a secret from a file. file %s "+
						"does not exist. If your value starts with @, escape it with "+
						"a backslash, e.g. %s='\\%s'",
					file,
					k,
					v,
				)
			}
			c, err := os.ReadFile(file)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to read file %s", file)
			}
			v = string(c)
		}
		envMap[k] = v
	}
	return envMap, nil
}
