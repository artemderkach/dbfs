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
	// 	subrouter.PathPrefix("/").HandlerFunc(rest.put).Methods("POST")
	// 	subrouter.PathPrefix("/").HandlerFunc(rest.delete).Methods("DELETE")

	return router
}

func sendErr(w http.ResponseWriter, err error) {
	log.Print(err)
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
	w.Write(b)
}

// put creates new record in database. Returns state of database after write
// Take "multipart/form-data" request with "file" key
func (rest *Rest) put(w http.ResponseWriter, r *http.Request) {
	keys := splitPath(r.URL.Path)

	file, header, err := r.FormFile("file")
	if err != nil {
		rest.sendErr(errors.Wrap(err, "error parsing FormFile"))
		return
	}

	keys = append(keys, header.Filename)
	err = rest.Store.Put("public", keys, file)
	if err != nil {
		rest.sendErr(errors.Wrap(err, "error writing file to storage"))
		return
	}

	b, err := rest.Store.View("public")
	if err != nil {
		rest.sendErr(errors.Wrap(err, "error retrieving view data from database"))
		return
	}
	w.Write(b)
}

// delete removes value from database
// returnes current state of tree (GET / route)
// func (rest *Rest) delete(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
//
// 	collection := vars["collection"]
// 	err := validateCollection(collection)
// 	if err != nil {
// 		err = errors.Wrap(err, "invalid collection parameter")
// 		fmt.Println(err)
// 		if _, err = w.Write([]byte(err.Error())); err != nil {
// 			fmt.Println(err)
// 		}
// 		return
// 	}
//
// 	reg, err := regexp.Compile("/" + collection)
// 	if err != nil {
// 		err = errors.Wrap(err, "error parsing regexp")
// 		fmt.Println(err)
// 		if _, err = w.Write([]byte(err.Error())); err != nil {
// 			fmt.Println(err)
// 		}
// 		return
// 	}
// 	path := reg.ReplaceAllString(r.URL.Path, "")
//
// 	err = rest.Store.Delete(collection, path)
// 	if err != nil {
// 		err = errors.Wrap(err, "error deleting file from database")
// 		fmt.Println(err)
// 		if _, err = w.Write([]byte(err.Error())); err != nil {
// 			fmt.Println(err)
// 		}
// 		return
// 	}
//
// 	b, err := rest.Store.View(collection)
// 	if err != nil {
// 		err = errors.Wrap(err, "error retrieving view data from database")
// 		fmt.Println(err)
// 		if _, err = w.Write([]byte(err.Error())); err != nil {
// 			fmt.Println(err)
// 		}
// 		return
// 	}
// 	if _, err = w.Write(b); err != nil {
// 		fmt.Println(err)
// 	}
// }

// dowload returns the value for filename
// func (rest *Rest) download(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	reg, err := regexp.Compile("^.+download")
// 	if err != nil {
// 		err = errors.Wrap(err, "error parsing regexp")
// 		fmt.Println(err)
// 		w.Write([]byte(err.Error()))
// 		return
// 	}
// 	path := reg.ReplaceAllString(r.URL.Path, "")
//
// 	collection := vars["collection"]
// 	err = validateCollection(collection)
// 	if err != nil {
// 		err = errors.Wrap(err, "invalid collection parameter")
// 		fmt.Println(err)
// 		w.Write([]byte(err.Error()))
// 		return
// 	}
//
// 	b, err := rest.Store.Get(collection, path)
// 	if err != nil {
// 		err = errors.Wrap(err, "error retrieving file data from database")
// 		fmt.Println(err)
// 		w.Write([]byte(err.Error()))
// 		return
// 	}
// 	w.Write(b)
// }

func (rest *Rest) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from DBFS"))
}
