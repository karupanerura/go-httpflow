package gotcha

import "fmt"

type UnexpectedContentTypeError struct {
	ContentType string
	Body        []byte
}

func (e *UnexpectedContentTypeError) Error() string {
	return fmt.Sprintf("Unexpected Content-Type: %s", e.ContentType)
}
