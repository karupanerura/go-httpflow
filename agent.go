package gotcha

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

func (a *Agent) RunSession(r Session) error {
	return a.RunSessionCtx(context.Background(), r)
}

func (a *Agent) RunSessionCtx(ctx context.Context, r Session) error {
	req, err := r.BuildRequest()
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	res, err := a.Client.Do(req)
	if err != nil {
		return err
	}

	return r.HandleResponse(res)
}
