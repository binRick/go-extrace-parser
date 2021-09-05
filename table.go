package main

import (
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func tbl() {
	tbl_str := &strings.Builder{}
	tbl := tablewriter.NewWriter(tbl_str)
	tbl.SetRowLine(true)

	tbl.SetAutoFormatHeaders(false)
	tbl.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	tbl.SetHeader([]string{
		"Exec",
		"Qty",
	})
	tbl.SetAutoMergeCells(false)
	tbl.SetBorder(true)
	tbl.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: true})
	tbl.SetColumnColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiYellowColor},
	)
	tbl.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgHiYellowColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor, tablewriter.Bold, tablewriter.BgBlackColor},
	)
	for _, i := range exec_qtys[1:5] {
		tbl.Append([]string{
			fmt.Sprintf("%s",
				i.Exec,
			),
			fmt.Sprintf("%d",
				i.Qty,
			),
		})
	}
	tbl.SetFooter([]string{"", ""})
	tbl.SetCaption(true, fmt.Sprintf("Processed"))
	tbl.Render()
	fmt.Printf("%s\n", tbl_str.String())
}
