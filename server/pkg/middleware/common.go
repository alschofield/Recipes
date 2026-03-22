package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"recipes/pkg/response"

	"golang.org/x/time/rate"
)

// Middleware transforms an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain applies middleware in left-to-right order.
func Chain(handler http.Handler, middleware ...Middleware) http.Handler {
	wrapped := handler
	for i := len(middleware) - 1; i >= 0; i-- {
		wrapped = middleware[i](wrapped)
	}
	return wrapped
}

// Recoverer converts unexpected panics into 500 responses.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				response.InternalError(w, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders adds common hardening headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// CORS enforces an explicit allowlist and handles preflight requests.
func CORS(allowedOrigins []string) Middleware {
	allowAll := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				_, isAllowed := allowed[origin]
				if allowAll || isAllowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
					w.Header().Set("Access-Control-Max-Age", "600")
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// BodyLimit bounds request body size to reduce abuse risk.
func BodyLimit(maxBytes int64) Middleware {
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit applies a per-client-IP token bucket.
func RateLimit(limit rate.Limit, burst int) Middleware {
	l := &ipLimiter{
		limit:   limit,
		burst:   burst,
		clients: map[string]*rate.Limiter{},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := requestIP(r)
			if !l.allow(clientIP) {
				response.WriteError(w, http.StatusTooManyRequests, "Too many requests", "RATE_LIMITED")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ParseAllowedOrigins parses comma-separated origins from env.
func ParseAllowedOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"http://localhost:3000"}
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	if len(origins) == 0 {
		return []string{"http://localhost:3000"}
	}

	return origins
}

// ParseMaxBodyBytes parses max body bytes with fallback.
func ParseMaxBodyBytes(raw string, fallback int64) int64 {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}

	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v <= 0 {
		return fallback
	}

	return v
}

type ipLimiter struct {
	mu      sync.Mutex
	limit   rate.Limit
	burst   int
	clients map[string]*rate.Limiter
}

func (l *ipLimiter) allow(clientIP string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, ok := l.clients[clientIP]
	if !ok {
		limiter = rate.NewLimiter(l.limit, l.burst)
		l.clients[clientIP] = limiter
	}

	return limiter.Allow()
}

func requestIP(r *http.Request) string {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	if strings.TrimSpace(r.RemoteAddr) == "" {
		return "unknown"
	}

	return fmt.Sprintf("remote:%s", r.RemoteAddr)
}
