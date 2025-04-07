package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ref struct {
	path  string
	isDir bool
}

func newFiles() *tview.Grid {
	picker := fileTree()
	grid := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0).
		SetBorders(true).
		AddItem(textView("Select bbolt db file to view"),
			0, 0, 1, 1, 0, 0, false).
		AddItem(textView("press enter to expand directory or select file"),
			2, 0, 1, 1, 0, 0, false).
		AddItem(picker, 1, 0, 1, 1, 0, 0, true)
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			node := picker.GetCurrentNode().GetReference().(ref)
			if !node.isDir {
				log.Println("switching to", node.path)
				if err := InitDatabase(node.path); err != nil {
					errDisp.SetText(err.Error())
					pager.ShowPage("error")
					return nil
				}
				pager.SwitchToPage("main")
				fn := tree.GetInputCapture()
				fn(tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModCtrl))
				return nil
			}
		}
		return event
	})
	return grid
}

func fileTree() *tview.TreeView {
	rootDir := "."
	root := tview.NewTreeNode(rootDir).
		SetColor(tcell.ColorRed)
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	// Add the current directory to the root node.
	add(root, rootDir)

	//tree.SetDoneFunc(func(key tcell.Key) {
	//log.Println("file done func")
	//node := tree.GetCurrentNode().GetReference().(ref)
	//c <- node.path
	//tree = nil
	//})

	// If a directory was selected, open it.
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			log.Println("root node was selected")
			return // Selecting the root node does nothing.
		}
		ref, ok := reference.(ref)
		if !ok {
			log.Println("invalid type assertion")
			return
		}
		if ref.isDir {
			children := node.GetChildren()
			if len(children) == 0 {
				// Load and show files in this directory.
				add(node, node.GetText())
			} else {
				// Collapse if visible, expand if collapsed.
				node.SetExpanded(!node.IsExpanded())
			}
		}
	})
	return tree
}

// A helper function which adds the files and directories of the given path
// to the given target node.
func add(target *tview.TreeNode, path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Println(err)
		return
	}
	for _, entry := range entries {
		ref := ref{
			path:  filepath.Join(path, entry.Name()),
			isDir: entry.IsDir(),
		}
		node := tview.NewTreeNode(entry.Name()).
			SetReference(ref).
			SetSelectable(true)
		if entry.IsDir() {
			node.SetColor(tcell.ColorGreen)
		}
		target.AddChild(node)
	}
}
