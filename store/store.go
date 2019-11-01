package store

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type Store struct {
	Path string
	DB   *bolt.DB
}

// Open
func (store *Store) open() (*bolt.DB, error) {
	db, err := bolt.Open(store.Path, 0600, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error opening database")
	}

	return db, nil
}

// Drop deletes database despite it's not empty
func (store *Store) Drop() error {
	err := os.Remove(store.Path)
	if err != nil {
		return errors.Wrap(err, "error dropping database")
	}

	return nil
}

// Get nestedly searches for keys array, one by one, inside given collection
// function could have 3 outcomes
// 1. In case of last element is bucket, will return tree view of it
// 2. In case of last element is file, will return file content
// 3. Error either from invalid key or smth other
func (store *Store) Get(collection string, keys []string) ([]byte, error) {
	db, err := store.open()
	if err != nil {
		return nil, errors.Wrap(err, "error opening database")
	}
	defer db.Close()

	var result []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return errors.Wrap(err, "bucket not exists")
		}

		// handle case for top level bucket
		if len(keys) == 0 {
			result = []byte(nestedView(b, ""))
			return nil
		}

		for i, key := range keys {
			// skip last element, it will be checked after loop
			if i == len(keys)-1 {
				continue
			}
			b = b.Bucket([]byte(key))
			if b == nil {
				return errors.Wrap(err, "bucket \""+key+"\" not exists")
			}
		}

		// check the last element (could be file or bucket)
		lastElem := keys[len(keys)-1]

		// if the last element is bucket
		if b.Bucket([]byte(lastElem)) != nil {
			result = []byte(nestedView(b.Bucket([]byte(lastElem)), ""))
			return nil
		}
		// if the last element is file
		v := b.Get([]byte(lastElem))
		if v != nil {
			result = make([]byte, len(v))
			copy(result, v)
			return nil
		}

		return errors.New("bucket \"" + lastElem + "\" not exists")
	})
	if err != nil {
		return nil, errors.Wrap(err, "error getting elements from bucket")
	}

	return result, err
}

// Put creates bucket for each "key" passed in params, except for the last one
// last one is used as file name
// if only one "key" passed, file will be created in root directory
// this function cannot create bucket without file, so reader is required
func (store *Store) Put(collection string, keys []string, file io.Reader) error {
	db, err := store.open()
	if err != nil {
		return errors.Wrap(err, "error opening database")
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			return errors.Wrap(err, "error opening bucket")
		}

		for i, key := range keys {
			// last element should be the file, other ones - folders
			if i+1 != len(keys) {
				// not possible to create bucket, if this name is used for file
				if b.Get([]byte(key)) != nil {
					return errors.Errorf("name \"%s\" already used", key)
				}
				b, err = b.CreateBucketIfNotExists([]byte(key))
				if err != nil {
					return errors.Wrap(err, "error opening bucket")
				}

				continue
			}

			// last element should not exists as bucket
			if b.Bucket([]byte(key)) != nil {
				return errors.Errorf("name \"%s\" already used", key)
			}

			f, err := ioutil.ReadAll(file)
			if err != nil {
				return errors.Wrap(err, "error reading file from reader")
			}
			b.Put([]byte(key), f)
		}

		return nil
	})

	if err != nil {
		return errors.Wrap(err, "error updating database")
	}

	return nil
}

// Delete removes element from database
// in case it is a bucket, remove this bucket and all elements under this bucket
// bucket removes recursively
func (store *Store) Delete(collection string, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	db, err := store.open()
	if err != nil {
		return errors.Wrap(err, "error opening database")
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return errors.Wrap(err, "bucket not exists")
		}
		for i, key := range keys {
			// skip last element, it will be checked after loop
			if i == len(keys)-1 {
				continue
			}
			b = b.Bucket([]byte(key))
			if b == nil {
				return errors.Wrap(err, "bucket \""+key+"\" not exists")
			}
		}
		lastElem := keys[len(keys)-1]

		err := b.Delete([]byte(lastElem))
		if err != bolt.ErrIncompatibleValue {
			return err
		}
		err = b.DeleteBucket([]byte(lastElem))
		return err
	})
	if err != nil {
		return errors.Wrap(err, "error updating database")
	}

	return nil
}

// nestedView runs throug every bucket recursively
// in the end we'll get tree view of data
func nestedView(b *bolt.Bucket, indent string) string {
	var view string

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
