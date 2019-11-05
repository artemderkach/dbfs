package main

import (
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	// kill current process after 1 second
	go func() {
		time.Sleep(1000 * time.Millisecond)
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.Nil(t, e)
	}()

	go func() {
		main()
	}()

	// give service .5 seconds to start
	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080")
	require.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
