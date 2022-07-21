package envcli

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
	"go.jetpack.io/envsec/tux"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

const nameRegexStr = "^[a-zA-Z_][a-zA-Z0-9_]*"

var nameRegex = regexp.MustCompile(nameRegexStr)

func SetEnvMap(ctx context.Context, store envsec.Store, envId envsec.EnvId, envMap map[string]string) error {
	err := ensureValidNames(lo.Keys(envMap))
	if err != nil {
		return errors.WithStack(err)
	}

	err = store.SetAll(ctx, envId, envMap)
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

// listEnv returns an ordered list of (name, value) pairs
func listEnv(cmd *cobra.Command, store envsec.Store, envId envsec.EnvId) ([][]string, error) {

	envVars, err := store.List(cmd.Context(), envId)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	orderedNames := lo.Keys(envVars)

	c := collate.New(language.English, collate.Loose, collate.Numeric)
	c.SortStrings(orderedNames)

	orderedEnvVars := [][]string{}
	for _, name := range orderedNames {
		orderedEnvVars = append(orderedEnvVars, []string{name, envVars[name]})
	}

	return orderedEnvVars, nil
}

func printEnv(
	cmd *cobra.Command,
	envId envsec.EnvId,
	orderedEnvVars [][]string, // list of (name, value) pairs
	flagPrintValues bool,
) error {

	err := tux.WriteHeader(cmd.OutOrStdout(), "Environment: %s\n", strings.ToLower(envId.EnvName))
	if err != nil {
		return errors.WithStack(err)
	}
	table := tablewriter.NewWriter(cmd.OutOrStdout())
	table.SetHeader([]string{"Name", "Value"})
	tableValues := [][]string{}
	for _, envVar := range orderedEnvVars {
		valueToPrint := "*****"
		if flagPrintValues {
			valueToPrint = envVar[1]
		}

		tableValues = append(tableValues, []string{envVar[0] /*name*/, valueToPrint})
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
