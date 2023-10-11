// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
	"go.jetpack.io/envsec/internal/tux"
)

const nameRegexStr = "^[a-zA-Z_][a-zA-Z0-9_]*"

var nameRegex = regexp.MustCompile(nameRegexStr)

func SetEnvMap(ctx context.Context, store envsec.Store, envID envsec.EnvID, envMap map[string]string) error {
	err := ensureValidNames(lo.Keys(envMap))
	if err != nil {
		return errors.WithStack(err)
	}

	err = store.SetAll(ctx, envID, envMap)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func ensureValidNames(names []string) error {
	for _, name := range names {

		// Any variation of jetpack_ or JETPACK_ prefix is not allowed
		lowerName := strings.ToLower(name)
		if strings.HasPrefix(lowerName, "jetpack_") {
			return errors.Errorf(
				"name %s cannot start with JETPACK_ (or lowercase)",
				name,
			)
		}

		if !nameRegex.MatchString(name) {
			return errors.Errorf(
				"name %s must match the regular expression: %s ",
				name,
				nameRegexStr,
			)
		}
	}
	return nil
}

func printEnv(
	cmd *cobra.Command,
	envID envsec.EnvID,
	envVars []envsec.EnvVar, // list of (name, value) pairs
	flagPrintValues bool,
	flagFormat string,
) error {
	envVarsMaskedValue := []envsec.EnvVar{}
	// Masking envVar values if printValue flag isn't set
	for _, envVar := range envVars {
		valueToPrint := "*****"
		if flagPrintValues {
			valueToPrint = envVar.Value
		}
		envVarsMaskedValue = append(envVarsMaskedValue, envsec.EnvVar{
			Name:  envVar.Name,
			Value: valueToPrint,
		})

	}

	switch flagFormat {
	case "table":
		return printTableFormat(cmd, envID, envVarsMaskedValue)
	case "dotenv":
		return printDotenvFormat(envVarsMaskedValue)
	case "json":
		return printJSONFormat(envVarsMaskedValue)
	default:
		return errors.New("incorrect format. Must be one of table|dotenv|json")
	}

}

func printTableFormat(cmd *cobra.Command,
	envID envsec.EnvID,
	envVars []envsec.EnvVar, // list of (name, value) pairs
) error {
	err := tux.WriteHeader(cmd.OutOrStdout(), "Environment: %s\n", strings.ToLower(envID.EnvName))
	if err != nil {
		return errors.WithStack(err)
	}
	table := tablewriter.NewWriter(cmd.OutOrStdout())
	table.SetHeader([]string{"Name", "Value"})
	tableValues := [][]string{}
	for _, envVar := range envVars {
		tableValues = append(tableValues, []string{envVar.Name /*name*/, envVar.Value})
	}
	table.AppendBulk(tableValues)

	if len(tableValues) == 0 {
		fmt.Println("No environment variables currently defined.")
	} else {
		table.Render()
	}

	// Add an empty line after the table is rendered.
	fmt.Println()

	return nil
}

func printDotenvFormat(envVars []envsec.EnvVar) error {
	keyValsToPrint := ""
	for _, envVar := range envVars {
		keyValsToPrint += fmt.Sprintf("%s=%q\n", envVar.Name, envVar.Value)
	}

	// Add an empty line after the table is rendered.
	fmt.Println(keyValsToPrint)

	return nil
}

func printJSONFormat(envVars []envsec.EnvVar) error {
	data, err := json.MarshalIndent(envVars, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))

	return nil
}

func fileExists(path string) (bool, error) {
	fileinfo, err := os.Stat(path)
	if err == nil {
		if !fileinfo.IsDir() {
			// It is a file!
			return true, nil
		}
		// It is a directory
		return false, nil
	}

	// No such file was found:
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	// Some other error:
	return false, errors.WithStack(err)
}
