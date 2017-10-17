package httpflow

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mockErrReader struct {
	err error
}

func (m *mockErrReader) Read(_ []byte) (int, error) {
	return 0, m.err
}

func TestBinaryResponseHandler(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		res := &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte{123, 45, 67, 89}))}
		handler := &BinaryResponseHandler{}
		err := handler.HandleResponse(res)
		if err != nil {
			t.Fatal(err)
		}

		body := handler.GetBody()
		if diff := cmp.Diff(body, []byte{123, 45, 67, 89}); diff != "" {
			t.Errorf("Should no diff, but got: %s", diff)
		}
	})

	t.Run("ErrorOnRead", func(t *testing.T) {
		const msg = "READ ERROR DAYO"
		res := &http.Response{Body: ioutil.NopCloser(&mockErrReader{err: errors.New(msg)})}
		handler := &BinaryResponseHandler{}
		err := handler.HandleResponse(res)
		if err == nil {
			t.Fatal("Should not be nil")
		}
		if err.Error() != msg {
			t.Error(err)
		}
	})
}

func TestStringResponseHandler(t *testing.T) {
	res := &http.Response{Body: ioutil.NopCloser(strings.NewReader("foo"))}
	handler := &StringResponseHandler{}
	err := handler.HandleResponse(res)
	if err != nil {
		t.Fatal(err)
	}

	body := handler.GetBody()
	if body != "foo" {
		t.Errorf("Should get foo, but got: %s", body)
	}
}

func TestJsonResponseHandler(t *testing.T) {
	type Body struct {
		Foo string
	}
	t.Run("GetDecoder()", func(t *testing.T) {
		res := &http.Response{
			Body: ioutil.NopCloser(strings.NewReader(`{"foo":"bar"}`)),
		}
		handler := &JsonResponseHandler{}
		err := handler.HandleResponse(res)
		if err != nil {
			t.Fatal(err)
		}

		var body Body
		err = handler.GetDecoder().Decode(&body)
		if err != nil {
			t.Fatal(err)
		}
		if body.Foo != "bar" {
			t.Errorf("Should get bar, but got: %s", body.Foo)
		}
	})

	t.Run("DecodeJSON()", func(t *testing.T) {
		t.Run("Content-Type:application/json", func(t *testing.T) {
			res := &http.Response{
				Header: http.Header{"Content-Type": {"application/json"}},
				Body:   ioutil.NopCloser(strings.NewReader(`{"foo":"bar"}`)),
			}
			handler := &JsonResponseHandler{}
			err := handler.HandleResponse(res)
			if err != nil {
				t.Fatal(err)
			}

			var body Body
			err = handler.DecodeJSON(&body)
			if err != nil {
				t.Fatal(err)
			}
			if body.Foo != "bar" {
				t.Errorf("Should get bar, but got: %s", body.Foo)
			}
		})

		t.Run("Content-Type:text/plain", func(t *testing.T) {
			res := &http.Response{
				Header: http.Header{"Content-Type": {"text/plain"}},
				Body:   ioutil.NopCloser(bytes.NewReader([]byte{123, 45, 67, 89})),
			}
			handler := &JsonResponseHandler{}
			err := handler.HandleResponse(res)
			if err != nil {
				t.Fatal(err)
			}

			var body Body
			err = handler.DecodeJSON(&body)
			if err == nil {
				t.Fatal("Should not be nil")
			}
			if subErr, ok := err.(*UnexpectedContentTypeError); !ok {
				t.Error(err)
			} else if subErr.ContentType != "text/plain" {
				t.Errorf("Should get text/plain, but got: %s", subErr.ContentType)
			} else if diff := cmp.Diff(subErr.Body, []byte{123, 45, 67, 89}); diff != "" {
				t.Errorf("Should no diff, but got: %s", diff)
			}
		})
	})

	t.Run("IsJSON()", func(t *testing.T) {
		t.Run("application/json", func(t *testing.T) {
			res := &http.Response{
				Header: http.Header{"Content-Type": {"application/json"}},
				Body:   ioutil.NopCloser(strings.NewReader(`{"dummy":"dummy"}`)),
			}
			handler := &JsonResponseHandler{}
			err := handler.HandleResponse(res)
			if err != nil {
				t.Fatal(err)
			}

			if !handler.IsJSON() {
				t.Error("Should be detected as json")
			}
		})

		t.Run("application/json; charset=utf-8", func(t *testing.T) {
			res := &http.Response{
				Header: http.Header{"Content-Type": {"application/json; charset=utf-8"}},
				Body:   ioutil.NopCloser(strings.NewReader(`{"dummy":"dummy"}`)),
			}
			handler := &JsonResponseHandler{}
			err := handler.HandleResponse(res)
			if err != nil {
				t.Fatal(err)
			}

			if !handler.IsJSON() {
				t.Error("Should be detected as json")
			}
		})

		t.Run("application/json+hal", func(t *testing.T) {
			res := &http.Response{
				Header: http.Header{"Content-Type": {"application/json+hal"}},
				Body:   ioutil.NopCloser(strings.NewReader(`{"dummy":"dummy"}`)),
			}
			handler := &JsonResponseHandler{}
			err := handler.HandleResponse(res)
			if err != nil {
				t.Fatal(err)
			}

			if !handler.IsJSON() {
				t.Error("Should be detected as json")
			}
		})

		t.Run("application/problem+json", func(t *testing.T) {
			res := &http.Response{
				Header: http.Header{"Content-Type": {"application/problem+json"}},
				Body:   ioutil.NopCloser(strings.NewReader(`{"dummy":"dummy"}`)),
			}
			handler := &JsonResponseHandler{}
			err := handler.HandleResponse(res)
			if err != nil {
				t.Fatal(err)
			}

			if !handler.IsJSON() {
				t.Error("Should be detected as json")
			}
		})

		t.Run("text/html", func(t *testing.T) {
			res := &http.Response{
				Header: http.Header{"Content-Type": {"text/html"}},
				Body:   ioutil.NopCloser(strings.NewReader(`<html><head><title>dummy</title></head><body>dummy</body></html>`)),
			}
			handler := &JsonResponseHandler{}
			err := handler.HandleResponse(res)
			if err != nil {
				t.Fatal(err)
			}

			if handler.IsJSON() {
				t.Error("Should not be detected as json")
			}
		})
	})
}
