package main

import (
	"strings"

	"github.com/rivo/tview"
)

func moveForm(node dbNode) *tview.Form {
	currentPath := strings.Join(node.path, " ")
	form := tview.NewForm().
		AddTextView("current path", currentPath, 40, 2, true, true).
		AddInputField("new path", currentPath, 40, nil, nil).
		AddButton("Cancel", func() {
			pager.RemovePage("move")
			app.SetFocus(tree)
		})
	form.AddButton("Submit", func() {
		newpath := strings.Split(form.GetFormItem(1).(*tview.InputField).GetText(), " ")
		node := getCurrentNode("move")
		if err := moveItem(node, newpath); err != nil {
			errDisp.SetText(err.Error())
			pager.ShowPage("error").RemovePage("move")
			return
		}
		reloadAndSetSelection(newpath)
		pager.RemovePage("move")
		app.SetFocus(tree)
	})
	return form
}
