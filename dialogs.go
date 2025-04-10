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

func addKeyForm(node dbNode, dialog string) *tview.Form {
	form := tview.NewForm().
		AddInputField("path:", strings.Join(node.path, " "), 0, nil, nil).
		AddInputField("name", "", 0, nil, nil).
		AddTextArea("value", "", 0, 12, 0, nil).
		AddButton("Cancel", func() {
			pager.RemovePage(dialog)
			app.SetFocus(tree)
		})
	form.AddButton("Add", func() {
		newpath := strings.Split(form.GetFormItem(0).(*tview.InputField).GetText(), " ")
		name := form.GetFormItem(1).(*tview.InputField).GetText()
		value := form.GetFormItem(2).(*tview.TextArea).GetText()
		if err := addKey(newpath, name, value); err != nil {
			errDisp.SetText(err.Error())
			pager.ShowPage("error").RemovePage(dialog)
			return
		}
		reloadAndSetSelection(append(newpath, name))
		tree.GetCurrentNode().Expand()
		pager.RemovePage(dialog)
		app.SetFocus(tree)
	})
	form.SetButtonsAlign(tview.AlignCenter).
		SetBorder(true).SetTitle("Add Key").SetTitleAlign(tview.AlignCenter)
	return form
}

func addBucketForm(node dbNode, dialogName string) *tview.Form {
	form := tview.NewForm().
		AddInputField("parent bucket:", strings.Join(node.path, " "), 0, nil, nil).
		AddInputField("bucket name:", "", 0, nil, nil).
		AddButton("Cancel", func() {
			pager.RemovePage(dialogName)
			app.SetFocus(tree)
		})
	form.AddButton("Add", func() {
		path := []string{}
		parent := form.GetFormItem(0).(*tview.InputField).GetText()
		if parent != "" {
			path = strings.Split(parent, " ")
		}
		name := form.GetFormItem(1).(*tview.InputField).GetText()
		if err := addBucket(path, name); err != nil {
			errDisp.SetText(err.Error())
			pager.ShowPage("error").HidePage(dialogName)
			return
		}
		reloadAndSetSelection(append(path, name))
		tree.GetCurrentNode().Expand()
		pager.RemovePage(dialogName)
		app.SetFocus(tree)
	}).AddTextView("to create root bucket", "use empty parent bucket", 0, 2, true, false)
	form.SetButtonsAlign(tview.AlignCenter).
		SetBorder(true).SetTitle("Add Bucket").SetTitleAlign(tview.AlignCenter)
	return form
}

func deleteForm(name string) *tview.Form {
	form := tview.NewForm()
	form.AddTextView("name:", name, 0, 1, false, false)
	form.AddButton("Cancel", func() {
		pager.HidePage("delete")
	}).AddButton("Delete", func() {
		key := tree.GetCurrentNode().GetReference().([]string)
		node, ok := dbNodes[strings.Join(key, " -> ")]
		if !ok {
			errDisp.SetText("no node: " + strings.Join(key, ":"))
			pager.ShowPage("error").HidePage("delete")
		}
		if err := deleteEntry(node); err != nil {
			errDisp.SetText(err.Error())
			pager.ShowPage("error").HidePage("delete")
		}
		newpath := node.path[:len(node.path)-1]
		reloadAndSetSelection(newpath)
		selectNode(newpath)
		tree.GetCurrentNode().Expand()
		pager.HidePage("delete")
	}).
		SetButtonsAlign(tview.AlignCenter)
	form.SetBorder(true).SetTitle("Delete Item").SetTitleAlign(tview.AlignCenter)
	return form
}

func emptyForm(node dbNode, dialog string) *tview.Form {
	form := tview.NewForm().
		AddTextView("path:", strings.Join(node.path, " "), 0, 1, true, true).
		AddButton("Cancel", func() {
			pager.RemovePage(dialog)
			app.SetFocus(tree)
		}).
		AddButton("Empty", func() {
			if err := emptyBucket(node); err != nil {
				errDisp.SetText(err.Error())
				pager.ShowPage("error").RemovePage(dialog)
				return
			}
			reloadAndSetSelection(node.path)
			pager.RemovePage(dialog)
			app.SetFocus(tree)
		}).
		SetButtonsAlign(tview.AlignCenter)
	form.SetBorder(true).SetTitle("Empty Bucket").SetTitleAlign(tview.AlignCenter)
	return form
}

func moveForm(node dbNode) *tview.Form {
	currentPath := strings.Join(node.path, " ")
	form := tview.NewForm().
		AddTextView("current path", currentPath, 0, 1, true, true).
		AddInputField("new path", currentPath, 0, nil, nil).
		AddButton("Cancel", func() {
			pager.RemovePage("move")
			app.SetFocus(tree)
		}).
		SetButtonsAlign(tview.AlignCenter)
	form.AddButton("Submit", func() {
		newpath := strings.Split(form.GetFormItem(1).(*tview.InputField).GetText(), " ")
		node := getCurrentNode("move")
		log.Println("moving from", node.path, "to", newpath)
		if err := moveItem(node, newpath); err != nil {
			errDisp.SetText(err.Error())
			pager.ShowPage("error").RemovePage("move")
			return
		}
		reloadAndSetSelection(newpath)
		pager.RemovePage("move")
		app.SetFocus(tree)
	})
	form.SetBorder(true).SetTitle("Move Item").SetTitleAlign(tview.AlignCenter)
	return form
}

func editForm(node dbNode, dialog string) *tview.Form {
	form := tview.NewForm().
		AddTextView("path:", strings.Join(node.path, " "), 0, 1, true, false).
		AddTextArea("value:", string(node.value), 0, 12, 0, nil).
		AddButton("cancel", func() {
			pager.RemovePage(dialog)
			app.SetFocus(tree)
		})
	form.AddButton("Submit", func() {
		if err := editNode(node, form.GetFormItem(1).(*tview.TextArea).GetText()); err != nil {
			errDisp.SetText(err.Error())
			pager.ShowPage("error").RemovePage(dialog)
			return
		}
		reloadAndSetSelection(node.path)
		pager.RemovePage(dialog)
		app.SetFocus(tree)
	})
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetBorder(true).SetTitle("Edit Key").SetTitleAlign(tview.AlignCenter)
	log.Println(tview.DefaultFormFieldHeight, tview.DefaultFormFieldWidth)
	return form
}
