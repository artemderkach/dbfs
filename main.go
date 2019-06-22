package main

import (
	"fmt"
	"net/http"

	"github.com/boltdb/bolt"
)

func main() {
	fmt.Println("running dbfs on :8080")
	http.HandleFunc("/db", db)
	http.HandleFunc("/", home)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from DBFS"))
}

func db(w http.ResponseWriter, r *http.Request) {
	db, err := bolt.Open("db.bolt", 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	w.Write([]byte("Hello from boltdb"))
}
