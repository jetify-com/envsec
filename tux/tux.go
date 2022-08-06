// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package tux

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/muesli/termenv"
	"github.com/pkg/errors"
)

type Tux struct {
	inReader   io.Reader
	outWriter  io.Writer
	errWriter  io.Writer
	styleSheet StyleSheet
}

func New() *Tux {
	// For now hardcoding the profile (because it's not working otherwise)
	// but need to change this to auto-detect appropriately.
	lipgloss.SetColorProfile(termenv.ANSI256)
	return &Tux{
		inReader:  os.Stdin,
		outWriter: os.Stdout,
		errWriter: os.Stderr,
	}
}

func (tux *Tux) SetOut(w io.Writer) {
	tux.outWriter = w
}

func (tux *Tux) SetErr(w io.Writer) {
	tux.errWriter = w
}

func (tux *Tux) SetIn(r io.Reader) {
	tux.inReader = r
}

func (tux *Tux) SetStyleSheet(styleSheet StyleSheet) {
	tux.styleSheet = styleSheet
}

func trimRight(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func (tux *Tux) PrintT(text string, data any) {
	// TODO: Initialize once when creating the tux object?
	templateFuncs := template.FuncMap{
		"trimTrailingWhitespaces": trimRight,
		"rpad":                    rpad,
		"style":                   StyleFunc(tux.styleSheet),
	}
	tpl := template.Must(template.New("tpl").Funcs(templateFuncs).Parse(text))
	err := tpl.Execute(tux.outWriter, data)
	if err != nil {
		tux.MustPrintErr(err)
		return
	}
}

func (tux *Tux) MustPrintErr(a ...any) {
	_, err := fmt.Fprint(tux.errWriter, a...)
	if err != nil {
		panic(err)
	}
}

// TODO: Migrate to style sheets
func WriteHeader(w io.Writer, format string, a ...any) error {
	headerPrintfFunc := color.New(color.FgHiCyan, color.Bold).SprintfFunc()
	message := headerPrintfFunc(format, a...)
	_, err := io.WriteString(w, message)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// QuotedTerms will wrap each term in single-quotation marks
func QuotedTerms(terms []string) []string {
	q := []string{}
	for _, term := range terms {
		// wrap the term in single-quote
		q = append(q, "'"+term+"'")
	}
	return q
}

func Plural[T any](items []T, singular string, plural string) string {
	if len(items) == 1 {
		return singular
	}
	return plural
}
