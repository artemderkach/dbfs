package rest

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/mind-rot/dbfs/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestView(t *testing.T) {
	tt := []struct {
		Path         string
		ResponseBody string
	}{
		{
			"",
			"Neo\nanswer\nme\n  and\nmust\n  have\n    been\n      like\n",
		},
		{
			"/",
			"Neo\nanswer\nme\n  and\nmust\n  have\n    been\n      like\n",
		},
		{
			"/must",
			"have\n  been\n    like\n",
		},
		{
			"/answer",
			"42",
		},
		{
			"/invalid",
			"error retrieving view data from database: error getting elements from bucket: bucket \"invalid\" not found",
		},
	}

	r, err := getRest()
	require.Nil(t, err)
	defer r.Store.Drop()

	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	for _, test := range tt {
		u, err := url.Parse(ts.URL)
		u.Path = path.Join(u.Path, basePath, test.Path)

		resp, err := http.Get(u.String())
		require.Nil(t, err)

		msg, err := ioutil.ReadAll(resp.Body)
		require.Nil(t, err)

		assert.Equal(t, test.ResponseBody, string(msg))
	}
}

func TestPut(t *testing.T) {
	tt := []struct {
		Path         string
		FileContent  string
		ResponseBody string
	}{
		{
			"/new",
			"data",
			"Neo\nanswer\nme\n  and\nmust\n  have\n    been\n      like\nnew\n",
		},
		{
			"/must",
			"some more data",
			"error writing file to storage: error updating database: name \"must\" already used",
		},
	}

	r, err := getRest()
	require.Nil(t, err)
	defer r.Store.Drop()

	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	for _, test := range tt {
		file := strings.NewReader(test.FileContent)
		require.Nil(t, err)

		u, err := url.Parse(ts.URL)
		u.Path = path.Join(u.Path, basePath, test.Path)

		resp, err := http.Post(u.String(), "application/octet-stream", file)
		require.Nil(t, err)

		msg, err := ioutil.ReadAll(resp.Body)
		require.Nil(t, err)

		assert.Equal(t, test.ResponseBody, string(msg))
	}
}

func TestDelete(t *testing.T) {
	tt := []struct {
		Path         string
		ResponseBody string
	}{
		{
			"/must/have",
			"Neo\nanswer\nme\n  and\nmust\n",
		},
		{
			"/must/have",
			"error deleting file from database: error updating database: bucket not found",
		},
	}

	r, err := getRest()
	require.Nil(t, err)
	defer r.Store.Drop()

	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	for _, test := range tt {
		u, err := url.Parse(ts.URL)
		u.Path = path.Join(u.Path, basePath, test.Path)

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
		require.Nil(t, err)

		resp, err := client.Do(req)
		require.Nil(t, err)

		msg, err := ioutil.ReadAll(resp.Body)
		require.Nil(t, err)

		assert.Equal(t, test.ResponseBody, string(msg))
	}
}

func TestRegister(t *testing.T) {
	tt := []struct {
		Email        string
		ResponseBody string
	}{
		{
			"random@gmail.com",
			"registered",
		},
	}

	r, err := getRest()
	require.Nil(t, err)
	defer r.Store.Drop()

	ts := httptest.NewServer(r.Router())
	defer ts.Close()

	for _, test := range tt {
		req := &registerRequest{
			Email: test.Email,
		}
		reqBody, err := json.Marshal(req)
		require.Nil(t, err)

		u, err := url.Parse(ts.URL)
		u.Path = path.Join(u.Path, "/register")

		resp, err := http.Post(u.String(), "application/json", bytes.NewReader(reqBody))
		require.Nil(t, err)

		msg, err := ioutil.ReadAll(resp.Body)
		require.Nil(t, err)

		assert.Equal(t, test.ResponseBody, string(msg))
	}
}

// func TestPrivate(t *testing.T) {
// 	tt := []struct {
// 		url      string
// 		header   string
// 		response string
// 	}{
// 		{
// 			"/private/download/try",
// 			"1",
// 			"leave \"if err != nil\" alone",
// 		},
// 		{
// 			"/private/download/try",
// 			"2",
// 			"permission denied",
// 		},
// 	}
//
// 	r, err := getRest()
// 	require.Nil(t, err)
// 	defer r.Store.Drop()
//
// 	ts := httptest.NewServer(r.Router())
// 	defer ts.Close()
//
// 	for _, test := range tt {
// 		client := &http.Client{}
// 		req, err := http.NewRequest(http.MethodGet, ts.URL+test.url, nil)
// 		require.Nil(t, err)
// 		req.Header.Add("Custom-Auth", test.header)
//
// 		resp, err := client.Do(req)
// 		require.Nil(t, err)
//
// 		body, err := ioutil.ReadAll(resp.Body)
// 		require.Nil(t, err)
//
// 		assert.Equal(t, test.response, string(body))
// 	}
//
// }

func getRest() (*Rest, error) {
	s, err := getStore()
	if err != nil {
		return nil, err
	}

	r := &Rest{
		Store: s,
		Email: &emailMock{},
	}

	return r, nil
}

type emailMock struct{}

func (e *emailMock) Send(targetEmail, msgBody string) (string, error) {
	return "OK", nil
}

func getStore() (*store.Store, error) {
	s := &store.Store{
		Path: "/tmp/db123",
	}
	r := strings.NewReader("42")
	err := s.Put("public", []string{"answer"}, r)
	if err != nil {
		return nil, err
	}

	r = strings.NewReader("The One")
	err = s.Put("public", []string{"Neo"}, r)
	if err != nil {
		return nil, err
	}

	r = strings.NewReader("The Boys")
	err = s.Put("public", []string{"me", "and"}, r)
	if err != nil {
		return nil, err
	}

	r = strings.NewReader("blinking guy")
	err = s.Put("public", []string{"must", "have", "been", "like"}, r)
	if err != nil {
		return nil, err
	}
	// r = strings.NewReader("leave \"if err != nil\" alone")
	// err = s.Put("private", "try", r)
	// if err != nil {
	// 	return nil, err
	// }

	return s, nil
}
