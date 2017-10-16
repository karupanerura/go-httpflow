package gotcha

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func mustParseURL(s string) *url.URL {
	url, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return url
}

func TestRawRequester(t *testing.T) {
	url := mustParseURL("http://localhost/")
	t.Run("GET", func(t *testing.T) {
		r := &RawRequestBuilder{
			RequestMethod: http.MethodGet,
			RequestURL:    url,
		}

		req, err := r.BuildRequest()
		if err != nil {
			t.Fatal(err)
		}
		if req.Method != http.MethodGet {
			t.Errorf("Should got GET method, but got: %s", req.Method)
		}
		if req.URL.String() != url.String() {
			t.Errorf("Should equals with %s, but got: %s", url, req.URL)
		}
	})

	t.Run("POST", func(t *testing.T) {
		header := http.Header{}
		header.Set("Content-Type", "text/plain")
		r := &RawRequestBuilder{
			RequestMethod: http.MethodPost,
			RequestHeader: header,
			RequestURL:    url,
			RequestBody:   strings.NewReader("foo"),
		}

		req, err := r.BuildRequest()
		if err != nil {
			t.Fatal(err)
		}
		if req.Method != http.MethodPost {
			t.Errorf("Should got POST method, but got: %s", req.Method)
		}
		if s := req.URL.String(); s != url.String() {
			t.Errorf("Should be %s, but got: %s", url, s)
		}
		if s := req.Header.Get("Content-Type"); s != "text/plain" {
			t.Errorf("Should be text/plain, but got: %s", s)
		}
		if body, err := ioutil.ReadAll(req.Body); string(body) != "foo" || err != nil {
			t.Errorf("Should be foo, but got: %s, error: %v", string(body), err)
		}
	})

	t.Run("Error", func(t *testing.T) {
		r := &RawRequestBuilder{
			RequestMethod: "INVALID!@#$%^&**()_+|-=\\`~",
			RequestURL:    url,
		}

		req, err := r.BuildRequest()
		if req != nil {
			t.Errorf("Should be nil, but got: %v", req)
		}
		if err == nil {
			t.Fatalf("Should not be nil")
		}
	})
}

func TestNobodyRequester(t *testing.T) {
	url := mustParseURL("http://localhost/")
	r := &NobodyRequestBuilder{
		RequestMethod: http.MethodGet,
		RequestURL:    url,
	}

	req, err := r.BuildRequest()
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != http.MethodGet {
		t.Errorf("Should got GET method, but got: %s", req.Method)
	}
	if req.URL.String() != url.String() {
		t.Errorf("Should equals with %s, but got: %s", url, req.URL)
	}
}

func TestJsonRequester(t *testing.T) {
	url := mustParseURL("http://localhost/")
	t.Run("GET", func(t *testing.T) {
		r := &JsonRequestBuilder{
			RequestMethod: http.MethodGet,
			RequestURL:    url,
		}

		req, err := r.BuildRequest()
		if err != nil {
			t.Fatal(err)
		}
		if req.Method != http.MethodGet {
			t.Errorf("Should got GET method, but got: %s", req.Method)
		}
		if req.URL.String() != url.String() {
			t.Errorf("Should equals with %s, but got: %s", url, req.URL)
		}
	})

	t.Run("POST", func(t *testing.T) {
		header := http.Header{}
		header.Set("Content-Type", "application/json")
		r := &JsonRequestBuilder{
			RequestMethod: http.MethodPost,
			RequestHeader: header,
			RequestURL:    url,
			RequestBody:   map[string]string{"foo": "bar"},
		}

		req, err := r.BuildRequest()
		if err != nil {
			t.Fatal(err)
		}
		if req.Method != http.MethodPost {
			t.Errorf("Should got POST method, but got: %s", req.Method)
		}
		if s := req.URL.String(); s != url.String() {
			t.Errorf("Should be %s, but got: %s", url, s)
		}
		if s := req.Header.Get("Content-Type"); s != "application/json" {
			t.Errorf("Should be application/json, but got: %s", s)
		}
		if body, err := ioutil.ReadAll(req.Body); string(body) != `{"foo":"bar"}` || err != nil {
			t.Errorf("Should be {\"foo\":\"bar\"}, but got: %s, error: %v", string(body), err)
		}
	})

	t.Run("Error", func(t *testing.T) {
		header := http.Header{}
		header.Set("Content-Type", "application/json")
		r := &JsonRequestBuilder{
			RequestMethod: http.MethodPost,
			RequestHeader: header,
			RequestURL:    url,
			RequestBody:   map[struct{}]struct{}{}, // invalid
		}

		req, err := r.BuildRequest()
		if req != nil {
			t.Errorf("Should be nil, but got: %v", req)
		}
		if err == nil {
			t.Fatalf("Should not be nil")
		}
	})
}
