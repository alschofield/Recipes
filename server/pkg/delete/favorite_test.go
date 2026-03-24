package delete

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemoveFavoriteRequiresPathParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/favorites//", nil)
	rr := httptest.NewRecorder()

	RemoveFavorite(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestWriteRemoveFavoriteResponseReplayHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	writeRemoveFavoriteResponse(rr, 0)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if got := rr.Header().Get("Idempotency-Status"); got != "replayed" {
		t.Fatalf("expected replay header, got %q", got)
	}
}

func TestWriteRemoveFavoriteResponseNoReplayHeaderWhenDeleted(t *testing.T) {
	rr := httptest.NewRecorder()
	writeRemoveFavoriteResponse(rr, 1)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if got := rr.Header().Get("Idempotency-Status"); got != "" {
		t.Fatalf("expected no replay header, got %q", got)
	}
}
