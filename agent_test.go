package httpflow

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

type mockClient struct {
	mockResponse
	mockError error
}

func (c mockClient) Do(req *http.Request) (*http.Response, error) {
	return c.mockResponse.MockResponse(req), c.mockError
}

type mockResponse struct {
	statusCode int
	headersMap map[string]string
	body       []byte
}

func (r mockResponse) MockResponse(req *http.Request) *http.Response {
	status := strconv.Itoa(r.statusCode) + " " + http.StatusText(r.statusCode)
	header := http.Header{}
	for name, value := range r.headersMap {
		header.Set(name, value)
	}
	return &http.Response{
		Status:           status,
		StatusCode:       r.statusCode,
		Proto:            "HTTP/1.0",
		ProtoMajor:       1,
		ProtoMinor:       0,
		Header:           header,
		Body:             ioutil.NopCloser(bytes.NewReader(r.body)),
		ContentLength:    int64(len(r.body)),
		TransferEncoding: []string{},
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          req,
		TLS:              nil,
	}
}

type mockSession struct {
	request   *http.Request
	reqerr    error
	buildreq  int
	response  *http.Response
	reserr    error
	handleres int
}

func (m *mockSession) BuildRequest(context.Context) (*http.Request, error) {
	m.buildreq++
	return m.request, m.reqerr
}

func (m *mockSession) HandleResponse(res *http.Response) error {
	m.response = res
	m.handleres++
	return m.reserr
}

func TestNewAgent(t *testing.T) {
	client := &http.Client{}
	agent := NewAgent(client)
	if agent.Client != client {
		t.Errorf("Should got same pointer, but got: %+v", agent.Client)
	}
}

func TestAgentRunSession(t *testing.T) {
	agent := Agent{Client: mockClient{mockResponse: mockResponse{200, map[string]string{"Content-Type": "text/plain"}, []byte("this is example.com")}}}
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	session := &mockSession{request: req}
	err = agent.RunSession(session)
	if err != nil {
		t.Error(err)
	}
	if session.reserr != nil {
		t.Error(err)
	}

	if session.buildreq != 1 {
		t.Errorf("Should called once, but called %d times", session.buildreq)
	}
	if session.handleres != 1 {
		t.Errorf("Should called once, but called %d times", session.handleres)
	}

	res := session.response
	if res == nil {
		t.Fatal("Response should not be nil")
	}

	if res.StatusCode != 200 {
		t.Errorf("Should be 200, but got: %d", res.StatusCode)
	}
	if s := res.Header.Get("Content-Type"); s != "text/plain" {
		t.Errorf("Should be text/plain, but got: %s", s)
	}
	if body, err := ioutil.ReadAll(res.Body); string(body) != "this is example.com" || err != nil {
		t.Errorf("Should be this is example.com, but got: %s, err: %v", string(body), err)
	}
}

func TestAgentRunSessionCtx(t *testing.T) {
	t.Run("No Error", func(t *testing.T) {
		agent := Agent{Client: mockClient{mockResponse: mockResponse{200, map[string]string{"Content-Type": "text/plain"}, []byte("this is example.com")}}}
		req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
		if err != nil {
			t.Fatal(err)
		}

		session := &mockSession{request: req}
		err = agent.RunSessionCtx(context.Background(), session)
		if err != nil {
			t.Error(err)
		}
		if session.reserr != nil {
			t.Error(err)
		}

		if session.buildreq != 1 {
			t.Errorf("Should called once, but called %d times", session.buildreq)
		}
		if session.handleres != 1 {
			t.Errorf("Should called once, but called %d times", session.handleres)
		}

		res := session.response
		if res == nil {
			t.Fatal("Response should not be nil")
		}

		if res.StatusCode != 200 {
			t.Errorf("Should be 200, but got: %d", res.StatusCode)
		}
		if s := res.Header.Get("Content-Type"); s != "text/plain" {
			t.Errorf("Should be text/plain, but got: %s", s)
		}
		if body, err := ioutil.ReadAll(res.Body); string(body) != "this is example.com" || err != nil {
			t.Errorf("Should be this is example.com, but got: %s, err: %v", string(body), err)
		}
	})

	t.Run("Request Building Error", func(t *testing.T) {
		agent := Agent{Client: mockClient{}}

		const msg = "MOCK REQUEST BUILDING ERROR DAYO"
		session := &mockSession{reqerr: errors.New(msg)}
		err := agent.RunSessionCtx(context.Background(), session)
		if err == nil {
			t.Fatal("Should not be nil")
		}
		if err.Error() != msg {
			t.Error(err)
		}

		if session.buildreq != 1 {
			t.Errorf("Should called once, but called %d times", session.buildreq)
		}
		if session.handleres != 0 {
			t.Errorf("Should not called, but called %d times", session.handleres)
		}
	})

	t.Run("Request Error", func(t *testing.T) {
		const msg = "MOCK REQUEST ERROR DAYO"
		agent := Agent{Client: mockClient{mockError: errors.New(msg)}}

		req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
		if err != nil {
			t.Fatal(err)
		}

		session := &mockSession{request: req}
		err = agent.RunSessionCtx(context.Background(), session)
		if err == nil {
			t.Fatal("Should not be nil")
		}
		if err.Error() != msg {
			t.Error(err)
		}

		if session.buildreq != 1 {
			t.Errorf("Should called once, but called %d times", session.buildreq)
		}
		if session.handleres != 0 {
			t.Errorf("Should not called, but called %d times", session.handleres)
		}
	})

	t.Run("Response Handling Error", func(t *testing.T) {
		agent := Agent{Client: mockClient{mockResponse: mockResponse{200, map[string]string{"Content-Type": "text/plain"}, []byte("this is example.com")}}}
		req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
		if err != nil {
			t.Fatal(err)
		}

		const msg = "MOCK RESPONSE ERROR DAYO"
		session := &mockSession{request: req, reserr: errors.New(msg)}
		err = agent.RunSessionCtx(context.Background(), session)
		if err == nil {
			t.Fatal("Should not be nil")
		}
		if err.Error() != msg {
			t.Error(err)
		}

		if session.buildreq != 1 {
			t.Errorf("Should called once, but called %d times", session.buildreq)
		}
		if session.handleres != 1 {
			t.Errorf("Should called once, but called %d times", session.handleres)
		}
	})
}
