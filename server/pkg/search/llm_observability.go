package search

import (
	"net/http"
	"os"
	"time"

	"recipes/pkg/llmjudge"
	"recipes/pkg/response"
)

type llmFallbackMetricsSnapshot struct {
	Requests      uint64 `json:"requests"`
	Success       uint64 `json:"success"`
	RequestErrors uint64 `json:"requestErrors"`
	SchemaErrors  uint64 `json:"schemaErrors"`
	RepairsTried  uint64 `json:"repairsTried"`
	RepairsOK     uint64 `json:"repairsOK"`
}

func getLLMFallbackMetricsSnapshot() llmFallbackMetricsSnapshot {
	return llmFallbackMetricsSnapshot{
		Requests:      llmFallbackMetrics.Requests.Load(),
		Success:       llmFallbackMetrics.Success.Load(),
		RequestErrors: llmFallbackMetrics.RequestErrors.Load(),
		SchemaErrors:  llmFallbackMetrics.SchemaErrors.Load(),
		RepairsTried:  llmFallbackMetrics.RepairsTried.Load(),
		RepairsOK:     llmFallbackMetrics.RepairsOK.Load(),
	}
}

// HandleLLMHealth returns fallback runtime config + counters.
func HandleLLMHealth(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadLLMRuntimeConfig()

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
			"configured":           configured,
			"baseUrl":              baseURL,
			"model":                model,
			"promptProfileDefault": profileDefault,
			"promptProfileComplex": profileComplex,
			"enableSafetyRepair":   enableRepair,
			"disableThinkingTag":   disableThinkingTag,
			"maxTokens":            maxTokens,
			"timeoutSeconds":       timeoutSeconds,
			"repairTimeoutSeconds": repairTimeoutSeconds,
		},
		"metrics": getLLMFallbackMetricsSnapshot(),
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
	response.WriteJSON(w, http.StatusOK, map[string]any{
		"service":            "recipes-llm-fallback",
		"observedAt":         time.Now().UTC().Format(time.RFC3339Nano),
		"llmFallbackMetrics": getLLMFallbackMetricsSnapshot(),
		"judgeMetrics":       llmjudge.Snapshot(),
		"judgeEnabled":       llmjudge.Enabled(),
		"judgeMinConfidence": llmjudge.MinConfidence(),
	})
}
