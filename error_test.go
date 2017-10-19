package httpflow

import "testing"

func TestUnexpectedContentTypeError(t *testing.T) {
	err := &UnexpectedContentTypeError{ContentType: "text/plain"}
	if s := err.Error(); s != "Unexpected Content-Type: text/plain" {
		t.Errorf("Unexpected error message: %s", s)
	}
}
