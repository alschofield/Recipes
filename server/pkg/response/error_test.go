package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteError(rr, http.StatusBadRequest, "Invalid input", "BAD_REQUEST")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content-type application/json, got %s", got)
	}

	var body APIError
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Error != "Invalid input" || body.Code != "BAD_REQUEST" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestWriteErrorDetails(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteErrorDetails(rr, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED", "Missing token")

	var body APIError
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Details != "Missing token" {
		t.Fatalf("expected details to be set, got %q", body.Details)
	}
}
