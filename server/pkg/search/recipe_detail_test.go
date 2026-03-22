package search

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRecipeDetailRequiresRecipeID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/recipes/detail/", nil)
	rr := httptest.NewRecorder()

	GetRecipeDetail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}
