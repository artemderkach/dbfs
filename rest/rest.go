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
	w.Write([]byte("db data"))
}

func (rest *Rest) put(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		err = errors.Wrap(err, "error parsing FormFile")
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
}

func (rest *Rest) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from DBFS"))
}
