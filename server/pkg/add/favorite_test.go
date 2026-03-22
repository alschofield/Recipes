package add

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddFavoriteRequiresPathParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/favorites//", nil)
	rr := httptest.NewRecorder()

	AddFavorite(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}
