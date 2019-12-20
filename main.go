package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/mind-rot/dbfs/email"
	"github.com/mind-rot/dbfs/rest"
	"github.com/mind-rot/dbfs/store"
	"github.com/pkg/errors"
)

type Config struct {
	APP_PORT        string `env:"APP_PORT" envDefault:"8080"`
	DB_PATH         string `env:"DB_PATH" envDefault:"/tmp/mydb.bolt"`
	MAILGUN_API_KEY string `env:"MAILGUN_API_KEY" envDefault:""`
	MAILGUN_DOMAIN  string `env:"MAILGUN_DOMAIN" envDefault:""`
}

func main() {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		err = errors.Wrap(err, "error parsing environment variables")
		log.Fatal(err)
	}
	fmt.Println(config)

	r := &rest.Rest{
		Store: &store.Store{
			Path: config.DB_PATH,
		},
		Email: email.New(config.MAILGUN_API_KEY, config.MAILGUN_DOMAIN),
	}

	fmt.Println("starting dbfs on localhost:" + config.APP_PORT)
	err := http.ListenAndServe(":"+config.APP_PORT, r.Router())
	if err != nil {
		log.Fatal(errors.Wrap(err, "error starting dbfs server"))
	}
}
