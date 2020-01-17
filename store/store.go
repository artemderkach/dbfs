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
			return errors.Errorf("bucket \"%s\" not exists", collection)
		}

		// handle case for top level bucket
		if len(keys) == 0 {
			result, err = view(tx, b, "")
			return errors.Wrap(err, "error viewing node")
		}

		// skip last element, it will be checked after loop
		for i := 0; i < len(keys)-1; i += 1 {
			b = b.Bucket([]byte(keys[i]))
			if b == nil {
				return errors.Errorf("bucket \"%s\" not found", keys[i])
			}
		}

		// check the last element (could be file or bucket)
		lastElem := keys[len(keys)-1]

		// if the last element is bucket
		if b.Bucket([]byte(lastElem)) != nil {
			result, err = view(tx, b.Bucket([]byte(lastElem)), "")
			return errors.Wrap(err, "error creating view")
		}
		// if the last element is file
		v := b.Get([]byte(lastElem))
		if v != nil {
			result = make([]byte, len(v))
			copy(result, v)
			return nil
		}

		return errors.Errorf("bucket \"%s\" not found", lastElem)
	})

	return result, errors.Wrap(err, "error getting elements from bucket")
}

// Put creates bucket for each "key" passed in params, except for the last one
// last one is used as file name
// if only one "key" passed, file will be created in root directory
// this function cannot create bucket without file, so reader is required
func (store *Store) Put(collection string, keys []string, file io.Reader) error {
	// protect reserved name
	if len(keys) > 0 && keys[0] == "shared" {
		return errors.New("'shared' name is reserved")
	}

	db, err := store.open()
	if err != nil {
		return errors.New("error opening database")
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return errors.Errorf("bucket \"%s\" not exists", collection)
		}

		// skip last element, it will be checked after loop
		// last element should be the file, other ones - folders
		for i := 0; i < len(keys)-1; i += 1 {
			// not possible to create bucket, if this name is used for file
			if b.Get([]byte(keys[i])) != nil {
				return errors.Errorf("name \"%s\" already used", keys[i])
			}
			b, err = b.CreateBucketIfNotExists([]byte(keys[i]))
			if err != nil {
				return errors.Wrap(err, "error opening bucket")
			}
		}

		lastElem := keys[len(keys)-1]

		// last element should not exists as bucket
		if b.Bucket([]byte(lastElem)) != nil {
			return errors.Errorf("name \"%s\" already used", lastElem)
		}

		f, err := ioutil.ReadAll(file)
		if err != nil {
			return errors.Wrap(err, "error reading file from reader")
		}
		b.Put([]byte(lastElem), f)

		return nil
	})

	return errors.Wrap(err, "error updating database")
}

// Delete removes element from database
// in case it is a bucket, remove this bucket and all elements under this bucket
// bucket removes recursively
func (store *Store) Delete(collection string, keys []string) error {
	// protect reserved name
	if len(keys) > 0 && keys[0] == "shared" {
		return errors.New("'shared' name is reserved")
	}

	if len(keys) == 0 {
		return errors.New("empty delete request")
	}
	db, err := store.open()
	if err != nil {
		return errors.Wrap(err, "error opening database")
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return errors.Errorf("bucket \"%s\" not exists", collection)
		}
		// skip last element, it will be checked after loop
		for i := 0; i < len(keys)-1; i += 1 {
			b = b.Bucket([]byte(keys[i]))
			if b == nil {
				return errors.Errorf("bucket \"%s\" not found", keys[i])
			}

		}
		lastElem := keys[len(keys)-1]

		if b.Get([]byte(lastElem)) != nil {
			return b.Delete([]byte(lastElem))
		}

		err := b.DeleteBucket([]byte(lastElem))
		if err != nil && err != bolt.ErrIncompatibleValue {
			return err
		}

		return nil
	})

	return errors.Wrap(err, "error updating database")
}

// Create creates bucket for new user
func (store *Store) Create(collection string) error {
	db, err := store.open()
	if err != nil {
		return errors.Wrap(err, "error openiong database")
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte(collection))
		if err != nil {
			return errors.Wrap(err, "error creating bucket")
		}
		return nil
	})

	return errors.Wrap(err, "error updating database")
}

// Copy element from given collection to target collection
func (store *Store) Share(collection string, from []string, target string) error {
	db, err := store.open()
	if err != nil {
		return errors.Wrap(err, "error openiong database")
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(target))

		return errors.Wrap(err, "error creating new bucket")
	})

	if err != nil {
		return errors.Wrap(err, "error sharing bucket")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		fromBucket := tx.Bucket([]byte(collection))
		if err != nil {
			return errors.Wrap(err, "error getting existing bucket")
		}

		shared, err := fromBucket.CreateBucketIfNotExists([]byte("shared"))
		if err != nil {
			return err
		}

		// create reference to shared elements
		err = shared.Put([]byte(target), []byte(""))
		if err != nil {
			return errors.Wrap(err, "error creating shared bucket")
		}

		targetBucket := tx.Bucket([]byte(target))
		if len(from) == 0 {
			return copyChilds(fromBucket, targetBucket)
		}

		targetName := ""
		for _, bucketName := range from {
			targetName = bucketName
			fromBucket = fromBucket.Bucket([]byte(bucketName))
			if fromBucket == nil {
				return errors.Errorf("bucket \"%s\" not exists", bucketName)
			}
		}

		return copyBucket(fromBucket, targetBucket, targetName)
	})

	return errors.Wrap(err, "error updating database")
}

// view
func view(tx *bolt.Tx, b *bolt.Bucket, indent string) ([]byte, error) {
	result := nestedView(b, indent)

	sharedResult, err := sharedView(tx, b, indent)
	if sharedResult != "" {
		result += sharedResult
	}

	return []byte(result), errors.Wrap(err, "error creating view")
}

// nestedView runs throug every bucket recursively
// in the end we'll get tree view of data
func nestedView(b *bolt.Bucket, indent string) string {
	var view string

	c := b.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		// shared element will be processed after loop
		if string(k) == "shared" {
			continue
		}

		view += indent + string(k) + "\n"

		// nestedBucket will be nil if "k" is'n bucket
		nestedBucket := b.Bucket(k)
		if nestedBucket != nil {
			view += nestedView(nestedBucket, indent+"  ")
		}
	}

	return view
}

// sharedView takes names from 'shared' node, than use this names to search for its view on top-level
func sharedView(tx *bolt.Tx, b *bolt.Bucket, indent string) (string, error) {
	shared := b.Bucket([]byte("shared"))
	if shared == nil {
		return "", nil
	}

	result := ""

	result += indent + "shared" + "\n"
	c := shared.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		result += indent + "  " + string(k) + "\n"
		// nestedBucket will be nil if "k" is'n bucket
		nestedBucket := tx.Bucket(k)
		if nestedBucket != nil {
			result += nestedView(nestedBucket, indent+"  "+"  ")
		}
	}

	return result, nil
}

// interface is needed due to possible input as "*bolt.Bucket" or "*bolt.Tx"
type bucket interface {
	Bucket([]byte) *bolt.Bucket
	CreateBucket([]byte) (*bolt.Bucket, error)
	CreateBucketIfNotExists([]byte) (*bolt.Bucket, error)
	Put([]byte, []byte) error
	ForEach(func([]byte, []byte) error) error
}

// copyBucket copies "source" bucket and all his childs inside "target"
func copyBucket(source bucket, target bucket, name string) error {
	err := source.ForEach(func(k, v []byte) error {
		newTarget, err := target.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return err
		}

		nestedBucket := source.Bucket(k)
		if nestedBucket == nil {
			return newTarget.Put(k, v)
		}

		return copyBucket(nestedBucket, newTarget, string(k))
	})

	return errors.Wrap(err, "error while recursive copying")
}

// copyChilds copies "source" bucket childs to "target"
// this function is needed for root copying
func copyChilds(source *bolt.Bucket, target *bolt.Bucket) error {
	err := source.ForEach(func(k, v []byte) error {
		nestedBucket := source.Bucket(k)
		if nestedBucket == nil {
			return target.Put(k, v)
		}

		return copyBucket(nestedBucket, target, string(k))

	})

	return errors.Wrap(err, "error while copying childs")
}
