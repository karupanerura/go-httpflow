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

type Requester interface {
	RequestBuilder
	ResponseHandler
}

func (a *Agent) Do(r Requester) error {
	return a.DoCtx(context.Background(), r)
}

func (a *Agent) DoCtx(ctx context.Context, r Requester) error {
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
