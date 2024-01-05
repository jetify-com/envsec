package tux

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

func FTable(w io.Writer, rows [][]string) {
	table := tablewriter.NewWriter(w)
	for _, row := range rows {
		table.Append(row)
	}
	table.Render()
}
