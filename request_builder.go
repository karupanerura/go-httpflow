package httpflow

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type RequestBuilder interface {
	BuildRequest() (*http.Request, error)
}

type RawRequestBuilder struct {
	RequestMethod string
	RequestHeader http.Header
	RequestURL    *url.URL
	RequestBody   io.Reader
}

var _ RequestBuilder = &RawRequestBuilder{}

func (r *RawRequestBuilder) BuildRequest() (*http.Request, error) {
	req, err := http.NewRequest(r.RequestMethod, r.RequestURL.String(), r.RequestBody)
	if err != nil {
		return nil, err
	}

	if r.RequestHeader != nil {
		for name := range r.RequestHeader {
			value := r.RequestHeader.Get(name)
			req.Header.Set(name, value)
		}
	}
	return req, nil
}

type NobodyRequestBuilder struct {
	RequestMethod string
	RequestHeader http.Header
	RequestURL    *url.URL
}

var _ RequestBuilder = &NobodyRequestBuilder{}

func (r *NobodyRequestBuilder) BuildRequest() (*http.Request, error) {
	raw := &RawRequestBuilder{
		RequestMethod: r.RequestMethod,
		RequestHeader: r.RequestHeader,
		RequestURL:    r.RequestURL,
	}
	return raw.BuildRequest()
}

type JSONRequestBuilder struct {
	RequestMethod string
	RequestHeader http.Header
	RequestURL    *url.URL
	RequestBody   interface{}
}

var _ RequestBuilder = &JSONRequestBuilder{}

func (r *JSONRequestBuilder) BuildRequest() (*http.Request, error) {
	var header http.Header
	if r.RequestHeader == nil {
		header = http.Header{}
		header.Set(contentTypeHeaderName, "application/json")
	} else if r.RequestHeader.Get(contentTypeHeaderName) == "" {
		header = http.Header{}
		for name, value := range r.RequestHeader {
			header[name] = value
		}
		header.Set(contentTypeHeaderName, "application/json")
	} else {
		header = r.RequestHeader
	}

	var reader io.Reader
	if r.RequestBody != nil {
		body, err := json.Marshal(r.RequestBody)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(body)
	}

	raw := &RawRequestBuilder{
		RequestMethod: r.RequestMethod,
		RequestHeader: header,
		RequestURL:    r.RequestURL,
		RequestBody:   reader,
	}
	return raw.BuildRequest()
}
