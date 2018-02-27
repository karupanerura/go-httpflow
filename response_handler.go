package httpflow

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/text/encoding"
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
	expectedStatusCodes []int
	StatusCode
	Header http.Header
}

var _ ResponseHandler = &NobodyResponseHandler{}

func (h *NobodyResponseHandler) ExpectStatusCode(statusCodes ...int) {
	h.expectedStatusCodes = append(h.expectedStatusCodes, statusCodes...)
}

func (h *NobodyResponseHandler) HandleResponse(res *http.Response) error {
	h.StatusCode = StatusCode(res.StatusCode)
	h.Header = res.Header
	h.RawResponseHandler.HandleResponse(res) // always be nil

	if h.expectedStatusCodes != nil {
		ok := false
		for _, statusCode := range h.expectedStatusCodes {
			if res.StatusCode == statusCode {
				ok = true
				break
			}
		}
		if !ok {
			return &UnexpectedStatusCodeError{StatusCode: h.StatusCode}
		}
	}
	return nil
}

type BinaryResponseHandler struct {
	NobodyResponseHandler
	body []byte
}

var _ ResponseHandler = &BinaryResponseHandler{}

func (h *BinaryResponseHandler) HandleResponse(res *http.Response) (err error) {
	rawBody := res.Body
	defer rawBody.Close()

	h.body, err = ioutil.ReadAll(rawBody)
	if err != nil {
		return
	}
	res.Body = ioutil.NopCloser(bytes.NewReader(h.body))

	err = h.NobodyResponseHandler.HandleResponse(res)
	if uerr, ok := err.(*UnexpectedStatusCodeError); ok {
		uerr.Body = h.body
	}
	return
}

func (h *BinaryResponseHandler) Bytes() []byte {
	return h.body
}

type StringResponseHandler struct {
	BinaryResponseHandler
}

var _ ResponseHandler = &StringResponseHandler{}

func (h *StringResponseHandler) String() string {
	body := h.Bytes()
	return string(body)
}

func (h *StringResponseHandler) GetEncoding() (encoding.Encoding, error) {
	_, params, err := mime.ParseMediaType(h.Header.Get(contentTypeHeaderName))
	if err != nil {
		return nil, err
	}

	if charset, ok := params["charset"]; ok {
		return ianaindex.MIME.Encoding(charset)
	}
	return nil, nil
}

func (h *StringResponseHandler) GetDecodedString() (string, error) {
	body := h.String()

	encoding, err := h.GetEncoding()
	if err != nil {
		return "", err
	}

	if encoding != nil {
		return encoding.NewDecoder().String(body)
	}

	return body, nil
}

type JSONResponseHandler struct {
	BinaryResponseHandler
}

var _ ResponseHandler = &JSONResponseHandler{}

func (h *JSONResponseHandler) IsJSON() bool {
	contentType := strings.TrimSpace(h.Header.Get(contentTypeHeaderName))
	parts := strings.SplitN(contentType, ";", 2)
	mediatype := parts[0]
	return mediatype == "application/json" || strings.HasPrefix(mediatype, "application/json+") || (strings.HasPrefix(mediatype, "application/") && strings.HasSuffix(mediatype, "+json"))
}

func (h *JSONResponseHandler) GetDecoder() *json.Decoder {
	body := h.Bytes()
	reader := bytes.NewReader(body)
	return json.NewDecoder(reader)
}

func (h *JSONResponseHandler) DecodeJSON(v interface{}) error {
	if !h.IsJSON() {
		return &UnexpectedContentTypeError{
			ContentType: h.Header.Get(contentTypeHeaderName),
			Body:        h.Bytes(),
		}
	}
	return h.GetDecoder().Decode(v)
}

type FormResponseHandler struct {
	StringResponseHandler
}

var _ ResponseHandler = &FormResponseHandler{}

func (h *FormResponseHandler) IsForm() bool {
	contentType := strings.TrimSpace(h.Header.Get(contentTypeHeaderName))
	parts := strings.SplitN(contentType, ";", 2)
	mediatype := parts[0]
	return mediatype == "application/x-www-form-urlencoded"
}

func (h *FormResponseHandler) ParseForm() (url.Values, error) {
	if !h.IsForm() {
		return nil, &UnexpectedContentTypeError{
			ContentType: h.Header.Get(contentTypeHeaderName),
			Body:        h.Bytes(),
		}
	}

	// Don't follow enconding
	// SEE ALSO: https://www.w3.org/TR/html5/forms.html#application/x-www-form-urlencoded-encoding-algorithm
	body := h.String()
	return url.ParseQuery(body)
}
