package main

import (
	"fmt"
	"net/http"

	"github.com/boltdb/bolt"
)

func main() {
	fmt.Println("running dbfs on :8080")
	http.HandleFunc("/", home)
	http.HandleFunc("/initdb", initDB)
	http.HandleFunc("/db", dbHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
    	panic(err)
	}
}

func router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", home)

	return mux
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from DBFS"))
}

func initDB(w http.ResponseWriter, r *http.Request) {
	db, err := bolt.Open("db.bolt", 0600, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("files"))
		if err != nil {
			return err
		}
		b.Put([]byte("test"), []byte("42"))
		return nil
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("some"))
}

func dbHandler(w http.ResponseWriter, r *http.Request) {
	db, err := bolt.Open("db.bolt", 0600, nil)
	if err != nil {
		panic(err)
	}
	result := ""
	err = db.View(func(tx *bolt.Tx) error {
		fmt.Println("___OMG___")
		c := tx.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			result += fmt.Sprintf("key=%s, value=%s\n", k, v)
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	w.Write([]byte(result))
}
