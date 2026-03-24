package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestIdempotencyKeyReplaysDuplicateMutation(t *testing.T) {
	var calls atomic.Int64

	handler := IdempotencyKey(5 * time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	req1 := httptest.NewRequest(http.MethodPost, "/users/new", strings.NewReader(`{"username":"a"}`))
	req1.Header.Set("Idempotency-Key", "signup-key-1")
	req1.RemoteAddr = "203.0.113.10:1000"
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/users/new", strings.NewReader(`{"username":"a"}`))
	req2.Header.Set("Idempotency-Key", "signup-key-1")
	req2.RemoteAddr = "203.0.113.10:1000"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if calls.Load() != 1 {
		t.Fatalf("expected one underlying call, got %d", calls.Load())
	}
	if rr1.Code != http.StatusCreated || rr2.Code != http.StatusCreated {
		t.Fatalf("expected 201 statuses, got %d and %d", rr1.Code, rr2.Code)
	}
	if rr2.Header().Get("Idempotency-Status") != "replayed" {
		t.Fatalf("expected replayed status header, got %q", rr2.Header().Get("Idempotency-Status"))
	}
}

func TestIdempotencyKeySkipsWhenHeaderMissing(t *testing.T) {
	var calls atomic.Int64

	handler := IdempotencyKey(5 * time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))

	req1 := httptest.NewRequest(http.MethodPost, "/users/new", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/users/new", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if calls.Load() != 2 {
		t.Fatalf("expected two underlying calls, got %d", calls.Load())
	}
}
