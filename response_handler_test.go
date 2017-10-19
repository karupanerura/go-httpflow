package httpflow

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
)

type mockErrReader struct {
	err error
}

func (m *mockErrReader) Read(_ []byte) (int, error) {
	return 0, m.err
}

func TestRawResponseHandler(t *testing.T) {
	res := &http.Response{}
	handler := &RawResponseHandler{}
	err := handler.HandleResponse(res)
	if err != nil {
		t.Fatal(err)
	}

	if handler.RawResponse != res {
		t.Errorf("Should be same pointer, but got: %d", handler.RawResponse)
	}
}

func TestNobodyResponseHandler(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		header := http.Header{}
		header.Set("X-Waiwai", "wai-wai-")
		res := &http.Response{StatusCode: 200, Header: header}
		handler := &NobodyResponseHandler{}
		err := handler.HandleResponse(res)
		if err != nil {
			t.Fatal(err)
		}

		if handler.StatusCode != 200 {
			t.Errorf("Should be 200, but got: %d", res.StatusCode)
		}
		if diff := cmp.Diff(handler.Header, header); diff != "" {
			t.Errorf("Should no diff, but got: %s", diff)
		}
	})

	t.Run("ExpectStatusCode", func(t *testing.T) {
		t.Run("Expected", func(t *testing.T) {
			handler := &NobodyResponseHandler{}
			handler.ExpectStatusCode(200, 201)
			res := &http.Response{StatusCode: 200}
			err := handler.HandleResponse(res)
			if err != nil {
				t.Fatal(err)
			}

			if handler.StatusCode != 200 {
				t.Errorf("Should be 200, but got: %d", res.StatusCode)
			}
		})

		t.Run("Unexpected", func(t *testing.T) {
			handler := &NobodyResponseHandler{}
			handler.ExpectStatusCode(200, 201)
			res := &http.Response{StatusCode: 500}
			err := handler.HandleResponse(res)
			if err == nil {
				t.Fatal("Should not be nil")
			}

			if subErr, ok := err.(*UnexpectedStatusCodeError); !ok {
				t.Error(err)
			} else if subErr.StatusCode != 500 {
				t.Errorf("Should get 500, but got: %d", subErr.StatusCode)
			}
		})
	})
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

func mustEncodeString(e encoding.Encoding, src string) string {
	encoder := e.NewEncoder()
	dst, err := encoder.String(src)
	if err != nil {
		panic(err)
	}
	return dst
}

func TestStringResponseHandler(t *testing.T) {
	t.Run("text/plain", func(t *testing.T) {
		res := &http.Response{
			Header: http.Header{"Content-Type": {"text/plain"}},
			Body:   ioutil.NopCloser(strings.NewReader("foo")),
		}
		handler := &StringResponseHandler{}
		err := handler.HandleResponse(res)
		if err != nil {
			t.Fatal(err)
		}

		body, err := handler.GetBody()
		if err != nil {
			t.Fatal(err)
		}
		if body != "foo" {
			t.Errorf("Should get foo, but got: %s", body)
		}
	})

	t.Run("text/plain; charset=utf-8", func(t *testing.T) {
		res := &http.Response{
			Header: http.Header{"Content-Type": {"text/plain; charset=utf-8"}},
			Body:   ioutil.NopCloser(strings.NewReader(mustEncodeString(unicode.UTF8, "わかめ"))),
		}
		handler := &StringResponseHandler{}
		err := handler.HandleResponse(res)
		if err != nil {
			t.Fatal(err)
		}

		body, err := handler.GetBody()
		if err != nil {
			t.Fatal(err)
		}
		if body != "わかめ" {
			t.Errorf("Should get わかめ, but got: %s", body)
		}
	})

	t.Run("text/plain; charset=Shift_JIS", func(t *testing.T) {
		res := &http.Response{
			Header: http.Header{"Content-Type": {"text/plain; charset=Shift_JIS"}},
			Body:   ioutil.NopCloser(strings.NewReader(mustEncodeString(japanese.ShiftJIS, "かつを"))),
		}
		handler := &StringResponseHandler{}
		err := handler.HandleResponse(res)
		if err != nil {
			t.Fatal(err)
		}

		body, err := handler.GetBody()
		if err != nil {
			t.Fatal(err)
		}
		if body != "かつを" {
			t.Errorf("Should get かつを, but got: %s", body)
		}
	})

	t.Run("text/plain; charset=invalid-charset", func(t *testing.T) {
		res := &http.Response{
			Header: http.Header{"Content-Type": {"text/plain; charset=invalid-charset"}},
			Body:   ioutil.NopCloser(strings.NewReader("naiyo")),
		}
		handler := &StringResponseHandler{}
		err := handler.HandleResponse(res)
		if err != nil {
			t.Fatal(err)
		}

		body, err := handler.GetBody()
		if err == nil {
			t.Fatal("Should not be nil")
		}
		if body != "" {
			t.Errorf("Should get empty string, but got: %s", body)
		}
	})

	t.Run("invalid-content-type", func(t *testing.T) {
		res := &http.Response{
			Header: http.Header{"Content-Type": {"invalid-content-type^$#%#@$!@@#&*"}},
			Body:   ioutil.NopCloser(strings.NewReader("naiyo")),
		}
		handler := &StringResponseHandler{}
		err := handler.HandleResponse(res)
		if err != nil {
			t.Fatal(err)
		}

		body, err := handler.GetBody()
		if err == nil {
			t.Fatal("Should not be nil")
		}
		if body != "" {
			t.Errorf("Should get empty string, but got: %s", body)
		}
	})
}

func TestJSONResponseHandler(t *testing.T) {
	type Body struct {
		Foo string
	}
	t.Run("GetDecoder()", func(t *testing.T) {
		res := &http.Response{
			Body: ioutil.NopCloser(strings.NewReader(`{"foo":"bar"}`)),
		}
		handler := &JSONResponseHandler{}
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
			handler := &JSONResponseHandler{}
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
			handler := &JSONResponseHandler{}
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
			handler := &JSONResponseHandler{}
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
			handler := &JSONResponseHandler{}
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
			handler := &JSONResponseHandler{}
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
			handler := &JSONResponseHandler{}
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
			handler := &JSONResponseHandler{}
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
