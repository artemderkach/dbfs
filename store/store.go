package store

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type Store struct {
	Path       string
	Collection string
}

func (store *Store) Init() error {
	db, err := bolt.Open(store.Path, 0600, nil)
	if err != nil {
		return errors.Wrap(err, "error opening database")
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(store.Collection))
		if err != nil {
			return errors.Wrap(err, "error creating bucket")
		}

		return nil
	})

	return err
}

func (store *Store) Drop() error {
	err := os.Remove(store.Path)
	if err != nil {
		return errors.Wrap(err, "error droping database")
	}
	return nil
}

func (store *Store) View(filename string) (result []byte, err error) {
	db, err := bolt.Open(store.Path, 0600, nil)
	if err != nil {
		return []byte(""), errors.Wrap(err, "error opening database")
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(store.Collection))
		b.Get([]byte(filename))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			result = append(result, k...)
		}
		return nil
	})

	return result, err
}

func (store *Store) Put(filename string, file io.Reader) error {
	db, err := bolt.Open(store.Path, 0600, nil)
	if err != nil {
		return errors.Wrap(err, "error opening database")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(store.Collection))
		f, err := ioutil.ReadAll(file)
		if err != nil {
			return errors.Wrap(err, "error reading file from reader")
		}
		b.Put([]byte(filename), f)

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "error updating database")
	}

	return nil
}
