package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	ts := httptest.NewServer(router())
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	require.Nil(t, err)

	msg, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)

	t.Fatal(string(msg))
}
