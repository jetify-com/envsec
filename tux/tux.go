package tux

import (
	"io"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

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
