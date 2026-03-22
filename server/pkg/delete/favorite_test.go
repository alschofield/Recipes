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
