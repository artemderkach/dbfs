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

const defaultCollection = "public"

func TestView(t *testing.T) {
	tt := []struct {
		Path         string
		ResponseBody string
		Token        string
	}{
		{
			"",
			"Neo\nanswer\nme\n  and\nmust\n  have\n    been\n      like\n",
			defaultCollection,
		},
		{
			"/",
			"Neo\nanswer\nme\n  and\nmust\n  have\n    been\n      like\n",
			defaultCollection,
		},
		{
			"/must",
			"have\n  been\n    like\n",
			defaultCollection,
		},
		{
			"/answer",
			"42",
			defaultCollection,
		},
		{
			"/invalid",
			"cannot view node",
			"invalid token",
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
		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		require.Nil(t, err)
		req.Header.Set("Authorization", test.Token)

		resp, err := client.Do(req)
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
		Token        string
	}{
		{
			"/new",
			"data",
			"Neo\nanswer\nme\n  and\nmust\n  have\n    been\n      like\nnew\n",
			defaultCollection,
		},
		{
			"/must",
			"some more data",
			"cannot create node",
			defaultCollection,
		},
		{
			"/some/path",
			"content",
			"cannot create node",
			"invalid token",
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

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPost, u.String(), file)
		require.Nil(t, err)
		req.Header.Set("Authorization", test.Token)
		req.Header.Set("Content-Type", "application/octet-stream")

		resp, err := client.Do(req)
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
		Token        string
	}{
		{
			"/must/have",
			"Neo\nanswer\nme\n  and\nmust\n",
			defaultCollection,
		},
		{
			"/must/have",
			"cannot delete node",
			defaultCollection,
		},
		{
			"/must/have",
			"empty Authorization header",
			"",
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
		req.Header.Set("Authorization", test.Token)

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
			"registration successful. check email",
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

	err := s.Create(defaultCollection)
	if err != nil {
		return nil, err
	}

	r := strings.NewReader("42")
	err = s.Put(defaultCollection, []string{"answer"}, r)
	if err != nil {
		return nil, err
	}

	r = strings.NewReader("The One")
	err = s.Put(defaultCollection, []string{"Neo"}, r)
	if err != nil {
		return nil, err
	}

	r = strings.NewReader("The Boys")
	err = s.Put(defaultCollection, []string{"me", "and"}, r)
	if err != nil {
		return nil, err
	}

	r = strings.NewReader("blinking guy")
	err = s.Put(defaultCollection, []string{"must", "have", "been", "like"}, r)
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
