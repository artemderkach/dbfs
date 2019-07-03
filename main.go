package main

import (
	"fmt"
	"net/http"

	"github.com/mind-rot/dbfs/rest"
	"github.com/pkg/errors"
)

func main() {
	fmt.Println("running dbfs on :8080")
	r := rest.Rest{}

	err := http.ListenAndServe(":8080", r.Router())
	if err != nil {
		panic(errors.Wrap(err, "error starting dbfs server"))
	}
}

// func initDB(w http.ResponseWriter, r *http.Request) {
// 	db, err := bolt.Open("db.bolt", 0600, nil)
// 	if err != nil {
// 		w.Write([]byte(err.Error()))
// 		return
// 	}
// 	err = db.Update(func(tx *bolt.Tx) error {
// 		b, err := tx.CreateBucket([]byte("files"))
// 		if err != nil {
// 			return err
// 		}
// 		b.Put([]byte("test"), []byte("42"))
// 		return nil
// 	})
// 	if err != nil {
// 		w.Write([]byte(err.Error()))
// 		return
// 	}
// 	w.Write([]byte("some"))
// }

// func dbHandler(w http.ResponseWriter, r *http.Request) {
// 	db, err := bolt.Open("db.bolt", 0600, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	result := ""
// 	err = db.View(func(tx *bolt.Tx) error {
// 		fmt.Println("___OMG___")
// 		c := tx.Cursor()
// 		for k, v := c.First(); k != nil; k, v = c.Next() {
// 			result += fmt.Sprintf("key=%s, value=%s\n", k, v)
// 			fmt.Printf("key=%s, value=%s\n", k, v)
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer db.Close()

// 	w.Write([]byte(result))
// }
