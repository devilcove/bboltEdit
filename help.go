package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func helpView() *tview.Modal {
	return tview.NewModal().
		SetText("This would be the help menu").
		AddButtons([]string{"Close"}).
		SetBackgroundColor(tcell.ColorBlueViolet)
}

func errorView() *tview.Modal {
	modal := tview.NewModal().
		AddButtons([]string{"Close"}).
		SetBackgroundColor(tcell.ColorBlueViolet)
	modal.SetTitle("Error")
	return modal
}
