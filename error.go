package httpflow

import "fmt"

type UnexpectedContentTypeError struct {
	ContentType string
	Body        []byte
}

func (e *UnexpectedContentTypeError) Error() string {
	return fmt.Sprintf("Unexpected Content-Type: %s", e.ContentType)
}

type UnexpectedStatusCodeError struct {
	StatusCode
}

func (e *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("Unexpected StatusCode %d", e.StatusCode)
}
