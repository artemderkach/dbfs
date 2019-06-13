package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("running dbfs on :8080")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from DBFS"))
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
