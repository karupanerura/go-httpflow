package httpflow

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	"golang.org/x/text/encoding/ianaindex"
)

const contentTypeHeaderName = "Content-Type"

type ResponseHandler interface {
	HandleResponse(*http.Response) error
}

type RawResponseHandler struct {
	RawResponse *http.Response
}

var _ ResponseHandler = &RawResponseHandler{}

func (h *RawResponseHandler) HandleResponse(res *http.Response) error {
	h.RawResponse = res
	return nil
}

type NobodyResponseHandler struct {
	RawResponseHandler
	StatusCode int
	Header     http.Header
}

var _ ResponseHandler = &NobodyResponseHandler{}

func (h *NobodyResponseHandler) HandleResponse(res *http.Response) error {
	h.StatusCode = res.StatusCode
	h.Header = res.Header
	return h.RawResponseHandler.HandleResponse(res)
}

type BinaryResponseHandler struct {
	NobodyResponseHandler
	body []byte
}

var _ ResponseHandler = &BinaryResponseHandler{}

func (h *BinaryResponseHandler) HandleResponse(res *http.Response) (err error) {
	defer res.Body.Close()
	h.body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = h.NobodyResponseHandler.HandleResponse(res)
	return
}

func (h *BinaryResponseHandler) GetBody() []byte {
	return h.body
}

type StringResponseHandler struct {
	BinaryResponseHandler
}

var _ ResponseHandler = &StringResponseHandler{}

func (h *StringResponseHandler) GetBody() (string, error) {
	body := h.BinaryResponseHandler.GetBody()
	_, params, err := mime.ParseMediaType(h.Header.Get(contentTypeHeaderName))
	if err != nil {
		return "", err
	}

	if charset, ok := params["charset"]; ok {
		encoding, err := ianaindex.MIME.Encoding(charset)
		if err != nil {
			return "", err
		}

		body, err = encoding.NewDecoder().Bytes(body)
		if err != nil {
			return "", err
		}
	}

	return string(body), nil
}

type JsonResponseHandler struct {
	BinaryResponseHandler
}

var _ ResponseHandler = &JsonResponseHandler{}

func (h *JsonResponseHandler) IsJSON() bool {
	contentType := strings.TrimSpace(h.Header.Get(contentTypeHeaderName))
	parts := strings.SplitN(contentType, ";", 2)
	mediatype := parts[0]
	return mediatype == "application/json" || strings.HasPrefix(mediatype, "application/json+") || (strings.HasPrefix(mediatype, "application/") && strings.HasSuffix(mediatype, "+json"))
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
