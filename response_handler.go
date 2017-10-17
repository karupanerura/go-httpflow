package gotcha

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

const contentTypeHeaderName = "Content-Type"

type ResponseHandler interface {
	HandleResponse(*http.Response) error
}

type BinaryResponseHandler struct {
	body   []byte
	Header http.Header
}

var _ ResponseHandler = &BinaryResponseHandler{}

func (h *BinaryResponseHandler) HandleResponse(res *http.Response) (err error) {
	defer res.Body.Close()
	h.Header = res.Header
	h.body, err = ioutil.ReadAll(res.Body)
	return
}

func (h *BinaryResponseHandler) GetBody() []byte {
	return h.body
}

type StringResponseHandler struct {
	BinaryResponseHandler
}

var _ ResponseHandler = &StringResponseHandler{}

func (h *StringResponseHandler) GetBody() string {
	body := h.BinaryResponseHandler.GetBody()
	return string(body)
}

type JsonResponseHandler struct {
	BinaryResponseHandler
}

var _ ResponseHandler = &JsonResponseHandler{}

func (h *JsonResponseHandler) IsJSON() bool {
	contentType := h.Header.Get(contentTypeHeaderName)
	parts := strings.SplitN(contentType, ";", 2)
	mediatype := parts[0]
	return mediatype == "application/json" ||
		strings.HasPrefix(mediatype, "application/json+") ||
		mediatype == "application/problem+json" // RFC7807
}

func (h *JsonResponseHandler) GetDecoder() *json.Decoder {
	body := h.GetBody()
	reader := bytes.NewReader(body)
	return json.NewDecoder(reader)
}

func (h *JsonResponseHandler) DecodeJSON(v interface{}) error {
	if !h.IsJSON() {
		return &UnexpectedContentTypeError{
			ContentType: h.Header.Get(contentTypeHeaderName),
			Body:        h.GetBody(),
		}
	}
	return h.GetDecoder().Decode(v)
}
