package httpflow

import "testing"

func TestUnexpectedContentTypeError(t *testing.T) {
	err := &UnexpectedContentTypeError{ContentType: "text/plain"}
	if s := err.Error(); s != "Unexpected Content-Type: text/plain" {
		t.Errorf("Unexpected error message: %s", s)
	}
}

func TestUnexpectedStatusCodeError(t *testing.T) {
	err := &UnexpectedStatusCodeError{StatusCode: 500}
	if s := err.Error(); s != "Unexpected StatusCode 500" {
		t.Errorf("Unexpected error message: %s", s)
	}
}
