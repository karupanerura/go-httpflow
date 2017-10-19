package httpflow

import (
	"net/http"
	"strconv"
)

type StatusCode int

func (c StatusCode) IsInformational() bool {
	return 100 <= c && c < 200
}

func (c StatusCode) IsSuccessful() bool {
	return 200 <= c && c < 300
}

func (c StatusCode) IsRedirection() bool {
	return 300 <= c && c < 400
}

func (c StatusCode) IsClientError() bool {
	return 400 <= c && c < 500
}

func (c StatusCode) IsServerError() bool {
	return 500 <= c && c <= 599
}

func (c StatusCode) String() string {
	i := int(c)
	return strconv.Itoa(i) + " " + http.StatusText(i)
}
