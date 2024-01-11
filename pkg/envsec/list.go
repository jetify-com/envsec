package envsec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"go.jetpack.io/envsec/internal/tux"
)

func (e *Envsec) List(ctx context.Context) ([]EnvVar, error) {
	return e.Store.List(ctx, e.EnvID)
}

func PrintEnvVar(
	w io.Writer,
	envID EnvID,
	envVars []EnvVar, // list of (name, value) pairs
	expose bool,
	format string,
) error {
	envVarsMaskedValue := []EnvVar{}
	// Masking envVar values if printValue flag isn't set
	for _, envVar := range envVars {
		valueToPrint := "*****"
		if expose {
			valueToPrint = envVar.Value
		}
		envVarsMaskedValue = append(envVarsMaskedValue, EnvVar{
			Name:  envVar.Name,
			Value: valueToPrint,
		})

	}

	switch format {
	case "table":
		return printTableFormat(w, envID, envVarsMaskedValue)
	case "dotenv":
		return printDotenvFormat(envVarsMaskedValue)
	case "json":
		return printJSONFormat(envVarsMaskedValue)
	default:
		return errors.New("incorrect format. Must be one of table|dotenv|json")
	}
}

func printTableFormat(w io.Writer, envID EnvID, envVars []EnvVar) error {
	err := tux.WriteHeader(w, "Environment: %s\n", strings.ToLower(envID.EnvName))
	if err != nil {
		return errors.WithStack(err)
	}
	table := tablewriter.NewWriter(w)
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

func printDotenvFormat(envVars []EnvVar) error {
	keyValsToPrint := ""
	for _, envVar := range envVars {
		keyValsToPrint += fmt.Sprintf("%s=%q\n", envVar.Name, envVar.Value)
	}

	// Add an empty line after the table is rendered.
	fmt.Println(keyValsToPrint)

	return nil
}

func printJSONFormat(envVars []EnvVar) error {
	data, err := json.MarshalIndent(envVars, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))

	return nil
}
