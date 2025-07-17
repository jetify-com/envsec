// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"go.jetify.com/envsec/internal/tux"
)

// TODO: move to file
var usageTmpl = heredoc.Doc(`
{{ "Usage:" | style "h2" }}
	{{if .Runnable}}{{.UseLine | style "command" }}{{end}}
	{{- if .HasAvailableSubCommands}} {{"<command>" | style "subcommand"}}{{end}}


{{- if gt (len .Aliases) 0}}

{{ "Aliases:" | style "h2" }}
	{{.NameAndAliases}}
{{- end}}


{{- if .HasExample}}

{{ "Examples:" | style "h2" }}
	{{.Example}}
{{- end}}


{{- if .HasAvailableSubCommands}}

{{ "Available Commands:" | style "h2" }}
	{{- range .Commands}}
		{{- if (or .IsAvailableCommand (eq .Name "help"))}}
	{{rpad .Name .NamePadding | style "subcommand"}} {{.Short}}
		{{- end}}
	{{- end}}
{{- end}}


{{- if .HasAvailableLocalFlags}}

{{ "Flags:" | style "h2" }}
{{ .LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end}}


{{- if .HasAvailableInheritedFlags}}

{{ "Global Flags:" | style "h2" }}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end}}


{{- if .HasHelpSubCommands}}

{{ "Additional help topics:" | style "h2" }}
	{{- range .Commands}}
		{{- if .IsAdditionalHelpTopicCommand}}
			{{rpad .CommandPath .CommandPathPadding}} {{.Short}}
		{{- end}}
	{{- end}}
{{- end}}


{{- if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
{{- end}}
`)

var baseStyle = tux.StyleSheet{
	Styles: map[string]tux.StyleRule{
		"h1": {
			Bold:       true,
			Foreground: "$purple",
		},
		"h2": {
			Bold: true,
			// Foreground: "$purple",
		},
		"command": {
			Foreground: "$cyan",
		},
		"subcommand": {
			Foreground: "$magenta",
		},
		"flag": {
			Bold:       true,
			Foreground: "$purple",
		},
	},
	Tokens: map[string]string{
		"$purple":  "#bd93f9",
		"$yellow":  "#ffb86c",
		"$cyan":    "51",
		"$magenta": "#ff79c6",
		"$green":   "#50fa7b",
	},
}

func UsageFunc(cmd *cobra.Command) error {
	t := tux.New()
	t.SetIn(cmd.InOrStdin())
	t.SetOut(cmd.OutOrStdout())
	t.SetErr(cmd.ErrOrStderr())
	t.SetStyleSheet(baseStyle)
	t.PrintT(usageTmpl, cmd)
	return nil
}
