package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func newTree(detail *tview.TextArea) *tview.TreeView {

	rootDir := "."
	root := tview.NewTreeNode(rootDir).
		SetColor(tcell.ColorRed)
	root.SetChildren(getNodes())
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root) //.
		//SetTopLevel(1)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
		updateDetail(detail, node)
	})
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		updateDetail(detail, node)
	})
	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Println("tree key handler", event.Key(), event.Rune(), event.Modifiers())
		switch event.Modifiers() {
		case tcell.ModAlt:
			switch event.Rune() {
			case 'c':
				tree.GetRoot().CollapseAll()
			case 'e':
				tree.GetRoot().ExpandAll()
			}
		case tcell.ModCtrl:
			switch event.Rune() {
			case 'r':
				reloadDB()
				root.SetChildren(getNodes())
				tree.SetRoot(root)
			}
		case tcell.ModNone:
			switch event.Rune() {
			case 'r':
				log.Println("alt-r")
				node := tree.GetCurrentNode()
				if node.GetReference() == nil {
					errDisp.SetText("cannot rename root node")
					pager.ShowPage("error")
					return nil
				}
				log.Println(node.GetReference(), node.GetLevel(), node.GetText())
				ref := node.GetReference().([]string)
				log.Println(ref)
				log.Println(dbNodes[strings.Join(ref, " -> ")])
				rename := modal(node.GetText())
				pager.AddPage("rename", rename, true, true)
				return nil
			case 'd':
				node := tree.GetCurrentNode()
				if node.GetReference() == nil {
					errDisp.SetText("cannot delete root node")
					pager.ShowPage("error")
					return nil
				}
				//ref := node.GetReference().([]string)
				delete := delModal(node.GetText())
				pager.AddPage("delete", delete, true, true)
				return nil
			}
		}
		return event
	})

	return tree

}

func updateDetail(detail *tview.TextArea, node *tview.TreeNode) {
	reference := node.GetReference()
	value := ""
	if reference != nil {
		value = strings.Join(reference.([]string), " -> ")
	}
	entry, ok := dbNodes[value]
	if !ok {
		log.Println("invalid value", value)
		return
	}
	if entry.kind == "bucket" {
		value = fmt.Sprintf("Bucket:\n\nPath: %s\nName: %s",
			strings.Join(entry.path, " -> "), string(entry.name))
	} else {
		value = fmt.Sprintf("Key:\n\nPath: %s\nName: %s\n\nValue:\n%s",
			strings.Join(entry.path, " -> "), string(entry.name), prettyString(entry.value))
	}
	detail.SetText(value, true)
}

func prettyString(s []byte) string {
	var data bytes.Buffer
	if err := json.Indent(&data, s, "", "\t"); err != nil {
		return string(s)
	}
	return data.String()
}

func renameForm(name string) *tview.Form {
	f := tview.NewForm()
	f.AddInputField("new name", name, 20, nil, nil)
	f.AddButton("cancel", func() {
		pager.HidePage("rename")
	}).
		AddButton("rename", func() {
			key := tree.GetCurrentNode().GetReference().([]string)
			node, ok := dbNodes[strings.Join(key, " -> ")]
			if !ok {
				log.Println("rename err: no node")
				errDisp.SetText("no node: " + strings.Join(key, ":"))
				pager.ShowPage("error")
				pager.HidePage("rename")
			}
			newName := f.GetFormItem(0).(*tview.InputField).GetText()
			if err := renameEntry(node, newName); err != nil {
				log.Println("rename err", err)
				errDisp.SetText(err.Error())
				pager.ShowPage("error")
				pager.HidePage("rename")
				return
			}
			reloadDB()
			root := tree.GetRoot()
			root.SetChildren(getNodes())
			tree.SetRoot(root)
			newpath := node.path
			newpath[len(node.path)-1] = newName
			selectNode(newpath)
			pager.HidePage("rename")
		})
	f.SetBorder(true).SetTitle("Rename").SetTitleAlign(tview.AlignCenter)
	return f
}

func deleteForm(name string) *tview.Form {
	form := tview.NewForm()
	form.AddTextView("name to delete", name, 20, 1, false, false)
	form.AddButton("cancel", func() {
		pager.HidePage("delete")
	}).AddButton("delete", func() {
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
		reloadDB()
		root := tree.GetRoot()
		root.SetChildren(getNodes())
		tree.SetRoot(root)
		newpath := node.path[:len(node.path)-1]
		log.Println("newpath after delete", newpath, node.path)
		selectNode(newpath)
		pager.HidePage("delete")
	})
	form.SetBorder(true).SetTitle("Delete").SetTitleAlign(tview.AlignCenter)

	return form
}

func modal(name string) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, 40, 0).
		SetRows(0, 10, 0).
		AddItem(renameForm(name), 1, 1, 1, 1, 0, 0, true)
}

func delModal(name string) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, 40, 0).
		SetRows(0, 10, 0).
		AddItem(deleteForm(name), 1, 1, 1, 1, 0, 0, true)
}

func selectNode(path []string) {
	node := tree.GetRoot()
	for _, name := range path {
		node = getChild(node, name)
	}
	for _, n := range tree.GetPath(node) {
		n.Expand()
	}
	tree.SetCurrentNode(node)
	fn := tree.GetSelectedFunc()
	fn(node)
}

func getChild(node *tview.TreeNode, name string) *tview.TreeNode {
	children := node.GetChildren()
	for _, child := range children {
		if child.GetText() == name {
			return child
		}
	}
	return nil
}
