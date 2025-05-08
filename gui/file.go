package main

import (
	"strings"

	"cogentcore.org/core/core"
)

func loadFile(filepath string) error {
	path := strings.Split(filepath, "/")
	db.Close()
	if err := openDB(filepath); err != nil {
		return err
	}
	nodes := getNodes()
	panes.AsFrame().NodeBase.DeleteChildren()
	left := core.NewFrame(panes)
	core.NewFrame(panes)
	tr := core.NewTree(left).SetText(path[len(path)-1])
	addNodes(tr, nodes)
	tr.Scene.ContextMenus = nil
	tr.ContextMenus = nil
	tr.SetReadOnly(true)
	panes.Update()
	return nil
}
