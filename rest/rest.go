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
	router.HandleFunc("/{collection}/view", rest.view).Methods("GET")
	router.HandleFunc("/{collection}/put", rest.put).Methods("POST")
	router.HandleFunc("/{collection}/download/{filename}", rest.download).Methods("GET")
	router.HandleFunc("/{collection}/delete/{filename}", rest.delete).Methods("DELETE")

	return router
}

// delete removes value from database
// returnes current state of database (/view route)
func (rest *Rest) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	filename := vars["filename"]
	collection := vars["collection"]
	err := validateCollection(collection)
	if err != nil {
		err = errors.Wrap(err, "invalid collection parameter")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	err = rest.Store.Delete(collection, filename)
	if err != nil {
		err = errors.Wrap(err, "error deleting file from database")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := rest.Store.View(collection)
	if err != nil {
		err = errors.Wrap(err, "error retrieving view data from database")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

// dowload returns the value for filename
func (rest *Rest) download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	filename := vars["filename"]
	collection := vars["collection"]
	err := validateCollection(collection)
	if err != nil {
		err = errors.Wrap(err, "invalid collection parameter")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := rest.Store.Get(collection, filename)
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
	vars := mux.Vars(r)

	collection := vars["collection"]
	err := validateCollection(collection)
	if err != nil {
		err = errors.Wrap(err, "invalid collection parameter")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := rest.Store.View(collection)
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
	vars := mux.Vars(r)

	file, header, err := r.FormFile("file")
	if err != nil {
		err = errors.Wrap(err, "error parsing FormFile")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	collection := vars["collection"]
	err = validateCollection(collection)
	if err != nil {
		err = errors.Wrap(err, "invalid collection parameter")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	err = rest.Store.Put(collection, header.Filename, file)
	if err != nil {
		err = errors.Wrap(err, "error writing file to storage")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := rest.Store.View(collection)
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

func validateCollection(collection string) (err error) {
	collections := [2]string{"public", "private"}

	for _, v := range collections {
		if v == collection {
			return nil
		}
	}

	return errors.New("collection should be 'public' or 'pribate'")
}
