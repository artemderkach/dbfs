package store

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type Store struct {
	Path string
	DB   *bolt.DB
}

func (store *Store) Open() error {
	db, err := bolt.Open(store.Path, 0600, nil)
	if err != nil {
		return errors.Wrap(err, "erorr opening database")
	}

	store.DB = db

	return nil
}

func (store *Store) Drop() error {
	store.DB.Close()
	err := os.Remove(store.Path)
	if err != nil {
		return errors.Wrap(err, "error dropping database")
	}

	return nil
}

func (store *Store) View(collection string) ([]byte, error) {
	err := store.Open()
	if err != nil {
		return nil, errors.Wrap(err, "error opening database")
	}
	defer store.DB.Close()

	var result []byte
	err = store.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return errors.Errorf("bucket %s not exists", collection)
		}
		view := nestedView(b, "")
		result = []byte(view)

		return nil
	})

	return result, err
}

func (store *Store) Put(collection, path string, file io.Reader) error {
	elements := strings.Split(path, "/")
	filterdElements := make([]string, 0)
	for _, element := range elements {
		if element == "" || element == "/" {
			continue
		}
		filterdElements = append(filterdElements, element)
	}

	err := store.Open()
	if err != nil {
		return errors.Wrap(err, "error opening database")
	}
	defer store.DB.Close()

	err = store.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			return errors.Wrap(err, "error opening bucket")
		}

		for i, filterdElement := range filterdElements {
			// last element should be the file, other ones - folders
			if i+1 != len(filterdElements) {
				b, err = b.CreateBucketIfNotExists([]byte(filterdElement))
				if err != nil {
					return errors.Wrap(err, "error opening bucket")
				}

				continue
			}

			f, err := ioutil.ReadAll(file)
			if err != nil {
				return errors.Wrap(err, "error reading file from reader")
			}
			b.Put([]byte(filterdElement), f)
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "error updating database")
	}

	return nil
}

// Get nestedly searches for keys array, one by one, inside given collection
// function could have 3 outcomes
// 1. In case of last element is bucket, will return tree view of it
// 2. In case of last element is file, will return file content
// 3. Error either from invalid key or smth other
func (store *Store) Get(collection string, keys []string) ([]byte, error) {
	fmt.Println("======>", collection, keys)
	err := store.Open()
	if err != nil {
		return nil, errors.Wrap(err, "error opening database")
	}
	defer store.DB.Close()

	var result []byte
	err = store.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			return errors.Wrap(err, "error opening bucket")
		}
		for i, key := range keys {
			// skip last element, it will be checked after loop
			if i == len(keys)-1 {
				continue
			}
			b = b.Bucket([]byte(key))
			if b == nil {
				return errors.Wrap(err, "bucket "+key+" not exists")
			}
		}

		// check the last element (could be file or bucket)
		lastElem := ""
		if len(keys) != 0 {
			lastElem = keys[len(keys)-1]
		}
		lastBucket := b.Bucket([]byte(lastElem))
		if lastBucket == nil {
			v := b.Get([]byte(lastElem))
			result = make([]byte, len(v))
			copy(result, v)
		} else {
			result = []byte(nestedView(b, "  "))
		}

		return nil
	})

	return result, err
}

func (store *Store) Delete(collection, path string) error {
	elements := strings.Split(path, "/")
	filterdElements := make([]string, 0)
	for _, element := range elements {
		if element == "" || element == "/" {
			continue
		}
		filterdElements = append(filterdElements, element)
	}

	err := store.Open()
	if err != nil {
		return errors.Wrap(err, "error opening database")
	}
	defer store.DB.Close()

	err = store.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			return errors.Wrap(err, "error opening bucket")
		}
		for i, filterdElement := range filterdElements {

			// last element should be the file, other ones - folders
			if i+1 != len(filterdElements) {
				b, err = b.CreateBucketIfNotExists([]byte(filterdElement))
				if err != nil {
					return errors.Wrap(err, "error opening bucket")
				}

				continue
			}

			err := b.Delete([]byte(filterdElement))
			if err != bolt.ErrIncompatibleValue {
				return err
			}
			err = b.DeleteBucket([]byte(filterdElement))
			return err
		}

		return nil
	})

	return err
}

// nestedView runs throug every bucket recursively
// in the end we'll get tree view of data
func nestedView(b *bolt.Bucket, indent string) (view string) {
	c := b.Cursor()

	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		view += indent + string(k) + "\n"

		nestedBucket := b.Bucket(k)
		if nestedBucket != nil {
			view += nestedView(nestedBucket, indent+"  ")
		}
	}

	return view
}

// nestedBucket searches for nested bukcets with input string name
// func nestedBucket(b *bolt.Bucket, []string) (*bolt.Bucket, error) {
// }
