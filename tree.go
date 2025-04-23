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

func newTree(detail *tview.TextView) *tview.TreeView { //nolint:funlen
	treeKeys := []key{
		{"c", "(c)opy key or bucket"},
		{"b", "create new (b)ucket"},
		{"d", "(d)elete key or bucket"},
		{"e", "(e)mpty bucket or (e)dit key"},
		{"a", "(a)dd new key"},
		{"m", "(m)ove key or bucket"},
		{"o", "(o)pen file selection"},
		{"r", "(r)ename key or bucket"},
		{"s", "(s)earch for key or bucket"},
		{"x", "e(x)pand all nodes"},
		{"?", "show help"},
		{"Enter", "expand or colapse node"},
		{"Ctrl-R", "reload database"},
		{"Ctrl-C", "colapse all nodes"},
		{"Ctrl-C", "expand all nodes"},
	}

	rootDir := "."
	root := tview.NewTreeNode(rootDir).
		SetColor(tcell.ColorRed)
	root.SetChildren(getNodes())
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
		updateDetail(detail, node)
	})
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		updateDetail(detail, node)
	})
	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Println("tree key handler", tcell.KeyNames[event.Key()])
		switch event.Key() {
		// callapse all nodes
		case tcell.KeyCtrlC:
			tree.GetRoot().CollapseAll()
			// reload database
		case tcell.KeyCtrlR:
			reloadDB()
			root.SetChildren(getNodes())
			tree.SetRoot(root)
			// expand all nodes
		case tcell.KeyCtrlX:
			tree.GetRoot().ExpandAll()
			// exit app
		case tcell.KeyEsc:
			app.Stop()
			// change focue
		case tcell.KeyTAB:
			app.SetFocus(detail)
			// key handling
		case tcell.KeyRune:
			log.Println("tree key handler, runes", event.Rune())
			switch event.Rune() {
			// collapse node
			case 'c':
				node := getCurrentNode()
				copy := dialog(copyForm(node, "dialog"), 60, 12)
				pager.AddPage("dialog", copy, true, true)
			// add bucket
			case 'b':
				node := getCurrentNode()
				bucket := dialog(addBucketForm(node, "dialog"), 60, 12)
				pager.AddPage("dialog", bucket, true, true)
				return nil
			// delete bucket/key
			case 'd':
				node := getCurrentNode()
				if node.path == nil {
					showError("cannot delete root node")
					return nil
				}
				deleteItem := modal(deleteForm(node, "dialog"), 40, 7)
				pager.AddPage("dialog", deleteItem, true, true)
				return nil
			// empty bucket or edit key
			case 'e':
				node := getCurrentNode()
				if node.path == nil {
					showError("not applicable to root node")
					return nil
				}
				if node.kind == "bucket" { //nolint:goconst
					empty := dialog(emptyForm(node, "dialog"), 60, 7)
					pager.AddPage("dialog", empty, true, true)
					log.Println("focus empty modal")
					app.SetFocus(empty)
					return nil
				}
				log.Println("edit key")
				edit := dialog(editForm(node, "dialog"), 60, 20)
				pager.AddPage("dialog", edit, true, true)
				return nil
			// add key
			case 'a':
				node := getCurrentNode()
				if node.path == nil {
					showError("cannot add key to root")
					pager.ShowPage("error")
					return nil
				}
				key := modal(addKeyForm(node, "dialog"), 60, 22)
				pager.AddPage("dialog", key, true, true)
				return nil
			// move bucket/key
			case 'm':
				node := getCurrentNode()
				if node.path == nil {
					showError("cannot move root node")
					pager.ShowPage("error")
					return nil
				}
				move := modal(moveForm(node, "dialog"), 60, 10)
				pager.AddPage("dialog", move, true, true)
				return nil
			// open filepicker
			case 'o':
				file := dialog(newFiles(), 60, 30)
				pager.AddPage("file", file, true, true)
				app.SetFocus(file)
				return nil
			// rename bucket/key
			case 'r':
				node := getCurrentNode()
				if node.path == nil {
					showError("cannot rename root node")
					pager.ShowPage("error")
					return nil
				}
				rename := modal(renameForm(node, "dialog"), 40, 10)
				pager.AddPage("dialog", rename, true, true)
				return nil
			// show help
			case '?':
				help := helpDialog("Key Bindings", 100, 15, treeKeys, treeMoveKeys)
				pager.AddPage("help", help, true, true)
				app.SetFocus(help)
				return nil
			}
		}
		return event
	})
	tree.SetBorder(true).SetTitle("bbolt db viewer").SetTitleAlign(tview.AlignCenter)
	return tree
}

func updateDetail(detail *tview.TextView, node *tview.TreeNode) {
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
		value = fmt.Sprintf("Key:\n\nPath: %s\nName: %s\n\nValue:\n\n%s",
			strings.Join(entry.path, " -> "), string(entry.name), prettyString(entry.value))
	}
	detail.SetText(value)
}

func prettyString(s []byte) string {
	var data bytes.Buffer
	if err := json.Indent(&data, s, "", "\t"); err != nil {
		return string(s)
	}
	return data.String()
}

func modal(p tview.Primitive, w, h int) tview.Primitive { //nolint:ireturn
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

func reloadAndSetSelection(path []string) {
	reloadDB()
	root := tree.GetRoot()
	root.SetChildren(getNodes())
	tree.SetRoot(root)
	selectNode(path)
}

func getCurrentNode() dbNode {
	treeNode := tree.GetCurrentNode()
	reference := treeNode.GetReference()
	if reference == nil {
		return dbNode{
			path: nil,
			kind: "bucket",
		}
	}
	path := reference.([]string)
	node, ok := dbNodes[strings.Join(path, " -> ")]
	if !ok {
		log.Println("involid node", path)
		return dbNode{}
	}
	return node
}
