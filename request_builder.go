package httpflow

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RequestBuilder interface {
	BuildRequest(context.Context) (*http.Request, error)
}

type RawRequestBuilder struct {
	RequestMethod      string
	RequestHeader      http.Header
	RequestURL         *url.URL
	RequestBody        io.Reader
	DefaultContentType string
}

var _ RequestBuilder = &RawRequestBuilder{}

func (r *RawRequestBuilder) BuildRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequest(r.RequestMethod, r.RequestURL.String(), r.RequestBody)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	if r.RequestHeader != nil {
		for name := range r.RequestHeader {
			value := r.RequestHeader.Get(name)
			req.Header.Set(name, value)
		}
	}

	if r.DefaultContentType != "" {
		if req.Header.Get(contentTypeHeaderName) == "" {
			req.Header.Set(contentTypeHeaderName, r.DefaultContentType)
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

func (r *NobodyRequestBuilder) BuildRequest(ctx context.Context) (*http.Request, error) {
	raw := &RawRequestBuilder{
		RequestMethod: r.RequestMethod,
		RequestHeader: r.RequestHeader,
		RequestURL:    r.RequestURL,
	}
	return raw.BuildRequest(ctx)
}

type FormRequestBuilder struct {
	RequestMethod string
	RequestHeader http.Header
	RequestURL    *url.URL
	RequestBody   url.Values
}

var _ RequestBuilder = &FormRequestBuilder{}

func (r *FormRequestBuilder) BuildRequest(ctx context.Context) (*http.Request, error) {
	var reader io.Reader
	if r.RequestBody != nil {
		reader = strings.NewReader(r.RequestBody.Encode())
	}

	raw := &RawRequestBuilder{
		RequestMethod:      r.RequestMethod,
		RequestHeader:      r.RequestHeader,
		RequestURL:         r.RequestURL,
		RequestBody:        reader,
		DefaultContentType: "application/x-www-form-urlencoded",
	}
	return raw.BuildRequest(ctx)
}

type JSONRequestBuilder struct {
	RequestMethod string
	RequestHeader http.Header
	RequestURL    *url.URL
	RequestBody   interface{}
}

var _ RequestBuilder = &JSONRequestBuilder{}

func (r *JSONRequestBuilder) BuildRequest(ctx context.Context) (*http.Request, error) {
	var reader io.Reader
	if r.RequestBody != nil {
		body, err := json.Marshal(r.RequestBody)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(body)
	}

	raw := &RawRequestBuilder{
		RequestMethod:      r.RequestMethod,
		RequestHeader:      r.RequestHeader,
		RequestURL:         r.RequestURL,
		RequestBody:        reader,
		DefaultContentType: "application/json",
	}
	return raw.BuildRequest(ctx)
}
