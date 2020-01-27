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
	APP_PORT            string `env:"APP_PORT" envDefault:"8080"`
	DB_PATH             string `env:"DB_PATH" envDefault:"/tmp/mydb.bolt"`
	MAILGUN_API_KEY     string `env:"MAILGUN_API_KEY" envDefault:""`
	MAILGUN_ROOT_DOMAIN string `env:"MAILGUN_ROOT_DOMAIN"`
	MAILGUN_SUBDOMAIN   string `env:"MAILGUN_SUBDOMAIN" envDefault:""`
	WHITELIST           string `env:"WHITELIST"`
	A                   string `env:"A"`
}

func main() {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		err = errors.Wrap(err, "error parsing environment variables")
		log.Fatal(err)
	}

	r := &rest.Rest{
		Store: &store.Store{
			Path: config.DB_PATH,
		},
		Email:     email.New(config.MAILGUN_API_KEY, config.MAILGUN_ROOT_DOMAIN, config.MAILGUN_SUBDOMAIN),
		Whitelist: config.WHITELIST,
	}

	fmt.Println("starting dbfs on localhost:" + config.APP_PORT)
	err := http.ListenAndServe(":"+config.APP_PORT, r.Router())
	if err != nil {
		log.Fatal(errors.Wrap(err, "error starting dbfs server"))
	}
}
