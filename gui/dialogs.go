package main

import (
	"encoding/json"
	"errors"
	"strings"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/texteditor"
)

func createBucketDialog(path string, button *core.Button) {
	d := core.NewBody("Create Bucket")
	core.NewText(d).SetText("Parent Bucket")
	parent := core.NewTextField(d).SetText(path)
	core.NewText(d).SetText("Bucket Name")
	name := core.NewTextField(d)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if strings.Contains(name.Text(), " ") {
				core.MessageDialog(button, "bucket name cannot contain spaces", "Create Bucket")
				return
			}
			path := []string{}
			if parent.Text() != "" {
				path = strings.Split(parent.Text(), " ")
			}
			path = append(path, name.Text())
			if _, err := CreateBucket(path); err != nil {
				core.ErrorDialog(button, err, "Create Bucket")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}

func deleteBucketDialog(path string, button *core.Button) {
	d := core.NewBody("Delete Bucket")
	core.NewText(d).SetText("Path")
	deletePath := core.NewTextField(d).SetText(path)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			path := strings.Split(deletePath.Text(), " ")
			if err := DeleteBucket(path); err != nil {
				core.ErrorDialog(button, err, "Delete Bucket")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}

func emptyBucketDialog(path string, button *core.Button) {
	d := core.NewBody("Empty Bucket")
	core.NewText(d).SetText("Path")
	deletePath := core.NewTextField(d).SetText(path)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			path := strings.Split(deletePath.Text(), " ")
			if err := EmptyBucket(path); err != nil {
				core.ErrorDialog(button, err, "Empty Bucket")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}

func addKeyDialog(path string, button *core.Button) {
	d := core.NewBody("Add Key")
	core.NewText(d).SetText("Parent Bucket")
	parent := core.NewTextField(d).SetText(path)
	core.NewText(d).SetText("Key Name")
	name := core.NewTextField(d)
	core.NewText(d).SetText("Key Value")
	value := texteditor.NewEditor(d).Buffer
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		core.NewButton(bar).SetText("validate json").OnClick(func(e events.Event) {
			if json.Valid(value.Text()) {
				core.MessageSnackbar(bar, "valid json")
			} else {
				core.MessageSnackbar(bar, "not valid json")
			}
		})
		d.AddOK(bar).OnClick(func(e events.Event) {
			if strings.Contains(name.Text(), " ") {
				core.ErrorDialog(button, errors.New("key name cannot contain spaces"), "Add Key")
				return
			}
			path := strings.Split(parent.Text(), " ")
			if err := CreateKey(name.Text(), toJSON(value.Text()), path); err != nil {
				core.ErrorDialog(button, err, "Add Key")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}

func moveBucketDialog(path string, button *core.Button) {
	d := core.NewBody("Move Bucket")
	core.NewText(d).SetText("Current Path")
	current := core.NewTextField(d).SetText(path)
	core.NewText(d).SetText("New Path")
	newPath := core.NewTextField(d)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if err := MoveBucket(strings.Split(current.Text(), " "),
				strings.Split(newPath.Text(), " ")); err != nil {
				core.ErrorDialog(button, err, "Add Key")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}

func moveKeyDialog(path string, button *core.Button) {
	d := core.NewBody("Move Key")
	core.NewText(d).SetText("Current Path")
	current := core.NewTextField(d).SetText(path)
	core.NewText(d).SetText("New Path")
	newPath := core.NewTextField(d)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if err := MoveKey(strings.Split(current.Text(), " "),
				strings.Split(newPath.Text(), " ")); err != nil {
				core.ErrorDialog(button, err, "Add Key")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}

func deleteKeyDialog(path string, button *core.Button) {
	d := core.NewBody("Delete Key")
	core.NewText(d).SetText("Path")
	deletionPath := core.NewTextField(d).SetText(path)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if err := DeleteKey(strings.Split(deletionPath.Text(), " ")); err != nil {
				core.ErrorDialog(button, err, "Delete Key")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}

func renameKeyDialog(path string, button *core.Button) {
	key := strings.Split(path, " ")
	d := core.NewBody("Rename Key")
	core.NewText(d).SetText("Path")
	core.NewText(d).SetText(strings.Join(key[0:len(key)-1], " "))
	core.NewText(d).SetText("Current Name")
	core.NewText(d).SetText(key[len(key)-1])
	core.NewText(d).SetText("New Name")
	newName := core.NewTextField(d)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if strings.Contains(newName.Text(), " ") {
				core.ErrorDialog(button, errors.New("key name cannot contain spaces"), "Rename Key")
				return
			}
			if err := RenameKey(key, newName.Text()); err != nil {
				core.ErrorDialog(button, err, "Rename Key")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}

func renameBucketDialog(orig string, button *core.Button) {
	path := strings.Split(orig, " ")
	d := core.NewBody("Rename Bucket")
	core.NewText(d).SetText("Path")
	core.NewText(d).SetText(strings.Join(path[0:len(path)-1], " "))
	core.NewText(d).SetText("Current Name")
	core.NewText(d).SetText(path[len(path)-1])
	core.NewText(d).SetText("New Name")
	newName := core.NewTextField(d)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if strings.Contains(newName.Text(), " ") {
				core.ErrorDialog(button, errors.New("bucket name cannot contain spaces"), "Rename Bucket")
				return
			}
			if err := RenameBucket(path, newName.Text()); err != nil {
				core.ErrorDialog(button, err, "Rename Bucket")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}

func copyKeyDialog(path string, button *core.Button) {
	d := core.NewBody("Copy Key")
	core.NewText(d).SetText("Key Path")
	core.NewText(d).SetText(path)
	core.NewText(d).SetText("New Path")
	newPath := core.NewTextField(d)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if err := CopyKey(strings.Split(path, " "), strings.Split(newPath.Text(), " ")); err != nil {
				core.ErrorDialog(button, err, "Copy Key")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}

func copyBucketDialog(path string, button *core.Button) {
	d := core.NewBody("Copy Bucket")
	core.NewText(d).SetText("Bucket Path")
	core.NewText(d).SetText(path)
	core.NewText(d).SetText("New Path")
	newPath := core.NewTextField(d)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if err := CopyBucket(strings.Split(path, " "), strings.Split(newPath.Text(), " ")); err != nil {
				core.ErrorDialog(button, err, "Copy Bucket")
				return
			}
			reload()
		})
	})
	d.RunDialog(button)
}
