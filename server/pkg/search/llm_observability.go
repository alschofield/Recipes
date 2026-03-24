package search

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"recipes/pkg/llmjudge"
	"recipes/pkg/response"
)

type llmFallbackMetricsSnapshot struct {
	Requests      uint64 `json:"requests"`
	Success       uint64 `json:"success"`
	RequestErrors uint64 `json:"requestErrors"`
	TimeoutErrors uint64 `json:"timeoutErrors"`
	SchemaErrors  uint64 `json:"schemaErrors"`
	RepairsTried  uint64 `json:"repairsTried"`
	RepairsOK     uint64 `json:"repairsOK"`
}

type llmAlertThresholds struct {
	TimeoutRate    float64 `json:"timeoutRate"`
	SchemaRate     float64 `json:"schemaErrorRate"`
	RepairFailRate float64 `json:"repairFailRate"`
}

type llmAlertRates struct {
	TimeoutRate    float64 `json:"timeoutRate"`
	SchemaRate     float64 `json:"schemaErrorRate"`
	RepairFailRate float64 `json:"repairFailRate"`
}

type llmAlertStatus struct {
	Status     string             `json:"status"`
	Thresholds llmAlertThresholds `json:"thresholds"`
	Rates      llmAlertRates      `json:"rates"`
	Breaches   map[string]bool    `json:"breaches"`
}

type LabeledMetricSample struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Value  uint64            `json:"value"`
	Labels map[string]string `json:"labels"`
}

func getLLMFallbackMetricsSnapshot() llmFallbackMetricsSnapshot {
	return llmFallbackMetricsSnapshot{
		Requests:      llmFallbackMetrics.Requests.Load(),
		Success:       llmFallbackMetrics.Success.Load(),
		RequestErrors: llmFallbackMetrics.RequestErrors.Load(),
		TimeoutErrors: llmFallbackMetrics.TimeoutErrors.Load(),
		SchemaErrors:  llmFallbackMetrics.SchemaErrors.Load(),
		RepairsTried:  llmFallbackMetrics.RepairsTried.Load(),
		RepairsOK:     llmFallbackMetrics.RepairsOK.Load(),
	}
}

func llmFallbackAlertSnapshot(metrics llmFallbackMetricsSnapshot) llmAlertStatus {
	thresholds := llmAlertThresholds{
		TimeoutRate:    parseEnvFloatWithFallback("LLM_ALERT_TIMEOUT_RATE_THRESHOLD", 0.20),
		SchemaRate:     parseEnvFloatWithFallback("LLM_ALERT_SCHEMA_ERROR_RATE_THRESHOLD", 0.15),
		RepairFailRate: parseEnvFloatWithFallback("LLM_ALERT_REPAIR_FAIL_RATE_THRESHOLD", 0.50),
	}

	repairFailures := uint64(0)
	if metrics.RepairsTried > metrics.RepairsOK {
		repairFailures = metrics.RepairsTried - metrics.RepairsOK
	}

	rates := llmAlertRates{
		TimeoutRate:    safeRate(metrics.TimeoutErrors, metrics.Requests),
		SchemaRate:     safeRate(metrics.SchemaErrors, metrics.Requests),
		RepairFailRate: safeRate(repairFailures, metrics.RepairsTried),
	}

	breaches := map[string]bool{
		"timeoutRate":     rates.TimeoutRate > thresholds.TimeoutRate,
		"schemaErrorRate": rates.SchemaRate > thresholds.SchemaRate,
		"repairFailRate":  rates.RepairFailRate > thresholds.RepairFailRate,
	}

	status := "ok"
	if breaches["timeoutRate"] || breaches["schemaErrorRate"] || breaches["repairFailRate"] {
		status = "degraded"
	}

	return llmAlertStatus{
		Status:     status,
		Thresholds: thresholds,
		Rates:      rates,
		Breaches:   breaches,
	}
}

func parseEnvFloatWithFallback(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		return fallback
	}

	return v
}

func safeRate(numerator, denominator uint64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func LLMFallbackMetricSamples(service string) []LabeledMetricSample {
	if service == "" {
		service = "recipes-server"
	}

	s := getLLMFallbackMetricsSnapshot()
	labels := map[string]string{
		"service":   service,
		"subsystem": "llm_fallback",
	}

	return []LabeledMetricSample{
		{
			Name:   "recipes_llm_fallback_requests_total",
			Type:   "counter",
			Value:  s.Requests,
			Labels: labels,
		},
		{
			Name:   "recipes_llm_fallback_success_total",
			Type:   "counter",
			Value:  s.Success,
			Labels: labels,
		},
		{
			Name:   "recipes_llm_fallback_request_errors_total",
			Type:   "counter",
			Value:  s.RequestErrors,
			Labels: labels,
		},
		{
			Name:   "recipes_llm_fallback_timeout_errors_total",
			Type:   "counter",
			Value:  s.TimeoutErrors,
			Labels: labels,
		},
		{
			Name:   "recipes_llm_fallback_schema_errors_total",
			Type:   "counter",
			Value:  s.SchemaErrors,
			Labels: labels,
		},
		{
			Name:   "recipes_llm_fallback_repairs_tried_total",
			Type:   "counter",
			Value:  s.RepairsTried,
			Labels: labels,
		},
		{
			Name:   "recipes_llm_fallback_repairs_ok_total",
			Type:   "counter",
			Value:  s.RepairsOK,
			Labels: labels,
		},
	}
}

// HandleLLMHealth returns fallback runtime config + counters.
func HandleLLMHealth(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadLLMRuntimeConfig()
	fallbackMetrics := getLLMFallbackMetricsSnapshot()
	alertStatus := llmFallbackAlertSnapshot(fallbackMetrics)
	rollout := currentFallbackRollout()

	status := "ok"
	errorMessage := ""
	configured := true
	baseURL := ""
	model := ""
	profileDefault := ""
	profileComplex := ""
	enableRepair := false
	disableThinkingTag := false
	maxTokens := 0
	timeoutSeconds := 0
	repairTimeoutSeconds := 0
	judgeEnabled := llmjudge.Enabled()
	judgeModel := os.Getenv("LLM_JUDGE_MODEL")
	if judgeModel == "" {
		judgeModel = "mistral:latest"
	}

	if err != nil {
		status = "degraded"
		errorMessage = err.Error()
		configured = false
	} else {
		baseURL = cfg.BaseURL
		model = cfg.Model
		profileDefault = cfg.PromptProfileDefault
		profileComplex = cfg.PromptProfileComplex
		enableRepair = cfg.EnableSafetyRepair
		disableThinkingTag = cfg.DisableThinkingTag
		maxTokens = cfg.MaxTokens
		timeoutSeconds = cfg.TimeoutSeconds
		repairTimeoutSeconds = cfg.RepairTimeoutSeconds
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"status":    status,
		"service":   "recipes-llm-fallback",
		"checkedAt": time.Now().UTC().Format(time.RFC3339),
		"runtime": map[string]any{
			"configured":            configured,
			"baseUrl":               baseURL,
			"model":                 model,
			"promptProfileDefault":  profileDefault,
			"promptProfileComplex":  profileComplex,
			"enableSafetyRepair":    enableRepair,
			"disableThinkingTag":    disableThinkingTag,
			"maxTokens":             maxTokens,
			"timeoutSeconds":        timeoutSeconds,
			"repairTimeoutSeconds":  repairTimeoutSeconds,
			"fallbackDisabled":      rollout.Disabled,
			"fallbackCanaryPercent": rollout.CanaryPercent,
		},
		"metrics": fallbackMetrics,
		"alerts":  alertStatus,
		"judge": map[string]any{
			"enabled":       judgeEnabled,
			"model":         judgeModel,
			"minConfidence": llmjudge.MinConfidence(),
			"metrics":       llmjudge.Snapshot(),
		},
		"error": errorMessage,
	})
}

// HandleLLMMetrics returns machine-friendly LLM/judge counters.
func HandleLLMMetrics(w http.ResponseWriter, r *http.Request) {
	fallbackMetrics := getLLMFallbackMetricsSnapshot()
	alertStatus := llmFallbackAlertSnapshot(fallbackMetrics)
	rollout := currentFallbackRollout()

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"service":            "recipes-llm-fallback",
		"observedAt":         time.Now().UTC().Format(time.RFC3339Nano),
		"llmFallbackMetrics": fallbackMetrics,
		"alerts":             alertStatus,
		"rollout":            rollout,
		"judgeMetrics":       llmjudge.Snapshot(),
		"judgeEnabled":       llmjudge.Enabled(),
		"judgeMinConfidence": llmjudge.MinConfidence(),
	})
}
