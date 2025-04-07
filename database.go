package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.etcd.io/bbolt"
)

var (
	db  *bbolt.DB
	old string
	//database   []Bucket
	dbNodes    map[string]dbNode = make(map[string]dbNode)
	dbError    error
	errInvalid error = errors.New("invalid bucket/key name")
)

type dbNode struct {
	path  []string
	kind  string
	name  []byte
	value []byte
}

func InitDatabase(file string) error {
	dbError = nil
	if db != nil {
		CloseDatabase()
	}
	var err error
	db, err = bbolt.Open(file, 0666, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		return err
	}
	old = file
	log.Println("loaded db file", file)
	header.SetText("bbolt database file: " + file)
	return nil
}

func reloadDB() {
	InitDatabase(old)
}

func CloseDatabase() {
	if db != nil {
		db.Close()
	}
}

func getNodes() []*tview.TreeNode {
	nodes := []*tview.TreeNode{}
	db.View(func(tx *bbolt.Tx) error {
		tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			node := process(name, nil, b)
			nodes = append(nodes, node)
			return nil
		})
		return nil
	})
	//log.Println(pretty.Sprint(dbNodes))
	//for k, v := range dbNodes {
	//log.Println(v.kind, string(v.name), k, v.path, string(v.value))
	//}
	return nodes
}

func process(name []byte, path []string, b *bbolt.Bucket) *tview.TreeNode {
	//log.Println("processing", string(name), path)
	path = append(path, string(name))
	dbNodes[strings.Join(path, " -> ")] = dbNode{
		path: path,
		name: name,
		kind: "bucket",
	}
	node := tview.NewTreeNode(string(name)).SetReference(path).
		SetSelectable(true).Collapse().SetColor(tcell.ColorGreen)
	b.ForEach(func(k, v []byte) error {
		if v == nil {
			nested := b.Bucket(k)
			child := process(k, path, nested)
			child.Collapse()
			node.AddChild(child)
		} else {
			childPath := append(path, string(k))
			node.AddChild(tview.NewTreeNode(string(k)).SetReference(childPath).
				SetSelectable(true)).Collapse()
			dbNodes[strings.Join(childPath, " -> ")] = dbNode{
				path:  childPath,
				kind:  "key",
				name:  k,
				value: v,
			}
		}
		return nil
	})

	return node
}

func renameEntry(node dbNode, value string) error {
	log.Println("rename entry", node.path, value)
	if node.kind == "bucket" {
		return renameBucket(node, value)
	}
	return renameKey(node, value)
}

func renameKey(node dbNode, value string) error {
	if db == nil {
		return errors.New("database not open")
	}
	name := node.path[len(node.path)-1]
	err := db.Update(func(tx *bbolt.Tx) error {
		b, err := getParentBucket(node.path, tx)
		if err != nil {
			return err
		}
		existing := b.Get([]byte(value))
		if existing != nil {
			return errors.New("key exists")
		}
		key := b.Get([]byte(name))
		if key == nil {
			return errors.New("invalid path: key does not exist")
		}
		if err := b.Put([]byte(value), key); err != nil {
			return err
		}
		if err := b.Delete([]byte(name)); err != nil {
			return err
		}
		return nil
	})
	return err

}

func renameBucket(node dbNode, value string) error {
	if db == nil {
		return errors.New("database not open")
	}
	err := db.Update(func(tx *bbolt.Tx) error {
		b, err := getParentBucket(node.path, tx)
		if err != nil {
			return err
		}
		newB, err := b.CreateBucket([]byte(value))
		if err != nil {
			return err
		}
		oldB := b.Bucket([]byte(node.path[len(node.path)-1]))
		if oldB == nil {
			return errors.New("invalid path: bucket does not exist")
		}
		err = oldB.ForEach(func(k, v []byte) error {
			if err := newB.Put(k, v); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func getParentBucket(path []string, tx *bbolt.Tx) (*bbolt.Bucket, error) {
	bucket := &bbolt.Bucket{}
	bucket = tx.Bucket([]byte(path[0]))
	if bucket == nil {
		return bucket, errors.New("invalid path: bucket does not exist")
	}
	for _, p := range path[1 : len(path)-1] {
		bucket = bucket.Bucket([]byte(p))
	}
	return bucket, nil
}
