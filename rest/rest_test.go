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

func TestView(t *testing.T) {
	URL := "/public/view"
	r, err := getRest()
	require.Nil(t, err)
	defer r.Store.Drop()

	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	resp, err := http.Get(ts.URL + URL)
	require.Nil(t, err)

	msg, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "Neo\nanswer\n", string(msg))
}

func TestPut(t *testing.T) {
	URL := "/public/put"
	r, err := getRest()
	require.Nil(t, err)
	defer r.Store.Drop()

	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	file := strings.NewReader("Ok, that was epic!\n")
	body, header, err := multipartFile("file", "filename.txt", file)
	require.Nil(t, err)

	_, err = http.Post(ts.URL+URL, header, body)
	require.Nil(t, err)

	view, err := r.Store.View("public")
	assert.Equal(t, "Neo\nanswer\nfilename.txt\n", string(view))
}

func TestGet(t *testing.T) {
	URL := "/public/download/answer"
	r, err := getRest()
	require.Nil(t, err)
	defer r.Store.Drop()

	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	resp, err := http.Get(ts.URL + URL)
	require.Nil(t, err)

	msg, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "42", string(msg))
}

func TestDelete(t *testing.T) {
	URL := "/public/delete/answer"
	r, err := getRest()
	require.Nil(t, err)
	defer r.Store.Drop()

	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, ts.URL+URL, nil)
	resp, err := client.Do(req)

	require.Nil(t, err)

	msg, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "Neo\n", string(msg))
}

func TestPrivate(t *testing.T) {
	tt := []struct {
		url      string
		header   string
		response string
	}{
		{
			"/private/download/try",
			"super_secret_password",
			"leave \"if err != nil\" alone",
		},
		{
			"/private/download/try",
			"super_wrong_password",
			"permission denied",
		},
	}

	r, err := getRest()
	require.Nil(t, err)
	defer r.Store.Drop()

	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	for _, test := range tt {
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, ts.URL+test.url, nil)
		require.Nil(t, err)
		req.Header.Add("Custom-Auth", test.header)

		resp, err := client.Do(req)
		require.Nil(t, err)

		body, err := ioutil.ReadAll(resp.Body)
		require.Nil(t, err)

		assert.Equal(t, test.response, string(body))
	}
}

func getRest() (*Rest, error) {
	s, err := getStore()
	if err != nil {
		return nil, err
	}
	r := &Rest{
		Store: s,
		APP_Pass:  "super_secret_password",
	}
	return r, nil
}

func getStore() (*store.Store, error) {
	s := &store.Store{
		Path: "/tmp/db123",
	}

	r := strings.NewReader("42")
	err := s.Put("public", "answer", r)
	if err != nil {
		return nil, err
	}

	r = strings.NewReader("The One")
	err = s.Put("public", "Neo", r)
	if err != nil {
		return nil, err
	}

	r = strings.NewReader("leave \"if err != nil\" alone")
	err = s.Put("private", "try", r)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func multipartFile(keyName string, fileName string, file io.Reader) (io.Reader, string, error) {
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(keyName, fileName)
	if err != nil {
		return nil, "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, "", err
	}
	writer.Close()

	return body, writer.FormDataContentType(), nil
}
