package main

import (
	"log"
	"strings"

	"cogentcore.org/core/core"
)

func loadFile(filepath string) error {
	db.Close()
	if err := openDB(filepath); err != nil {
		return err
	}
	reload()
	return nil
}

func reload() {
	log.Println("reloading nodes")
	path := strings.Split(dbFile, "/")
	root := path[len(path)-1]
	nodes := getNodes()
	panes.AsFrame().NodeBase.DeleteChildren()
	left := core.NewFrame(panes)
	core.NewFrame(panes)
	tr := core.NewTree(left).SetText(root)
	addNodes(tr, nodes)
	tr.Scene.ContextMenus = nil
	tr.ContextMenus = nil
	tr.SetReadOnly(true)
	//tr.Update()
	//panes.Update()
	log.Println("updating app window")
	app.Update()
}
