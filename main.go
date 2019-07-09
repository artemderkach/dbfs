package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mind-rot/dbfs/rest"
	"github.com/mind-rot/dbfs/store"
	"github.com/pkg/errors"
)

type App struct {
	Rest *rest.Rest
	Env  *Env
}

type Env struct {
	APP_PORT string
	DB_PATH  string
	APP_PASS string
}

func main() {
	fmt.Println("running dbfs on :8080")

	env := parseEnv()

	app := &App{
		Rest: &rest.Rest{
			Store: &store.Store{
				Path: env.DB_PATH,
			},
			APP_PASS: env.APP_PASS,
		},
		Env: env,
	}

	err := http.ListenAndServe(":"+env.APP_PORT, app.Rest.Router())
	if err != nil {
		panic(errors.Wrap(err, "error starting dbfs server"))
	}
}

func parseEnv() *Env {
	env := &Env{}

	port, exists := os.LookupEnv("APP_PORT")
	if !exists {
		port = "8080"
	}
	env.APP_PORT = port

	dbPath, exists := os.LookupEnv("DB_PATH")
	if !exists {
		dbPath = "/tmp/mydb.bolt"
	}
	env.DB_PATH = dbPath

	pass, exists := os.LookupEnv("APP_PASS")
	if !exists {
		dbPath = ""
	}
	env.APP_PASS = pass

	return env
}
