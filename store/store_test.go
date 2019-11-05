package store

import (
	"strings"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var DB_PATH = "/tmp/myTestDB.bolt"

// view of db
// 1
//   2
// The Ring
// a
//   b
//     c
// Hello there
func initStore() (*Store, error) {
	db, err := bolt.Open(DB_PATH, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	s := &Store{
		Path: DB_PATH,
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("public"))
		if err != nil {
			return err
		}

		b, err = b.CreateBucketIfNotExists([]byte("1"))
		if err != nil {
			return err
		}

		err = b.Put([]byte("2"), []byte("0"))
		if err != nil {
			return err
		}

		b, err = tx.CreateBucketIfNotExists([]byte("public"))
		if err != nil {
			return err
		}

		err = b.Put([]byte("The Ring"), []byte("My precious"))
		if err != nil {
			return err
		}

		// nested folders
		for _, v := range []string{"a", "b", "c"} {
			b, err = b.CreateBucketIfNotExists([]byte(v))
			if err != nil {
				return err
			}
		}

		err = b.Put([]byte("Hello there"), []byte("General Kenobi"))
		if err != nil {
			return err
		}
		return nil
	})
	return s, err
}

func TestGet(t *testing.T) {
	tt := []struct {
		Collection string
		Keys       []string
		Result     string
		Error      error
	}{
		{
			"public",
			[]string{},
			"1\n  2\nThe Ring\na\n  b\n    c\n      Hello there\n",
			nil,
		},
		{
			"public",
			[]string{"1"},
			"2\n",
			nil,
		},
		{
			"public",
			[]string{"The Ring"},
			"My precious",
			nil,
		},
		{
			"public",
			[]string{"a", "b", "c"},
			"Hello there\n",
			nil,
		},
		{
			"public",
			[]string{"a", "b", "c", "Hello there"},
			"General Kenobi",
			nil,
		},
		{
			"public",
			[]string{"invalid bucket"},
			"General Kenobi",
			errors.New("error getting elements from bucket: bucket \"invalid bucket\" not exists"),
		},
		{
			"public",
			[]string{"a", "b", "invalid file"},
			"General Kenobi",
			errors.New("error getting elements from bucket: bucket \"invalid bucket\" not exists"),
		},
	}

	s, err := initStore()
	require.Nil(t, err)
	defer s.Drop()
	require.Nil(t, err)

	for _, test := range tt {
		result, err := s.Get(test.Collection, test.Keys)
		if test.Error != nil {
			if err == nil {
				t.Error("error should not be nil")
				continue
			}
			assert.Equal(t, test.Error.Error(), err.Error())
			continue
		}
		require.Nil(t, err)

		assert.Equal(t, test.Result, string(result))
	}
}

func TestPut(t *testing.T) {
	tt := []struct {
		Collection  string
		Keys        []string
		FileContent string
		Error       error
	}{
		{
			"public",
			[]string{"Ok, that was epic!"},
			"Next Meme",
			nil,
		},
		{
			"public",
			[]string{"prequel", "It's over Anakin!"},
			"I have the highground!",
			nil,
		},
		{
			"public",
			[]string{"prequel", "It's a trap!"},
			"Acbar",
			nil,
		},
		{
			"public",
			[]string{"prequel"},
			"memes",
			errors.New("error updating database: name \"prequel\" already used"),
		},
		{
			"public",
			[]string{"The Ring", "omg this should fail"},
			"memes",
			errors.New("error updating database: name \"The Ring\" already used"),
		},
	}

	s, err := initStore()
	require.Nil(t, err)
	defer s.Drop()

	for _, test := range tt {
		file := strings.NewReader(test.FileContent)

		err := s.Put(test.Collection, test.Keys, file)
		if test.Error != nil {
			if err == nil {
				t.Error("error should not be nil")
				continue
			}
			assert.Equal(t, test.Error.Error(), err.Error())
			continue
		}
		require.Nil(t, err)
	}

	result, err := s.Get("public", []string{})
	require.Nil(t, err)

	expected := "1\n" +
		"  2\n" +
		"Ok, that was epic!\n" +
		"The Ring\n" +
		"a\n" +
		"  b\n" +
		"    c\n" +
		"      Hello there\n" +
		"prequel\n" +
		"  It's a trap!\n" +
		"  It's over Anakin!\n"

	assert.Equal(t, expected, string(result))
}

func TestDelete(t *testing.T) {
	tt := []struct {
		Collection string
		Keys       []string
		Error      error
	}{
		{
			"public",
			[]string{"1", "2"},
			nil,
		},
		{
			"public",
			[]string{"a"},
			nil,
		},
		{
			"public",
			[]string{"invalid key"},
			errors.New("error"),
		},
	}

	s, err := initStore()
	require.Nil(t, err)
	defer s.Drop()

	for _, test := range tt {
		err := s.Delete(test.Collection, test.Keys)
		require.Nil(t, err)

		if test.Error != nil {
			if err == nil {
				t.Error("error should no be nil")
				continue
			}
			assert.Equal(t, test.Error.Error(), err.Error())
		}

	}

	result, err := s.Get("public", []string{})
	require.Nil(t, err)

	expected := "1\n" +
		"The Ring\n"

	assert.Equal(t, expected, string(result))
}
