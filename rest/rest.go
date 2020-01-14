package rest

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mind-rot/dbfs/email"
	"github.com/mind-rot/dbfs/store"
	"github.com/pkg/errors"
)

const (
	basePath   = "/db"
	sharePath  = "/share"
	sharedPath = "/shared"
)

type Rest struct {
	Store *store.Store
	Email email.EmailService
}

// Router creates router instance with mapped routes
func (rest *Rest) Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", rest.home).Methods("GET")
	router.HandleFunc("/register", rest.register).Methods("POST")

	// actual db interactions
	dbSubrouter := router.PathPrefix(basePath).Subrouter()
	dbSubrouter.Use(rest.stripPrefix(basePath))
	dbSubrouter.PathPrefix("").HandlerFunc(rest.view).Methods("GET")
	dbSubrouter.PathPrefix("").HandlerFunc(rest.put).Methods("POST")
	dbSubrouter.PathPrefix("").HandlerFunc(rest.delete).Methods("DELETE")

	// share functionality
	sharedSubrouter := router.PathPrefix(sharedPath).Subrouter()
	sharedSubrouter.Use(rest.stripPrefix(sharedPath))
	sharedSubrouter.PathPrefix("").HandlerFunc(rest.shared).Methods("GET")

	shareSubrouter := router.PathPrefix(sharePath).Subrouter()
	shareSubrouter.Use(rest.stripPrefix(sharePath))
	shareSubrouter.PathPrefix("").HandlerFunc(rest.share).Methods("GET")

	return router
}

func sendErr(w http.ResponseWriter, err error, msg string) {
	if err == nil {
		log.Printf("[ERROR] %s", errors.New(msg))
	} else {
		log.Printf("[ERROR] %s", errors.Wrap(err, msg))
	}

	w.Write([]byte(msg))
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

// stripPrefix removes prefix from request url
func (rest *Rest) stripPrefix(prefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			next.ServeHTTP(w, r)
		})
	}
}

// view return the current state of database in form of tree view
func (rest *Rest) view(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		sendErr(w, nil, "empty Authorization header")
		return
	}

	keys := splitPath(r.URL.Path)
	b, err := rest.Store.Get(token, keys)
	if err != nil {
		sendErr(w, err, "cannot view node")
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
	token := r.Header.Get("Authorization")
	if token == "" {
		sendErr(w, nil, "empty Authorization header")
		return
	}

	err := rest.Store.Put(token, keys, r.Body)
	if err != nil {
		sendErr(w, err, "cannot create node")
		return
	}

	b, err := rest.Store.Get(token, nil)
	if err != nil {
		sendErr(w, err, "data written successfully, but cannot view result")
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
	token := r.Header.Get("Authorization")
	if token == "" {
		sendErr(w, nil, "empty Authorization header")
		return
	}

	err := rest.Store.Delete(token, keys)
	if err != nil {
		sendErr(w, err, "cannot delete node")
		return
	}

	b, err := rest.Store.Get(token, nil)
	if err != nil {
		sendErr(w, err, "data deleted successfully, but cannot view result")
		return
	}

	if _, err = w.Write(b); err != nil {
		log.Println(err)
	}
}

type registerRequest struct {
	Email string `json:"email"`
}

// register creates token for user
// token will also be users private bucket name
func (rest *Rest) register(w http.ResponseWriter, r *http.Request) {
	req := &registerRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		sendErr(w, err, "invalid request body")
		return
	}

	// generate token
	b := make([]byte, 64)
	rand.Read(b)
	token := fmt.Sprintf("%x", b)

	_, err = rest.Email.Send(req.Email, token)
	if err != nil {
		sendErr(w, err, "cannot send email with token")
		return
	}

	// create collection with token value
	err = rest.Store.Create(token)
	if err != nil {
		sendErr(w, err, "cannot register")
		return
	}

	w.Write([]byte("registration successful. check email"))
}

// create shared folder
func (rest *Rest) share(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		sendErr(w, nil, "empty Authorization header")
		return
	}

	// generate token
	b := make([]byte, 16)
	rand.Read(b)
	sharedToken := fmt.Sprintf("%x", b)

	keys := splitPath(r.URL.Path)
	err := rest.Store.Share(token, keys, sharedToken)
	if err != nil {
		sendErr(w, err, "cannot share node")
		return
	}

	b, err = rest.Store.Get(token, []string{})
	if err != nil {
		sendErr(w, err, "cannot view node")
		return
	}
	if _, err = w.Write(b); err != nil {
		log.Println(err)
	}

}

// shared is a route for getting shared info
func (rest *Rest) shared(w http.ResponseWriter, r *http.Request) {
	keys := splitPath(r.URL.Path)
	if len(keys) <= 1 {
		sendErr(w, nil, "shearch path should be provided")
		return
	}
	b, err := rest.Store.Get(keys[0], keys[1:len(keys)-1])
	if err != nil {
		sendErr(w, err, "cannot view node")
		return
	}
	if _, err = w.Write(b); err != nil {
		log.Println(err)
	}
}

func (rest *Rest) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from DBFS"))
}
