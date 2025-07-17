package tux

import (
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

func FTable(w io.Writer, rows [][]string) error {
	table := tablewriter.NewWriter(w)
	for _, row := range rows {
		if err := table.Append(row); err != nil {
			return errors.WithStack(err)
		}
	}
	if err := table.Render(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
