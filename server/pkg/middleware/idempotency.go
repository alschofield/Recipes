package middleware

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type idempotencyContextKey string

const idempotencyKeyContextKey idempotencyContextKey = "recipes.idempotency.key"

type idempotencyEntry struct {
	status      int
	body        []byte
	contentType string
	expiresAt   time.Time
}

type idempotencyStore struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]idempotencyEntry
}

func newIdempotencyStore(ttl time.Duration) *idempotencyStore {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &idempotencyStore{
		ttl:     ttl,
		entries: map[string]idempotencyEntry{},
	}
}

func (s *idempotencyStore) get(key string, now time.Time) (idempotencyEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[key]
	if !ok {
		return idempotencyEntry{}, false
	}
	if now.After(entry.expiresAt) {
		delete(s.entries, key)
		return idempotencyEntry{}, false
	}
	return entry, true
}

func (s *idempotencyStore) put(key string, status int, body []byte, contentType string, now time.Time) {
	if status >= 500 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[key] = idempotencyEntry{
		status:      status,
		body:        append([]byte(nil), body...),
		contentType: contentType,
		expiresAt:   now.Add(s.ttl),
	}
}

type captureResponseWriter struct {
	header      http.Header
	status      int
	body        bytes.Buffer
	wroteHeader bool
}

func newCaptureResponseWriter() *captureResponseWriter {
	return &captureResponseWriter{header: make(http.Header)}
}

func (w *captureResponseWriter) Header() http.Header {
	return w.header
}

func (w *captureResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.status = statusCode
	w.wroteHeader = true
}

func (w *captureResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(p)
}

// ParseIdempotencyTTL parses TTL duration with fallback.
func ParseIdempotencyTTL(raw string, fallback time.Duration) time.Duration {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}

	v, err := time.ParseDuration(raw)
	if err != nil || v <= 0 {
		return fallback
	}

	return v
}

// IdempotencyKey replays responses for duplicate mutation requests using Idempotency-Key header.
func IdempotencyKey(ttl time.Duration) Middleware {
	store := newIdempotencyStore(ttl)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			rawKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
			if rawKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			if len(rawKey) > 128 {
				rawKey = rawKey[:128]
			}

			scope := requestIP(r)
			if principal, ok := PrincipalFromContext(r.Context()); ok {
				scope = principal.UserID
			}

			composite := strings.Join([]string{r.Method, r.URL.Path, scope, rawKey}, "|")
			now := time.Now().UTC()

			if entry, ok := store.get(composite, now); ok {
				if entry.contentType != "" {
					w.Header().Set("Content-Type", entry.contentType)
				}
				w.Header().Set("Idempotency-Status", "replayed")
				w.Header().Set("Idempotency-Key", rawKey)
				w.WriteHeader(entry.status)
				_, _ = w.Write(entry.body)
				return
			}

			capture := newCaptureResponseWriter()
			ctx := context.WithValue(r.Context(), idempotencyKeyContextKey, rawKey)
			next.ServeHTTP(capture, r.WithContext(ctx))

			contentType := capture.header.Get("Content-Type")
			store.put(composite, capture.status, capture.body.Bytes(), contentType, now)

			for k, values := range capture.header {
				for _, v := range values {
					w.Header().Add(k, v)
				}
			}
			w.Header().Set("Idempotency-Status", "recorded")
			w.Header().Set("Idempotency-Key", rawKey)
			w.Header().Set("Idempotency-TTL-Seconds", strconv.Itoa(int(ttl.Seconds())))
			w.WriteHeader(capture.status)
			_, _ = w.Write(capture.body.Bytes())
		})
	}
}
