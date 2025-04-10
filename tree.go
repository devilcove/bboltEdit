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

			case 'b':
				_, node := getCurrentNodes()
				bucket := dialog(addBucketForm(node, "bucket"), 40, 10)
				pager.AddPage("bucket", bucket, true, true)
				return nil

			case 'd':
				node := tree.GetCurrentNode()
				if node.GetReference() == nil {
					errDisp.SetText("cannot delete root node")
					pager.ShowPage("error")
					return nil
				}
				delete := modal(deleteForm(node.GetText()), 40, 7)
				pager.AddPage("delete", delete, true, true)
				return nil

			case 'e':
				_, node := getCurrentNodes()
				if node.path == nil {
					errDisp.SetText("not applicable to root node")
					pager.ShowPage("error")
					return nil
				}
				if node.kind == "bucket" {
					empty := dialog(emptyForm(node, "empty"), 60, 7)
					pager.AddPage("empty", empty, true, true)
					log.Println("focus empty modal")
					app.SetFocus(empty)
					//return nil
				} else {
					log.Println("edit key")
					edit := dialog(editForm(node, "edit"), 60, 20)
					pager.AddPage("edit", edit, true, true)
					return nil
				}

			case 'k':
				_, node := getCurrentNodes()
				if node.path == nil {
					errDisp.SetText("cannot add key to root")
					pager.ShowPage("error")
					return nil
				}
				key := modal(addKeyForm(node, "key"), 60, 22)
				pager.AddPage("key", key, true, true)
				return nil

			case 'm':
				node := tree.GetCurrentNode()
				if node.GetReference() == nil {
					errDisp.SetText("cannot move root node")
					pager.ShowPage("error")
					return nil
				}
				ref := node.GetReference().([]string)
				dbNode, ok := dbNodes[strings.Join(ref, " -> ")]
				if !ok {
					errDisp.SetText("invalid node")
					pager.ShowPage("error")
					return nil
				}
				move := modal(moveForm(dbNode), 60, 10)
				pager.AddPage("move", move, true, true)

			case 'r':
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
				rename := modal(renameForm(node.GetText()), 40, 10)
				pager.AddPage("rename", rename, true, true)
				return nil

			}
		}
		return event
	})
	tree.SetBorder(true).SetTitle("bbolt db viewer").SetTitleAlign(tview.AlignCenter)
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

//	func emptyForm(name string) *tview.Grid {
//		first := tview.NewInputField().SetLabel("Bucket to empty ").SetText(name)
//		second := tview.NewTextView().SetLabel("press esc to cancel, enter to accept")
//		form := tview.NewGrid().
//			SetColumns(0, 40, 0).
//			SetRows(1, 1, 1).
//			//SetBorders(true).
//			AddItem(first, 1, 1, 1, 1, 0, 0, false).
//			AddItem(second, 3, 1, 1, 1, 0, 0, true)

func modal(p tview.Primitive, w, h int) tview.Primitive {
	modal := tview.NewGrid().
		SetColumns(0, w, 0).
		SetRows(0, h, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Println("modal key handler", event.Key())
		return event
	})
	return modal

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

func getCurrentNode(modal string) dbNode {
	if tree.GetCurrentNode().GetReference() == nil {
		return dbNode{}
	}
	key := tree.GetCurrentNode().GetReference().([]string)
	node, ok := dbNodes[strings.Join(key, " -> ")]
	if !ok {
		errDisp.SetText("no node: " + strings.Join(key, ":"))
		pager.ShowPage("error").HidePage(modal)
	}
	return node
}

func reloadAndSetSelection(path []string) {
	reloadDB()
	root := tree.GetRoot()
	root.SetChildren(getNodes())
	tree.SetRoot(root)
	selectNode(path)
}

func getCurrentNodes() (*tview.TreeNode, dbNode) {
	treeNode := tree.GetCurrentNode()
	reference := treeNode.GetReference()
	if reference == nil {
		return treeNode, dbNode{
			path: nil,
			kind: "bucket",
		}
	}
	path := reference.([]string)
	node, ok := dbNodes[strings.Join(path, " -> ")]
	if !ok {
		log.Println("involid node", path)
	}
	return treeNode, node
}
