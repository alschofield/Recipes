package llmjudge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultJudgeModel          = "mistral:latest"
	defaultJudgeTimeoutSeconds = 45
	defaultJudgeMaxTokens      = 500
)

type IngredientMetadataResult struct {
	Category         string   `json:"category"`
	AliasSuggestions []string `json:"aliasSuggestions"`
	AllergenHints    []string `json:"allergenHints"`
	RiskHints        []string `json:"riskHints"`
	Confidence       float64  `json:"confidence"`
	Evidence         string   `json:"evidence"`
}

type RecipeQualityResult struct {
	OverallScore            float64 `json:"overallScore"`
	CoherenceScore          float64 `json:"coherenceScore"`
	SafetyCompletenessScore float64 `json:"safetyCompletenessScore"`
	TechniqueScore          float64 `json:"techniqueScore"`
	Confidence              float64 `json:"confidence"`
	Notes                   string  `json:"notes"`
}

type JudgeMetrics struct {
	Requests  uint64 `json:"requests"`
	Success   uint64 `json:"success"`
	Failures  uint64 `json:"failures"`
	LastError string `json:"lastError"`
}

type judgeConfig struct {
	enabled           bool
	baseURL           string
	apiKey            string
	model             string
	disableThinking   bool
	timeoutSeconds    int
	maxTokens         int
	confidenceMinimum float64
}

var metrics = struct {
	requests  atomic.Uint64
	success   atomic.Uint64
	failures  atomic.Uint64
	lastError atomic.Value
}{}

func init() {
	metrics.lastError.Store("")
}

func Snapshot() JudgeMetrics {
	lastError, _ := metrics.lastError.Load().(string)
	return JudgeMetrics{
		Requests:  metrics.requests.Load(),
		Success:   metrics.success.Load(),
		Failures:  metrics.failures.Load(),
		LastError: lastError,
	}
}

func Enabled() bool {
	return parseEnvBool("LLM_JUDGE_ENABLED", false)
}

func MinConfidence() float64 {
	raw := strings.TrimSpace(os.Getenv("LLM_JUDGE_MIN_CONFIDENCE"))
	if raw == "" {
		return 0.65
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0.65
	}
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func JudgeIngredientMetadata(ctx context.Context, canonicalName string) (*IngredientMetadataResult, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.enabled {
		return nil, errors.New("judge disabled")
	}

	prompt := fmt.Sprintf(
		"Return JSON only for ingredient metadata. Ingredient: %s. Include category, aliasSuggestions, allergenHints, riskHints, confidence (0-1), evidence.",
		canonicalName,
	)

	content, err := chatJSON(ctx, cfg, prompt)
	if err != nil {
		recordFailure(err)
		return nil, err
	}

	var result IngredientMetadataResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		recordFailure(err)
		return nil, err
	}
	result.Category = strings.TrimSpace(result.Category)
	result.Evidence = strings.TrimSpace(result.Evidence)
	if result.Category == "" || result.Evidence == "" {
		err := errors.New("judge returned incomplete ingredient metadata")
		recordFailure(err)
		return nil, err
	}
	if result.Confidence < 0 || result.Confidence > 1 {
		err := errors.New("judge confidence out of range")
		recordFailure(err)
		return nil, err
	}
	metrics.success.Add(1)
	return &result, nil
}

func JudgeRecipeQuality(ctx context.Context, recipePayload map[string]any) (*RecipeQualityResult, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.enabled {
		return nil, errors.New("judge disabled")
	}

	encoded, _ := json.Marshal(recipePayload)
	prompt := "Return JSON only with overallScore, coherenceScore, safetyCompletenessScore, techniqueScore, confidence, notes for this recipe payload: " + string(encoded)

	content, err := chatJSON(ctx, cfg, prompt)
	if err != nil {
		recordFailure(err)
		return nil, err
	}

	var result RecipeQualityResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		recordFailure(err)
		return nil, err
	}
	if result.Confidence < 0 || result.Confidence > 1 {
		err := errors.New("judge confidence out of range")
		recordFailure(err)
		return nil, err
	}
	metrics.success.Add(1)
	return &result, nil
}

func loadConfig() (*judgeConfig, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("LLM_JUDGE_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = strings.TrimRight(strings.TrimSpace(os.Getenv("LLM_BASE_URL")), "/")
	}
	if baseURL == "" {
		return nil, errors.New("LLM_JUDGE_BASE_URL/LLM_BASE_URL not configured")
	}

	apiKey := strings.TrimSpace(os.Getenv("LLM_JUDGE_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	}
	if apiKey == "" {
		return nil, errors.New("LLM_JUDGE_API_KEY/LLM_API_KEY not configured")
	}

	model := strings.TrimSpace(os.Getenv("LLM_JUDGE_MODEL"))
	if model == "" {
		model = defaultJudgeModel
	}

	return &judgeConfig{
		enabled:           parseEnvBool("LLM_JUDGE_ENABLED", false),
		baseURL:           baseURL,
		apiKey:            apiKey,
		model:             model,
		disableThinking:   parseEnvBool("LLM_JUDGE_DISABLE_THINKING_TAG", true),
		timeoutSeconds:    parseEnvInt("LLM_JUDGE_TIMEOUT_SECONDS", defaultJudgeTimeoutSeconds),
		maxTokens:         parseEnvInt("LLM_JUDGE_MAX_TOKENS", defaultJudgeMaxTokens),
		confidenceMinimum: MinConfidence(),
	}, nil
}

func chatJSON(ctx context.Context, cfg *judgeConfig, prompt string) (string, error) {
	metrics.requests.Add(1)

	if cfg.disableThinking {
		prompt = "/no_think\n" + prompt
	}

	body := map[string]any{
		"model": cfg.model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a strict JSON judge. Return valid JSON only."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0,
		"max_tokens":  cfg.maxTokens,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}
	if cfg.disableThinking {
		body["think"] = false
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(cfg.timeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("judge provider error: %d", resp.StatusCode)
	}

	var providerResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&providerResp); err != nil {
		return "", err
	}
	if len(providerResp.Choices) == 0 || strings.TrimSpace(providerResp.Choices[0].Message.Content) == "" {
		return "", errors.New("empty judge response")
	}

	return providerResp.Choices[0].Message.Content, nil
}

func recordFailure(err error) {
	metrics.failures.Add(1)
	metrics.lastError.Store(err.Error())
}

func parseEnvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func parseEnvBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return v
}
