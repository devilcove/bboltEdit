package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

var (
	db             *bbolt.DB
	invalidPathErr = errors.New("invalid path")
	keyExistsErr   = errors.New("key exists")
	nodeMap        = make(map[string]TreeNode)
)

func openDB(file string) error {
	var err error
	if db != nil {
		closeDB()
	}
	db, err = bbolt.Open(file, 0o666, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		return err
	}
	return nil
}

func closeDB() {
	if db != nil {
		db.Close()
	}
}

func mapNodes(nodes []*TreeNode) {
	for _, node := range nodes {
		nodeMap[strings.Join(node.Path, " ")] = *node
		mapNodes(node.Children)
	}
}

func getNodes() []*TreeNode {
	allNodes := []*TreeNode{}
	db.View(func(tx *bbolt.Tx) error {
		tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			nodes := process(name, nil, b)
			allNodes = append(allNodes, nodes...)
			return nil
		})
		return nil
	})
	mapNodes(allNodes)
	return allNodes
}

func process(name []byte, path []string, b *bbolt.Bucket) []*TreeNode {
	nodes := []*TreeNode{}
	path = append(path, string(name))
	node := &TreeNode{
		Path:     path,
		Name:     name,
		IsBucket: true,
	}
	b.ForEach(func(k, v []byte) error {
		log.Println("checking", string(k), string(v))
		if v != nil {
			log.Println("key child", "key:", string(k), "value:", v, v == nil)
			child := &TreeNode{
				Path:     append(path, string(k)),
				IsBucket: false,
				Name:     k,
				Value:    v,
			}
			node.Children = append(node.Children, child)
		} else {
			log.Println("bucket child", string(k))
			nested := b.Bucket(k)
			children := process(k, path, nested)
			node.Children = append(node.Children, children...)
		}
		return nil
	})
	nodes = append(nodes, node)
	log.Println("add node", string(node.Name), node.Path, node.IsBucket)
	return nodes
}

func CreateBucket(path []string) (*bbolt.Bucket, error) {
	var bucket *bbolt.Bucket
	if path == nil {
		return nil, invalidPathErr
	}
	err := db.Update(func(tx *bbolt.Tx) (err error) {
		bucket, err = tx.CreateBucketIfNotExists([]byte(path[0]))
		if err != nil {
			return err
		}
		for _, p := range path[1:] {
			bucket, err = bucket.CreateBucketIfNotExists([]byte(p))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return bucket, err
}

func CreateKey(name, value string, path []string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := addBucket(path, tx)
		if err != nil {
			return err
		}
		if bucket.Get([]byte(name)) != nil {
			return keyExistsErr
		}
		return bucket.Put([]byte(name), []byte(value))
	})
}

func addBucket(path []string, tx *bbolt.Tx) (*bbolt.Bucket, error) {
	if len(path) == 0 {
		return nil, invalidPathErr
	}
	if len(path) == 1 {
		return tx.Bucket([]byte(path[0])), nil
	}
	bucket, err := getParentBucket(path, tx)
	if err != nil {
		return nil, err
	}
	return bucket.CreateBucket([]byte(path[len(path)-1]))
}

func DeleteBucket(path []string) error {
	name := []byte(path[len(path)-1])
	if path == nil {
		return invalidPathErr
	}
	return db.Update(func(tx *bbolt.Tx) error {
		parent, err := getParentBucket(path, tx)
		if err != nil {
			return err
		}
		if parent == nil {
			return tx.DeleteBucket(name)
		}
		return parent.Delete(name)
	})
}

func EmptyBucket(path []string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := getBucket(path, tx)
		if err != nil {
			return err
		}
		if err := bucket.ForEach(func(k, v []byte) error {
			if v == nil {
				return bucket.DeleteBucket(k)
			}
			return bucket.Delete(k)
		}); err != nil {
			return err
		}
		return nil
	})
}

func RenameItem(path []string, name string, isBucket bool) error {
	if isBucket {
		return RenameBucket(path, name)
	}
	return RenameKey(path, name)
}

func RenameBucket(path []string, newName string) error {
	currentName := path[len(path)-1]
	newBucket := &bbolt.Bucket{}
	oldBucket := &bbolt.Bucket{}
	return db.Update(func(tx *bbolt.Tx) error {
		parent, err := getParentBucket(path, tx)
		if err != nil {
			return err
		}
		if parent == nil {
			newBucket, err = tx.CreateBucket([]byte(newName))
			if err != nil {
				return err
			}
			oldBucket = tx.Bucket([]byte(currentName))
		} else {
			newBucket, err = parent.CreateBucket([]byte(newName))
			if err != nil {
				return err
			}
			oldBucket = tx.Bucket([]byte(currentName))
		}
		err = oldBucket.ForEach(func(k, v []byte) error {
			return newBucket.Put(k, v)
		})
		if err != nil {
			return err
		}
		if parent == nil {
			return tx.DeleteBucket([]byte(currentName))
		}
		return parent.DeleteBucket([]byte(currentName))
	})
}

func RenameKey(path []string, newName string) error {
	currentName := path[len(path)-1]
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := getParentBucket(path, tx)
		if err != nil {
			return err
		}
		existing := bucket.Get([]byte(newName))
		if existing != nil {
			return keyExistsErr
		}
		key := bucket.Get([]byte(currentName))
		if key == nil {
			return invalidPathErr
		}
		if err := bucket.Put([]byte(newName), key); err != nil {
			return err
		}
		return bucket.Delete([]byte(currentName))
	})
}

func DeleteItem(n *TreeNode) error {
	if n.IsBucket {
		return DeleteBucket(n.Path)
	}
	return DeleteKey(n.Path)
}

func DeleteKey(path []string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := getParentBucket(path, tx)
		if err != nil {
			return err
		}
		return bucket.Delete([]byte(path[len(path)-1]))
	})
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
		return &bbolt.Bucket{}, invalidPathErr
	}
	for _, p := range path[1:] {
		bucket = bucket.Bucket([]byte(p))
	}
	if bucket == nil {
		return &bbolt.Bucket{}, invalidPathErr
	}
	return bucket, nil
}
