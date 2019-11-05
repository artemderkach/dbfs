package rest

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mind-rot/dbfs/store"
	"github.com/pkg/errors"
)

type Rest struct {
	Store *store.Store
}

// func stripPrefix(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		r.URL.Path = strings.TrimPrefix(r.URL.Path, BasePath)
// 		next.ServeHTTP(w, r)
// 	})
// }

func (rest *Rest) Router() *mux.Router {
	router := mux.NewRouter()
	// subrouter := router.PathPrefix(BasePath).Subrouter()

	// subrouter.Use(stripPrefix)
	router.PathPrefix("").HandlerFunc(rest.view).Methods("GET")
	router.PathPrefix("").HandlerFunc(rest.put).Methods("POST")
	router.PathPrefix("").HandlerFunc(rest.delete).Methods("DELETE")

	return router
}

func sendErr(w http.ResponseWriter, err error) {
	log.Println(err)
	w.Write([]byte(err.Error()))
}

// splitPath will split input string by "/"
// olso it will filter out redundant chars
func splitPath(path string) []string {
	keys := strings.Split(path, "/")
	filterdKeys := make([]string, 0)
	// when splitting string, there could appear some garbage chars like "/" or ""
	for _, key := range keys {
		if key == "" || key == "/" {
			continue
		}
		filterdKeys = append(filterdKeys, key)
	}
	return filterdKeys
}

// view return the current state of database in form of tree view
func (rest *Rest) view(w http.ResponseWriter, r *http.Request) {
	keys := splitPath(r.URL.Path)
	b, err := rest.Store.Get("public", keys)
	if err != nil {
		sendErr(w, errors.Wrap(err, "error retrieving view data from database"))
		return
	}
	if _, err = w.Write(b); err != nil {
		log.Println(err)
	}
}

// put creates new record in database. Returns state of database after write
// Take "multipart/form-data" request with "file" key
func (rest *Rest) put(w http.ResponseWriter, r *http.Request) {
	keys := splitPath(r.URL.Path)

	err := rest.Store.Put("public", keys, r.Body)
	if err != nil {
		sendErr(w, errors.Wrap(err, "error writing file to storage"))
		return
	}

	b, err := rest.Store.Get("public", nil)
	if err != nil {
		sendErr(w, errors.Wrap(err, "error retrieving view data from database"))
		return
	}
	if _, err = w.Write(b); err != nil {
		log.Println(err)
	}
}

// delete removes value from database
// returnes current state of tree (GET / route)
func (rest *Rest) delete(w http.ResponseWriter, r *http.Request) {
	keys := splitPath(r.URL.Path)

	err := rest.Store.Delete("public", keys)
	if err != nil {
		sendErr(w, errors.Wrap(err, "error deleting file from database"))
		return
	}

	b, err := rest.Store.Get("public", nil)
	if err != nil {
		sendErr(w, errors.Wrap(err, "error retrieving view data from database"))
		return
	}

	if _, err = w.Write(b); err != nil {
		log.Println(err)
	}
}

func (rest *Rest) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from DBFS"))
}
