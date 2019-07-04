package rest

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mind-rot/dbfs/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHome(t *testing.T) {
	r := Rest{}
	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	require.Nil(t, err)

	msg, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "Hello from DBFS", string(msg))
}

func TestPut(t *testing.T) {
	URL := "/put"
	r, err := getRest()
	require.Nil(t, err)
	defer r.Store.Drop()

	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	file := strings.NewReader("Ok, that was epic!")
	body, err := multipartFile("file", "filename.txt", file)
	require.Nil(t, err)

	resp, err := http.Post(ts.URL+URL, "multipart/form-data", body)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// func TestView(t *testing.T) {
// 	URL := "/view"
// 	r := Rest{}
// 	ts := httptest.NewServer(r.Router())
// 	defer ts.Close()

// 	resp, err := http.Get(ts.URL + URL)
// 	require.Nil(t, err)

// 	msg, err := ioutil.ReadAll(resp.Body)
// 	require.Nil(t, err)
// 	assert.Equal(t, "db data", string(msg))
// }
func getRest() (*Rest, error) {
	s, err := getStore()
	if err != nil {
		return nil, err
	}
	r := &Rest{
		Store: s,
	}
	return r, nil
}

func getStore() (*store.Store, error) {
	s := &store.Store{
		Path:       "/tmp/db123",
		Collection: "files",
	}
	err := s.Init()
	if err != nil {
		return nil, err
	}

	return s, err
}

func multipartFile(keyName string, fileName string, file io.Reader) (io.Reader, error) {
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	_, err := writer.CreateFormFile(keyName, fileName)
	if err != nil {
		return nil, err
	}

	return body, nil
}
