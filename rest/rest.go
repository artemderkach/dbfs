package rest

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mind-rot/dbfs/email"
	"github.com/mind-rot/dbfs/store"
	"github.com/pkg/errors"
)

const basePath string = "/db"

type Rest struct {
	Store *store.Store
	Email *email.Email
}

// Router creates router instance with mapped routes
func (rest *Rest) Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", rest.home)
	router.HandleFunc("/register", rest.register)

	// actual db interactions
	subrouter := router.PathPrefix(basePath).Subrouter()
	subrouter.Use(rest.stripPrefix)
	subrouter.PathPrefix("").HandlerFunc(rest.view).Methods("GET")
	subrouter.PathPrefix("").HandlerFunc(rest.put).Methods("POST")
	subrouter.PathPrefix("").HandlerFunc(rest.delete).Methods("DELETE")

	return router
}

func sendErr(w http.ResponseWriter, err error) {
	log.Println(err)
	w.Write([]byte(err.Error()))
}

// splitPath will split input string by "/"
// also it will filter out redundant chars (imagine like "a///b/c", will lead to [a, b, c])
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

// stripPrefix removes basePath from request url
func (rest *Rest) stripPrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, basePath)
		next.ServeHTTP(w, r)
	})
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

// register creates token for user
func (rest *Rest) register(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("register"))
}

func (rest *Rest) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from DBFS"))
}
