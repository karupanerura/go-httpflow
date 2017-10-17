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

type JsonRequestBuilder struct {
	RequestMethod string
	RequestHeader http.Header
	RequestURL    *url.URL
	RequestBody   interface{}
}

var _ RequestBuilder = &JsonRequestBuilder{}

func (r *JsonRequestBuilder) BuildRequest() (*http.Request, error) {
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
		RequestHeader: r.RequestHeader,
		RequestURL:    r.RequestURL,
		RequestBody:   reader,
	}
	return raw.BuildRequest()
}
