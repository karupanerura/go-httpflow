package httpflow

import (
	"context"
	"net/http"
)

var DefaultAgent = &Agent{Client: http.DefaultClient}

type Agent struct {
	Client HTTPClient
}

func NewAgent(client *http.Client) *Agent {
	return &Agent{Client: client}
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Session interface {
	RequestBuilder
	ResponseHandler
}

type contextSession struct {
	Session
	ctx context.Context
}

func (s *contextSession) BuildRequest() (*http.Request, error) {
	req, err := s.Session.BuildRequest()
	if err != nil {
		return nil, err
	}

	return req.WithContext(s.ctx), nil
}

func (a *Agent) RunSession(session Session) error {
	req, err := session.BuildRequest()
	if err != nil {
		return err
	}

	res, err := a.Client.Do(req)
	if err != nil {
		return err
	}

	return session.HandleResponse(res)
}

func (a *Agent) RunSessionCtx(ctx context.Context, session Session) error {
	return a.RunSession(&contextSession{ctx: ctx, Session: session})
}
