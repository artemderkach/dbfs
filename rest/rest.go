package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mind-rot/dbfs/store"
	"github.com/pkg/errors"
)

type Rest struct {
	Store *store.Store
}

func (rest *Rest) Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", rest.home)
	router.HandleFunc("/view", rest.view)
	router.HandleFunc("/put", rest.put)

	return router
}

func (rest *Rest) view(w http.ResponseWriter, r *http.Request) {
	b, err :=rest.Store.View()
	if err != nil {
		err = errors.Wrap(err, "error retrieving view data from database")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

func (rest *Rest) put(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		err = errors.Wrap(err, "error parsing FormFile")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	err = rest.Store.Put(header.Filename, file)
	if err != nil {
		err = errors.Wrap(err, "error writing file to storage")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
	return
	}
	w.Write([]byte("Written successfully"))
}

func (rest *Rest) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from DBFS"))
}
