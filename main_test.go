package main

import (
	"io/ioutil"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	go func() {
		time.Sleep(2000 * time.Millisecond)
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.Nil(t, e)
	}()

	go func() {
		main()
	}()

	time.Sleep(1000 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080")
	require.Nil(t, err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "Hello from DBFS", string(body))
}
