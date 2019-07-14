package rest

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/mind-rot/dbfs/store"
	"github.com/pkg/errors"
)

type Rest struct {
	Store    *store.Store
	APP_PASS string
}

func (rest *Rest) Router() *mux.Router {
	router := mux.NewRouter()
	router.Use(rest.permissionCheck)
	router.HandleFunc("/", rest.home).Methods("GET")
	router.HandleFunc("/{collection}/view", rest.view).Methods("GET")
	router.PathPrefix("/{collection}/put").HandlerFunc(rest.put).Methods("POST")
	router.PathPrefix("/{collection}/download").HandlerFunc(rest.download).Methods("GET")
	router.PathPrefix("/{collection}/delete").HandlerFunc(rest.delete).Methods("DELETE")

	return router
}

// delete removes value from database
// returnes current state of database (/view route)
func (rest *Rest) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reg, err := regexp.Compile("^.+delete")
	if err != nil {
		err = errors.Wrap(err, "error parsing regexp")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
	path := reg.ReplaceAllString(r.URL.Path, "")

	collection := vars["collection"]
	err = validateCollection(collection)
	if err != nil {
		err = errors.Wrap(err, "invalid collection parameter")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	err = rest.Store.Delete(collection, path)
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
	reg, err := regexp.Compile("^.+download")
	if err != nil {
		err = errors.Wrap(err, "error parsing regexp")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
	path := reg.ReplaceAllString(r.URL.Path, "")

	collection := vars["collection"]
	err = validateCollection(collection)
	if err != nil {
		err = errors.Wrap(err, "invalid collection parameter")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := rest.Store.Get(collection, path)
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
	reg, err := regexp.Compile("^.+put")
	if err != nil {
		err = errors.Wrap(err, "error compiling regexp")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
	path := reg.ReplaceAllString(r.URL.Path, "")

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

	err = rest.Store.Put(collection, path+"/"+header.Filename, file)
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

func (rest *Rest) permissionCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		collection := vars["collection"]

		if collection == "private" && len(rest.APP_PASS) != 0 {
			pass := r.Header.Get("Custom-Auth")
			hashedPass := sha256.Sum256([]byte(pass))

			if fmt.Sprintf("%x", hashedPass) != rest.APP_PASS {
				fmt.Println(rest.APP_PASS)
				fmt.Println(fmt.Sprintf("%x", hashedPass))
				w.Write([]byte("permission denied"))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
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
