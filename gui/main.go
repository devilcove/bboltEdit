package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
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
	b := core.NewButton(app).SetText("File")
	b.SetMenu(func(m *core.Scene) {
		core.NewButton(m).SetText("Open File").OnClick(func(e events.Event) {
			current, _ := os.Getwd()
			d := core.NewBody("File Open")
			ft := filetree.NewTree(d).OpenPath(".")
			selected := ""
			ft.OnSelect(func(e events.Event) {
				ft.SelectedFunc(func(n *filetree.Node) {
					selected = string(n.Filepath)
				})
			})
			d.AddBottomBar(func(bar *core.Frame) {
				d.AddCancel(bar)
				core.NewButton(bar).SetText("Open Parent").OnClick(func(e events.Event) {
					current = filepath.Dir(current)
					ft.DeleteChildren()
					ft.OpenPath(current)
					d.Update()
				})
				d.AddOK(bar).OnClick(func(e events.Event) {
					log.Println("open file ", selected)
					if err := loadFile(selected); err != nil {
						core.ErrorDialog(d, err, "Open File")
					}
				})
				//fp.OnClose()
				//log.Println(fp.SelectedFile())
			})
			d.RunDialog(b)
		})
		core.NewButton(m).SetText("Quit").OnClick(func(e events.Event) {
			app.Close()
		})
	})
	b.Scene.ContextMenus = nil

	panes = core.NewSplits(app).SetSplits(.3, .7)
	left := core.NewFrame(panes)
	core.NewFrame(panes)

	tr := core.NewTree(left).SetText("test.db")
	addNodes(tr, nodes)
	tr.Scene.ContextMenus = nil
	tr.ContextMenus = nil
	tr.ContextMenus = append(tr.ContextMenus, mainContext)
	tr.SetReadOnly(true)

	app.RunMainWindow()
}

func addNodes(t *core.Tree, nodes []*TreeNode) {
	for _, node := range nodes {
		item := core.NewTree(t).SetText(string(node.Name))
		item.SetReadOnly(true)
		item.ContextMenus = nil
		if node.IsBucket {
			item.SetIcon(icons.AddCircle)
			item.ContextMenus = append(item.ContextMenus, bucketContext)
		} else {
			item.ContextMenus = append(item.ContextMenus, keyContext)
		}
		//item.ValueTitle = strings.Join(node.Path, " ")
		item.Name = strings.Join(node.Path, " ")
		item.OnSelect(func(e events.Event) {
			updateDetails(item.Name)
		})
		addNodes(item, node.Children)
	}
}

func mainContext(m *core.Scene) {
	button := core.NewButton(m).SetText("Create Bucket")
	button.OnClick(func(e events.Event) {
		createBucketDialog("", button)
	})
}

func keyContext(m *core.Scene) {
	button := core.NewButton(m)
	button.SetText("Delete Key").OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		deleteKeyDialog(path, button)
	})
	core.NewButton(m).SetText("Move Key").OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		moveKeyDialog(path, button)
	})
	core.NewButton(m).SetText("Rename Key").OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		renameKeyDialog(path, button)
	})
	core.NewButton(m).SetText("Copy Key").OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		copyKeyDialog(path, button)
	})
}

func bucketContext(m *core.Scene) {
	button := core.NewButton(m).SetText("Create Bucket")
	button.OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		createBucketDialog(path, button)
	})
	core.NewButton(m).SetText("Delete Bucket").OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		deleteBucketDialog(path, button)
	})
	core.NewButton(m).SetText("Empty Bucket").OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		emptyBucketDialog(path, button)
	})
	core.NewButton(m).SetText("Add Key").OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		addKeyDialog(path, button)
	})
	core.NewButton(m).SetText("Move Bucket").OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		moveBucketDialog(path, button)
	})
	core.NewButton(m).SetText("Rename Bucket").OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		renameBucketDialog(path, button)
	})
	core.NewButton(m).SetText("Copy Bucket").OnClick(func(e events.Event) {
		path := strings.ReplaceAll(m.This.AsTree().Name, "-menu", "")
		copyBucketDialog(path, button)
	})
}

func updateDetails(item string) {
	log.Println("updating details for ", item)
	node, ok := nodeMap[item]
	if !ok {
		log.Println("invalid node", item)
	}
	panes.AsFrame().NodeBase.DeleteChildAt(1)
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
		value := pretty(node.Value)
		buf := texteditor.NewEditor(details).Buffer.SetText(value)
		frame := core.NewFrame(details)
		core.NewButton(frame).SetText("Reset").OnClick(func(e events.Event) {
			buf.SetText(node.Value)
		})
		core.NewButton(frame).SetText("Validate Json").OnClick(func(e events.Event) {
			if json.Valid(buf.Text()) {
				core.MessageSnackbar(details, "valid json")
			} else {
				core.MessageSnackbar(details, "not valid json")
			}
		})
		core.NewButton(frame).SetText("Update").OnClick(func(e events.Event) {
			if err := UpdateKey(node, toJSON(buf.Text())); err != nil {
				core.ErrorDialog(details, err, "Update Key")
				return
			}
			reload()
		})
	}
	panes.Update()
}

func pretty(s []byte) []byte {
	var data bytes.Buffer
	if err := json.Indent(&data, s, "", "\t"); err != nil {
		return s
	}
	return data.Bytes()
}

func toJSON(orig []byte) []byte { //nolint:varnamelen
	var temp any
	if err := json.Unmarshal(orig, &temp); err != nil {
		return orig
	}
	bytes, err := json.Marshal(temp)
	if err != nil {
		return orig
	}
	return bytes
}
