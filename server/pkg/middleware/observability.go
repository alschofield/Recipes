package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const requestIDContextKey contextKey = "recipes.request_id"

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// RequestID adds a request id header/context for observability.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext returns the active request id.
func RequestIDFromContext(ctx context.Context) string {
	value, ok := ctx.Value(requestIDContextKey).(string)
	if !ok {
		return ""
	}
	return value
}

// RequestLogger logs requests in JSON lines format.
func RequestLogger(serviceName string) Middleware {
	if serviceName == "" {
		serviceName = "recipes"
	}

	logger := log.New(os.Stdout, "", 0)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now().UTC()
			recorder := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(recorder, r)

			payload := map[string]any{
				"timestamp":   time.Now().UTC().Format(time.RFC3339Nano),
				"service":     serviceName,
				"requestId":   RequestIDFromContext(r.Context()),
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      recorder.status,
				"durationMs":  time.Since(startedAt).Milliseconds(),
				"remoteAddr":  r.RemoteAddr,
				"userAgent":   r.UserAgent(),
				"contentType": r.Header.Get("Content-Type"),
			}

			encoded, err := json.Marshal(payload)
			if err == nil {
				logger.Println(string(encoded))
			}
		})
	}
}

type MetricsCollector struct {
	serviceName      string
	totalRequests    atomic.Uint64
	inFlightRequests atomic.Int64
	errorRequests    atomic.Uint64
	totalLatencyMs   atomic.Uint64
}

func NewMetricsCollector(serviceName string) *MetricsCollector {
	if serviceName == "" {
		serviceName = "recipes"
	}
	return &MetricsCollector{serviceName: serviceName}
}

func (m *MetricsCollector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now().UTC()
		m.inFlightRequests.Add(1)
		defer m.inFlightRequests.Add(-1)

		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		m.totalRequests.Add(1)
		m.totalLatencyMs.Add(uint64(time.Since(startedAt).Milliseconds()))
		if recorder.status >= 500 {
			m.errorRequests.Add(1)
		}
	})
}

func (m *MetricsCollector) Handler(w http.ResponseWriter, r *http.Request) {
	payload := m.Snapshot()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func (m *MetricsCollector) Snapshot() map[string]any {
	total := m.totalRequests.Load()
	latency := m.totalLatencyMs.Load()
	average := 0.0
	if total > 0 {
		average = float64(latency) / float64(total)
	}

	return map[string]any{
		"service":           m.serviceName,
		"requestsTotal":     total,
		"requestsInFlight":  m.inFlightRequests.Load(),
		"errors5xxTotal":    m.errorRequests.Load(),
		"avgLatencyMs":      average,
		"observedAt":        time.Now().UTC().Format(time.RFC3339Nano),
		"errorRateEstimate": safeRate(m.errorRequests.Load(), total),
	}
}

func safeRate(numerator, denominator uint64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

// ErrorNotifier sends webhook alerts for 5xx responses.
func ErrorNotifier(webhookURL, serviceName string) Middleware {
	if webhookURL == "" {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(recorder, r)

			if recorder.status < 500 {
				return
			}

			payload := map[string]any{
				"service":   serviceName,
				"requestId": RequestIDFromContext(r.Context()),
				"method":    r.Method,
				"path":      r.URL.Path,
				"status":    recorder.status,
				"time":      time.Now().UTC().Format(time.RFC3339Nano),
			}

			go sendAlert(webhookURL, payload)
		})
	}
}

func sendAlert(webhookURL string, payload map[string]any) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return
	}

	request, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(encoded))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 3 * time.Second}
	_, _ = client.Do(request)
}
