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
}

func (store *Store) Drop() error {
	err := os.Remove(store.Path)
	if err != nil {
		return errors.Wrap(err, "error dropping database")
	}
	return nil
}

func (store *Store) View(collection string) (result []byte, err error) {
	db, err := bolt.Open(store.Path, 0600, nil)
	if err != nil {
		return []byte(""), errors.Wrap(err, "error opening database")
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
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
	db, err := bolt.Open(store.Path, 0600, nil)
	if err != nil {
		return errors.Wrap(err, "error opening database")
	}
	defer db.Close()

	elements := strings.Split(path, "/")
	filterdElements := make([]string, 0)
	for _, element := range elements {
		if element == "" || element == "/" {
			continue
		}
		filterdElements = append(filterdElements, element)
	}

	err = db.Update(func(tx *bolt.Tx) error {
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

func (store *Store) Get(collection, path string) (result []byte, err error) {
	db, err := bolt.Open(store.Path, 0600, nil)
	if err != nil {
		return []byte(""), errors.Wrap(err, "error opening database")
	}
	defer db.Close()

	elements := strings.Split(path, "/")
	filterdElements := make([]string, 0)
	for _, element := range elements {
		if element == "" || element == "/" {
			continue
		}
		filterdElements = append(filterdElements, element)
	}

	err = db.Update(func(tx *bolt.Tx) error {
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

			v := b.Get([]byte(filterdElement))
			result = make([]byte, len(v))
			copy(result, v)
		}

		return nil
	})

	return result, err
}

func (store *Store) Delete(collection, path string) (err error) {
	db, err := bolt.Open(store.Path, 0600, nil)
	if err != nil {
		return errors.Wrap(err, "error opening database")
	}
	defer db.Close()

	elements := strings.Split(path, "/")
	filterdElements := make([]string, 0)
	for _, element := range elements {
		if element == "" || element == "/" {
			continue
		}
		filterdElements = append(filterdElements, element)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			return errors.Wrap(err, "error opening bucket")
		}
		fmt.Println("=========", filterdElements)
		for i, filterdElement := range filterdElements {

			// last element should be the file, other ones - folders
			if i+1 != len(filterdElements) {
				fmt.Println("========")
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
