package main

import (
	"fmt"
	"net/http"

	"github.com/mind-rot/dbfs/rest"
	"github.com/mind-rot/dbfs/store"
	"github.com/pkg/errors"
)

func main() {
	fmt.Println("running dbfs on :8080")
	r := getRest()

	err := http.ListenAndServe(":8080", r.Router())
	if err != nil {
		panic(errors.Wrap(err, "error starting dbfs server"))
	}
}

func getRest() *rest.Rest {
	s := &store.Store{
		Path:       "/tmp/mystore.bolt",
		Collection: "files",
	}
	r := &rest.Rest{
		Store: s,
	}

	return r
}
