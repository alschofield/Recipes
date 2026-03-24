package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type loadRequest struct {
	Ingredients []string `json:"ingredients"`
	Mode        string   `json:"mode"`
	Complex     bool     `json:"complex,omitempty"`
	DBOnly      bool     `json:"dbOnly,omitempty"`
}

func main() {
	endpoint := flag.String("url", "http://localhost:8081/recipes/search", "recipes search endpoint")
	concurrency := flag.Int("concurrency", 8, "number of concurrent workers")
	requests := flag.Int("requests", 200, "total requests to send")
	timeout := flag.Duration("timeout", 20*time.Second, "http client timeout")
	scenario := flag.String("scenario", "fallback-heavy", "scenario: fallback-heavy or db-only")
	token := flag.String("token", "", "optional bearer token")
	flag.Parse()

	if *concurrency <= 0 || *requests <= 0 {
		fmt.Fprintln(os.Stderr, "concurrency and requests must be > 0")
		os.Exit(1)
	}

	payload, err := scenarioPayload(*scenario)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	client := &http.Client{Timeout: *timeout}
	latencies := make([]float64, *requests)
	var success atomic.Uint64
	var fail atomic.Uint64
	var non200 atomic.Uint64

	jobs := make(chan int, *requests)
	wg := sync.WaitGroup{}
	started := time.Now()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				lat := runOnce(client, *endpoint, body, strings.TrimSpace(*token), &success, &fail, &non200)
				latencies[idx] = lat
			}
		}()
	}

	for i := 0; i < *requests; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	elapsed := time.Since(started)
	throughput := float64(*requests) / elapsed.Seconds()

	sort.Float64s(latencies)
	p50 := percentile(latencies, 0.50)
	p95 := percentile(latencies, 0.95)
	p99 := percentile(latencies, 0.99)

	fmt.Printf("scenario=%s endpoint=%s\n", *scenario, *endpoint)
	fmt.Printf("requests=%d concurrency=%d duration_ms=%.2f throughput_rps=%.2f\n", *requests, *concurrency, elapsed.Seconds()*1000, throughput)
	fmt.Printf("success=%d failures=%d non200=%d\n", success.Load(), fail.Load(), non200.Load())
	fmt.Printf("latency_ms p50=%.2f p95=%.2f p99=%.2f\n", p50, p95, p99)
}

func runOnce(client *http.Client, endpoint string, body []byte, token string, success, fail, non200 *atomic.Uint64) float64 {
	started := time.Now()
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		fail.Add(1)
		return msSince(started)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		fail.Add(1)
		return msSince(started)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		success.Add(1)
	} else {
		non200.Add(1)
	}

	return msSince(started)
}

func scenarioPayload(scenario string) (loadRequest, error) {
	switch strings.ToLower(strings.TrimSpace(scenario)) {
	case "fallback-heavy":
		return loadRequest{
			Ingredients: []string{"zaatar", "sumac", "pomegranate molasses", "preserved lemon", "freekeh", "black lime", "barberry", "urfa pepper", "saffron"},
			Mode:        "strict",
			Complex:     true,
			DBOnly:      false,
		}, nil
	case "db-only":
		return loadRequest{
			Ingredients: []string{"chicken", "rice", "garlic", "onion"},
			Mode:        "inclusive",
			DBOnly:      true,
		}, nil
	default:
		return loadRequest{}, fmt.Errorf("unsupported scenario %q", scenario)
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	rank := p * float64(len(sorted)-1)
	low := int(math.Floor(rank))
	high := int(math.Ceil(rank))
	if low == high {
		return sorted[low]
	}
	weight := rank - float64(low)
	return sorted[low]*(1-weight) + sorted[high]*weight
}

func msSince(started time.Time) float64 {
	return float64(time.Since(started).Microseconds()) / 1000.0
}
