package search

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListFavoritesRequiresUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/favorites", nil)
	rr := httptest.NewRecorder()

	ListFavorites(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}
