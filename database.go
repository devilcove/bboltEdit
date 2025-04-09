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

func renameBucket(node dbNode, name string) error {
	newBucket := &bbolt.Bucket{}
	oldBucket := &bbolt.Bucket{}
	oldName := node.path[len(node.path)-1]
	err := db.Update(func(tx *bbolt.Tx) error {
		b, err := getParentBucket(node.path, tx)
		if err != nil {
			return err
		}
		if b == nil {
			newBucket, err = tx.CreateBucket([]byte(name))
			if err != nil {
				return err
			}
			oldBucket = tx.Bucket([]byte(oldName))
		} else {
			newBucket, err = b.CreateBucket([]byte(name))
			if err != nil {
				return err
			}
			oldBucket = b.Bucket([]byte(oldName))
			if oldBucket == nil {
				return errors.New("invalid path: bucket does not exist")
			}
		}
		err = oldBucket.ForEach(func(k, v []byte) error {
			if err := newBucket.Put(k, v); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		if b == nil {
			if err := tx.DeleteBucket([]byte(oldName)); err != nil {
				return err
			}
		} else {
			if err := b.DeleteBucket([]byte(oldName)); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func getParentBucket(path []string, tx *bbolt.Tx) (*bbolt.Bucket, error) {
	if len(path) == 1 {
		//parent is root
		return nil, nil
	}
	return getBucket(path[:len(path)-1], tx)
}

func getBucket(path []string, tx *bbolt.Tx) (*bbolt.Bucket, error) {
	bucket := tx.Bucket([]byte(path[0]))
	if bucket == nil {
		return &bbolt.Bucket{}, errors.New("invalid path: bucket does not exit")
	}
	for _, p := range path[1:] {
		bucket = bucket.Bucket([]byte(p))
	}
	if bucket == nil {
		return &bbolt.Bucket{}, errors.New("invalid path: bucket does not exit")
	}
	return bucket, nil
}

func deleteEntry(node dbNode) error {
	if node.kind == "bucket" {
		return deleteBucket(node)
	}
	return deleteKey(node)
}

func deleteBucket(node dbNode) error {
	name := node.path[len(node.path)-1]
	return db.Update(func(tx *bbolt.Tx) error {
		parent, err := getParentBucket(node.path, tx)
		if err != nil {
			return err
		}
		if parent == nil {
			return tx.DeleteBucket([]byte(name))
		}
		return parent.DeleteBucket([]byte(name))
	})
}

func deleteKey(node dbNode) error {
	name := node.path[len(node.path)-1]
	return db.Update(func(tx *bbolt.Tx) error {
		parent, err := getParentBucket(node.path, tx)
		if err != nil {
			return err
		}
		if err := parent.Delete([]byte(name)); err != nil {
			return err
		}
		return nil
	})
}

func emptyBucket(node dbNode) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := getBucket(node.path, tx)
		if err != nil {
			return err
		}
		err = bucket.ForEach(func(k, v []byte) error {
			if v == nil {
				return bucket.DeleteBucket(k)
			}
			return bucket.Delete(k)
		})
		return nil
	})
}

func addBucket(node dbNode, name string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		if node.path == nil {
			if _, err := tx.CreateBucket([]byte(name)); err != nil {
				return err
			}
			return nil
		}
		bucket, err := getBucket(node.path, tx)
		if err != nil {
			return err
		}
		_, err = bucket.CreateBucket([]byte(name))
		return err
	})
}

func addKey(node dbNode, name, value string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := getBucket(node.path, tx)
		if err != nil {
			return err
		}
		if bucket.Get([]byte(name)) != nil {
			return errors.New("key exists")
		}
		return bucket.Put([]byte(name), []byte(value))
	})
}

func moveItem(node dbNode, newpath []string) error {
	if node.kind == "bucket" {
		return moveBucket(node, newpath)
	}
	return moveKey(node, newpath)
}

func moveKey(node dbNode, path []string) error {
	if len(path) < 2 {
		return errors.New("invalid path, destination too short")
	}
	return db.Update(func(tx *bbolt.Tx) error {
		parent, err := getParentBucket(node.path, tx)
		if err != nil {
			return err
		}
		if err := parent.Delete(node.name); err != nil {
			return err
		}

		bucket, err := createParentBucket(path, tx)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(path[len(path)-1]), []byte(node.value))
	})
}

func moveBucket(node dbNode, path []string) error {
	newname := path[len(path)-1]
	if newname != string(node.name) {
		//need to rename node first
		if err := renameBucket(node, newname); err != nil {
			return err
		}
		node.name = []byte(newname)
	}
	return db.Update(func(tx *bbolt.Tx) error {
		parent, err := getParentBucket(node.path, tx)
		if err != nil {
			return err
		}
		newparent, err := createParentBucket(path, tx)
		if err != nil {
			return err
		}
		if parent == newparent {
			// this is a rename operation already done above
			return nil
		}
		return tx.MoveBucket([]byte(node.name), parent, newparent)
	})
}

func createParentBucket(path []string, tx *bbolt.Tx) (*bbolt.Bucket, error) {
	if len(path) == 1 {
		return nil, nil
	}
	return createBucket(path[:len(path)-1], tx)

}

func createBucket(path []string, tx *bbolt.Tx) (*bbolt.Bucket, error) {
	if path == nil {
		return nil, errors.New("invalid path")
	}
	//create root bucket
	bucket, err := tx.CreateBucketIfNotExists([]byte(path[0]))
	if err != nil {
		return nil, err
	}
	// create nested bucket(s)
	for _, p := range path[1:] {
		bucket, err = bucket.CreateBucketIfNotExists([]byte(p))
		if err != nil {
			return nil, err
		}
	}
	return bucket, nil
}

func editNode(node dbNode, update string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := getParentBucket(node.path, tx)
		if err != nil {
			return err
		}
		return bucket.Put(node.name, []byte(update))
	})
}
