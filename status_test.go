package httpflow

import (
	"net/http"
	"testing"
)

func TestIsInformational(t *testing.T) {
	if StatusCode(99).IsInformational() {
		t.Error("Should be false")
	}
	if !StatusCode(100).IsInformational() {
		t.Error("Should be true")
	}
	if !StatusCode(199).IsInformational() {
		t.Error("Should be true")
	}
	if StatusCode(200).IsInformational() {
		t.Error("Should be false")
	}
}

func TestIsSuccessful(t *testing.T) {
	if StatusCode(199).IsSuccessful() {
		t.Error("Should be false")
	}
	if !StatusCode(200).IsSuccessful() {
		t.Error("Should be true")
	}
	if !StatusCode(299).IsSuccessful() {
		t.Error("Should be true")
	}
	if StatusCode(300).IsSuccessful() {
		t.Error("Should be false")
	}
}

func TestIsRedirection(t *testing.T) {
	if StatusCode(299).IsRedirection() {
		t.Error("Should be false")
	}
	if !StatusCode(300).IsRedirection() {
		t.Error("Should be true")
	}
	if !StatusCode(399).IsRedirection() {
		t.Error("Should be true")
	}
	if StatusCode(400).IsRedirection() {
		t.Error("Should be false")
	}
}

func TestIsClientError(t *testing.T) {
	if StatusCode(399).IsClientError() {
		t.Error("Should be false")
	}
	if !StatusCode(400).IsClientError() {
		t.Error("Should be true")
	}
	if !StatusCode(499).IsClientError() {
		t.Error("Should be true")
	}
	if StatusCode(500).IsClientError() {
		t.Error("Should be false")
	}
}

func TestIsServerError(t *testing.T) {
	if StatusCode(499).IsServerError() {
		t.Error("Should be false")
	}
	if !StatusCode(500).IsServerError() {
		t.Error("Should be true")
	}
	if !StatusCode(599).IsServerError() {
		t.Error("Should be true")
	}
	if StatusCode(600).IsServerError() {
		t.Error("Should be false")
	}
}

func TestString(t *testing.T) {
	if s := StatusCode(http.StatusOK).String(); s != "200 OK" {
		t.Errorf("Should be 200 OK, but got: %s", s)
	}
}
