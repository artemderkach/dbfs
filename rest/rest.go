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
	router.HandleFunc("/", rest.home).Methods("GET")
	router.HandleFunc("/view", rest.view).Methods("GET")
	router.HandleFunc("/put", rest.put).Methods("POST")
	router.HandleFunc("/download/{filename}", rest.download).Methods("GET")

	return router
}

// dowload returns the value for filename
func (rest *Rest) download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	filename := vars["filename"]

	b, err := rest.Store.Get(filename)
	if err != nil {
		err = errors.Wrap(err, "error retrieving file data from database")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

// view return the current state of database
func (rest *Rest) view(w http.ResponseWriter, r *http.Request) {
	b, err := rest.Store.View()
	if err != nil {
		err = errors.Wrap(err, "error retrieving view data from database")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

// put creates new record in database. Returns state of database after write
// Take "multipart/form-data" request with "file" key
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

	b, err := rest.Store.View()
	if err != nil {
		err = errors.Wrap(err, "error retrieving view data from database")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

func (rest *Rest) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from DBFS"))
}
