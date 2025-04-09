package main

import (
	"log"
	"strings"

	"github.com/rivo/tview"
)

func dialog(p tview.Primitive, w, h int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, h, 1, true).
			AddItem(nil, 0, 1, false), w, 1, true).
		AddItem(nil, 0, 1, false)
}

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

func editForm(node dbNode) *tview.Form {
	form := tview.NewForm().
		AddTextView("name:", string(node.name), 20, 1, true, false).
		AddTextArea("value:", string(node.value), 0, 12, 0, nil).
		AddButton("cancel", func() {
			pager.RemovePage("edit")
			app.SetFocus(tree)
		})
	form.AddButton("Submit", func() {
		log.Println("edit key", node.name, node.path)
		if err := editNode(node, form.GetFormItem(1).(*tview.TextArea).GetText()); err != nil {
			errDisp.SetText(err.Error())
			pager.ShowPage("error").RemovePage("edit")
			return
		}
		reloadAndSetSelection(node.path)
		pager.RemovePage("edit")
		app.SetFocus(tree)
	})
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetBorder(true).SetTitle("Edit Key").SetTitleAlign(tview.AlignCenter)
	log.Println(tview.DefaultFormFieldHeight, tview.DefaultFormFieldWidth)
	return form
}

func newBucketForm() *tview.Form {
	form := tview.NewForm().
		AddInputField("name:", "", 0, nil, nil).
		AddButton("Cancel", func() {
			pager.RemovePage("bucket")
			app.SetFocus(tree)
		}).
		SetButtonsAlign(tview.AlignCenter)
	form.AddButton("Submit", func() {
		name := form.GetFormItem(0).(*tview.InputField).GetText()
		node := getCurrentNode("bucket")
		if err := addBucket(node, name); err != nil {
			errDisp.SetText(err.Error())
			pager.ShowPage("error").RemovePage("bucket")
			return
		}
		reloadAndSetSelection(append(node.path, name))
		pager.RemovePage("bucket")
		app.SetFocus(tree)
	})
	form.SetBorder(true).SetTitle("Add Bucket").SetTitleAlign(tview.AlignCenter)
	return form
}
