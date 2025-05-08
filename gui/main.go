package main

import (
	"errors"
	"log"
	"strings"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/filetree"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/texteditor"
)

type TreeNode struct {
	Name     []byte
	IsBucket bool
	Value    []byte
	Path     []string
	//Click    rect.Rect
	Children []*TreeNode
}

type User struct {
	Name string
	Pass string
}

type Item struct {
	Type   *core.Text
	Path   *core.Text
	Name   *core.Text
	Value  *core.Text
	Editor *texteditor.Editor
}

var (
	app     *core.Body
	Details Item
	panes   *core.Splits
)

func main() {
	log.SetFlags(log.Lshortfile)
	openDB("test.db")
	nodes := getNodes()
	app = core.NewBody("BboltEdit")
	b := core.NewButton(app).SetText("Hello World")
	b.SetMenu(func(m *core.Scene) {
		core.NewButton(m).SetText("Open File").OnClick(func(e events.Event) {
			d := core.NewBody("File Open")
			ft := filetree.NewTree(d).OpenPath(".")
			selected := []string{}
			ft.OnSelect(func(e events.Event) {
				selected = []string{}
				ft.SelectedFunc(func(n *filetree.Node) {
					selected = append(selected, string(n.Filepath))
				})
			})
			panes := core.NewFrame(d)
			core.NewButton(panes).SetText("Close").OnClick(func(e events.Event) {
				d.Close()
			})
			core.NewButton(panes).SetText("Ok").OnClick(func(e events.Event) {
				log.Println(selected)
				if err := loadFile(strings.Join(selected, "")); err != nil {
					core.ErrorDialog(d, err, "Open File")
				}
				d.Close()
			})
			//fp.OnClose()
			//log.Println(fp.SelectedFile())
			d.RunDialog(b)
			log.Println(selected)
		})
		core.NewButton(m).SetText("Quit").OnClick(func(e events.Event) {
			app.Close()
		})
		core.NewButton(m).SetText("Send Message")
	})
	b.Scene.ContextMenus = nil
	b.ContextMenus = append(b.ContextMenus, func(m *core.Scene) {
		b = core.NewButton(m).SetText("Error")
		b.OnClick(func(e events.Event) {
			core.ErrorDialog(b, errors.New("this is an error"), "title")
		})
	})

	panes = core.NewSplits(app).SetSplits(.3, .7)
	left := core.NewFrame(panes)
	core.NewFrame(panes)

	tr := core.NewTree(left).SetText("test.db")
	addNodes(tr, nodes)
	tr.Scene.ContextMenus = nil
	tr.ContextMenus = nil
	tr.SetReadOnly(true)

	app.RunMainWindow()
}

func addNodes(t *core.Tree, nodes []*TreeNode) {
	for _, node := range nodes {
		item := core.NewTree(t).SetText(string(node.Name))
		item.SetReadOnly(true)
		if node.IsBucket {
			item.SetIcon(icons.AddCircle)
		}
		item.ContextMenus = nil
		item.ContextMenus = append(item.ContextMenus, nodeContext)
		item.Scene.Data = node
		item.ValueTitle = strings.Join(node.Path, " ")
		item.Name = strings.Join(node.Path, " ")
		item.OnSelect(func(e events.Event) {
			updateDetails(item.ValueTitle)
		})
		addNodes(item, node.Children)
	}
}

func nodeContext(m *core.Scene) {
	core.NewButton(m).SetText("Create").OnClick(func(e events.Event) {
		log.Println("Create")
	})
	core.NewButton(m).SetText("Delete").OnClick(func(e events.Event) {
		log.Println("info", m.This.AsTree().Parent)
	})
	core.NewButton(m).SetText("Empty").OnClick(func(e events.Event) {
		log.Println("Empty")
	})
	core.NewButton(m).SetText("Add Key").OnClick(func(e events.Event) {
		log.Println(m.This.AsTree().Name)
	})
	core.NewButton(m).SetText("Move").OnClick(func(e events.Event) {
		log.Println(m.This.AsTree().Name)
	})
	core.NewButton(m).SetText("Rename").OnClick(func(e events.Event) {
		log.Println(m.This.AsTree().Name)
	})
	core.NewButton(m).SetText("Search").OnClick(func(e events.Event) {
		log.Println(m.This.AsTree().Name)
	})
}

func updateDetails(item string) {
	log.Println("updating details for ", item)
	node, ok := nodeMap[item]
	if !ok {
		log.Println("invalid node", item)
	}
	log.Println(panes.Children)
	panes.AsFrame().NodeBase.DeleteChildAt(1)
	//panes.Children = panes.Children[:0]
	log.Println(panes.Children)
	details := core.NewFrame(panes)
	if node.IsBucket {
		core.NewText(details).SetText("Bucket:")
	} else {
		core.NewText(details).SetText("Key:")
	}
	core.NewSpace(details)
	core.NewText(details).SetText("Path:" + item)
	core.NewText(details).SetText("Name:" + string(node.Name))
	if !node.IsBucket {
		core.NewSpace(details)
		buf := texteditor.NewEditor(details).Buffer.SetText(node.Value)
		frame := core.NewFrame(details)
		core.NewButton(frame).SetText("Reset").OnClick(func(e events.Event) {
			buf.SetText(node.Value)
		})
		core.NewButton(frame).SetText("Validate Json").OnClick(func(e events.Event) {
			log.Println("validate json")
		})
		core.NewButton(frame).SetText("Update").OnClick(func(e events.Event) {
			log.Println("update key value")
		})
	}
	panes.Update()
}

// func setDetails(frame *core.Frame, item string) {
// 	node, ok := nodeMap[item]
// 	if !ok && item != "" {
// 		log.Println("invalid node")
// 	}
// 	frame.Cl
// 	core.NewText(frame).SetText

// }
