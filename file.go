package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ref struct {
	path  string
	isDir bool
}

var (
	top        *tview.TextView
	newRootDir = make(chan string)
)

func newFiles() *tview.Grid { //nolint:funlen
	rightKeys := []key{
		{"o", "open dialog to change directory"},
		{"p", "println node table to logs"},
		{"enter", "expand dir, select file"},
		{"?", "show this help"},
	}
	cwd, _ := os.Getwd()
	picker := fileTree(cwd)
	top = textView("Select file to view (" + cwd + ")")
	fileGrid := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0).
		// SetBorders(true).
		AddItem(top,
			0, 0, 1, 1, 0, 0, false).
		AddItem(textView("press enter to expand directory or select file"),
			2, 0, 1, 1, 0, 0, false).
		AddItem(picker, 1, 0, 1, 1, 0, 0, true)
	fileGrid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Println("file grid handler", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'o':
				selected := picker.GetCurrentNode().GetReference().(ref).path
				dirsearch := modal(dirForm("dir", selected, newRootDir), 40, 10)
				pager.AddPage("dir", dirsearch, true, true)
				pager.SendToFront("dir")
			case 'p':
				current := picker.GetRoot()
				log.Println("root", current.GetText())
				for _, child := range current.GetChildren() {
					log.Println("child", child.GetText())
				}
			case '?':
				help := helpDialog("Key Bindings", 100, 10, rightKeys, treeMoveKeys)
				pager.AddPage("help", help, true, true)
				app.SetFocus(help)
				return nil
			}

		case tcell.KeyEnter:
			r := picker.GetCurrentNode().GetReference()
			if r == nil {
				return nil
			}
			node := r.(ref)
			if !node.isDir {
				log.Println("selected file", node.path)
				if err := InitDatabase(node.path); err != nil {
					showError(err.Error())
					return nil
				}
				tree = newTree(details)
				grid = mainGrid()
				pager.AddPage("main", grid, true, true).RemovePage("file")
				app.SetFocus(tree)
				return nil
			}
		}
		return event
	})
	children := picker.GetCurrentNode().GetChildren()
	for _, child := range children {
		reference := child.GetReference().(ref)
		log.Println(child.GetText(), reference.path, reference.isDir)
	}
	go func(c chan string) {
		for dir := range c {
			log.Println("changing to dir", dir)
			app.QueueUpdateDraw(func() {
				fileGrid.RemoveItem(picker)
				fileGrid.RemoveItem(top)
				picker = fileTree(dir)
				root := picker.GetRoot()
				log.Println("new root for file picker")
				for _, child := range root.GetChildren() {
					log.Println("child", child.GetText())
				}
				fileGrid.AddItem(picker, 1, 0, 1, 1, 0, 0, true)
				top.SetText(dir)
				fileGrid.AddItem(top, 0, 0, 1, 1, 0, 0, false)
				file := dialog(fileGrid, 60, 30)
				pager.AddPage("file", file, true, true)
				app.Sync()
				app.SetFocus(picker)
			})
		}
	}(newRootDir)

	fileGrid.SetBorder(true)

	return fileGrid
}

func fileTree(dir string) *tview.TreeView {
	rootDir := ".."
	root := tview.NewTreeNode(rootDir).
		SetColor(tcell.ColorRed)
	root.SetReference(ref{
		path:  dir,
		isDir: true,
	})
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	// Add the current directory to the root node.
	add(root, dir)

	// If a directory was selected, open it.
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference, ok := node.GetReference().(ref)
		if !ok {
			log.Println("invalid type assertion")
			return
		}
		log.Println("processing node", reference.path, reference.isDir)
		if strings.Contains(node.GetText(), "..") {
			log.Println("reloding parent dir")
			children := node.GetChildren()
			for _, child := range children {
				node.RemoveChild(child)
			}
			dir = filepath.Dir(dir)
			top.SetText("select file to view (" + dir + ")")
			root.SetReference(ref{
				path:  dir,
				isDir: true,
			})
			root.SetText("..")
			add(node, dir)
			node.SetExpanded(true)
			return
		}
		if reference.isDir {
			log.Println("handle", node.GetText(), reference.path)
			children := node.GetChildren()
			log.Println(len(children), "children")
			if len(children) == 0 {
				// Load and show files in this directory.
				add(node, reference.path)
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
	log.Println("adding", target.GetText(), "to", path)
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
		log.Println("added node", entry.Name(), ref.path, ref.isDir)
		target.AddChild(node)
	}
}
