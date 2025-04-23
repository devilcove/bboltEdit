package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type key struct {
	name string
	help string
}

var treeMoveKeys = []key{
	{"j,↓,→ ", "move selection down by one node"},
	{"k,↑,←", "move selection up by one node"},
	{"g, home", "move selection to top node"},
	{"G, end", "move selection to bottom node"},
	{"Ctrl-F, page down", "move selection down by one page"},
	{"Ctrl-B, page up", "move selection up by one page"},
	{"Esc", "Close window/dialog"},
}

var mainKeys = []key{
	{"F1", "show about/help"},
	{"Esc", "close dialog/application"},
	{"Ctrl-Q", "close application"},
	{"", ""},
	{"?", "detailed help for a window"},
}

func helpDialog(title string, width, height int, right, left []key) tview.Primitive { //nolint:ireturn
	table := tview.NewTable()
	for i, key := range left {
		table.SetCell(i, 0, tview.NewTableCell(key.name).
			SetAlign(tview.AlignCenter).SetExpansion(1).SetTextColor(tcell.ColorGrey))
		table.SetCell(i, 1, tview.NewTableCell(key.help).
			SetAlign(tview.AlignLeft).SetExpansion(1).SetTextColor(tcell.ColorBlue))
	}
	for i, key := range right {
		table.SetCell(i, 2, tview.NewTableCell(key.name).
			SetAlign(tview.AlignCenter).SetExpansion(1).SetTextColor(tcell.ColorGrey))
		table.SetCell(i, 3, tview.NewTableCell(key.help).
			SetAlign(tview.AlignLeft).SetExpansion(1).SetTextColor(tcell.ColorBlue))
	}
	grid := tview.NewGrid().
		SetRows(1, 1, 0).
		SetColumns(0, 1).
		AddItem(tview.NewTextView().SetText(title).
			SetTextAlign(tview.AlignCenter), 0, 0, 2, 2, 0, 0, false).
		AddItem(table, 2, 0, 1, 1, 0, 0, true)
	grid.SetBorder(true)
	return dialog(grid, width, height)
}

func about(w, h int) tview.Primitive { //nolint:ireturn,varnamelen
	table := tview.NewTable()
	for i, key := range mainKeys {
		table.SetCell(i, 0, tview.NewTableCell(key.name).
			SetAlign(tview.AlignCenter).SetExpansion(1).SetTextColor(tcell.ColorGrey))
		table.SetCell(i, 1, tview.NewTableCell(key.help).
			SetAlign(tview.AlignLeft).SetExpansion(1).SetTextColor(tcell.ColorBlue))
	}
	about := "\n\nbboltEdit\n\nVersion 0.1.2\n\n© 2025 Matthew R Kasun\n\nhttps://github.com/devilcove/bboltEdit" //nolint:lll
	grid := tview.NewGrid().
		SetRows(0, 1, 1, 1, 1, 0).
		SetColumns(0, 1).
		AddItem(tview.NewTextView().SetText(about).
			SetTextAlign(tview.AlignCenter), 0, 0, 3, 1, 0, 0, false).
		AddItem(tview.NewTextView().SetText("Application Key Bindings").
			SetTextAlign(tview.AlignCenter), 3, 0, 2, 2, 0, 0, false).
		AddItem(table, 5, 0, 1, 1, 0, 0, true)
	grid.SetBorder(true)
	return dialog(grid, w, h)
}

func errorView(message string) *tview.Modal {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Close"}).
		SetBackgroundColor(tcell.ColorBlueViolet)
	modal.SetTitle("Error")
	return modal
}

func showError(message string) {
	dialog := errorView(message)
	pager.AddPage("error", dialog, true, true)
	app.SetFocus(dialog)
}
