package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"recipes/pkg/ingredients"
	"recipes/pkg/llmjudge"
	"recipes/pkg/middleware"
	"recipes/pkg/storage/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	fallbackTopMatchThreshold = 0.45
	fallbackMinResultCount    = 5
	defaultPromptVersion      = "v1"
	defaultPromptProfile      = "schema_first"
	defaultComplexProfile     = "safety_complex_first"
	defaultLLMTimeoutSeconds  = 180
	defaultRepairTimeout      = 90
	defaultLLMMaxTokens       = 1600
	strictPolicyNone          = "none"
	strictPolicyDegrade       = "degrade_inclusive"
)

type llmRuntimeConfig struct {
	APIKey               string
	BaseURL              string
	Model                string
	PromptProfileDefault string
	PromptProfileComplex string
	EnableSafetyRepair   bool
	DisableThinkingTag   bool
	MaxTokens            int
	TimeoutSeconds       int
	RepairTimeoutSeconds int
}

type llmFallbackRollout struct {
	Disabled      bool
	CanaryPercent int
}

var llmFallbackMetrics = struct {
	Requests      atomic.Uint64
	Success       atomic.Uint64
	RequestErrors atomic.Uint64
	TimeoutErrors atomic.Uint64
	SchemaErrors  atomic.Uint64
	RepairsTried  atomic.Uint64
	RepairsOK     atomic.Uint64
}{}

type llmRecipeResponse struct {
	Recipes []llmRecipe `json:"recipes"`
}

type llmRecipe struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Ingredients []llmRecipeIngredient `json:"ingredients"`
	Steps       []string              `json:"steps"`
	PrepMinutes int                   `json:"prepMinutes"`
	CookMinutes int                   `json:"cookMinutes"`
	Difficulty  string                `json:"difficulty"`
	Cuisine     string                `json:"cuisine"`
	DietaryTags []string              `json:"dietaryTags"`
	Servings    int                   `json:"servings"`
	SafetyNotes []string              `json:"safetyNotes"`
}

type llmRecipeIngredient struct {
	Name     string `json:"name"`
	Amount   string `json:"amount"`
	Optional bool   `json:"optional"`
}

func shouldTriggerLLMFallback(dbOnly bool, dbResults []SearchResultItem) bool {
	if dbOnly {
		return false
	}

	if len(dbResults) == 0 {
		return true
	}

	top := dbResults[0].MatchPercent
	return top < fallbackTopMatchThreshold && len(dbResults) < fallbackMinResultCount
}

func shouldExecuteLLMFallback(cacheKey string, dbOnly bool, dbResults []SearchResultItem) bool {
	if !shouldTriggerLLMFallback(dbOnly, dbResults) {
		return false
	}

	rollout := currentFallbackRollout()
	if rollout.Disabled {
		return false
	}

	if rollout.CanaryPercent >= 100 {
		return true
	}
	if rollout.CanaryPercent <= 0 {
		return false
	}

	if strings.TrimSpace(cacheKey) == "" {
		return true
	}

	return canaryBucket(cacheKey) <= rollout.CanaryPercent
}

func currentFallbackRollout() llmFallbackRollout {
	canary := 100
	rawCanary := strings.TrimSpace(os.Getenv("LLM_FALLBACK_CANARY_PERCENT"))
	if rawCanary != "" {
		if parsed, err := strconv.Atoi(rawCanary); err == nil {
			canary = parsed
		}
	}
	if canary < 0 {
		canary = 0
	}
	if canary > 100 {
		canary = 100
	}

	return llmFallbackRollout{
		Disabled:      parseEnvBool("LLM_FALLBACK_DISABLED", false),
		CanaryPercent: canary,
	}
}

func canaryBucket(key string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32()%100) + 1
}

func generateAndStoreFallbackRecipes(
	ctx context.Context,
	req SearchRequest,
	normalizedIngredients []string,
	aliases *aliasLookup,
) ([]SearchResultItem, error) {
	llmFallbackMetrics.Requests.Add(1)
	logLLMFallbackLifecycle(ctx, "triggered", "", nil)

	raw, model, promptProfile, err := callLLMForRecipes(ctx, req, normalizedIngredients)
	if err != nil {
		llmFallbackMetrics.RequestErrors.Add(1)
		if isTimeoutError(err) {
			llmFallbackMetrics.TimeoutErrors.Add(1)
		}
		logLLMFallbackMetrics("request_error", promptProfile, err)
		logLLMFallbackLifecycle(ctx, "skipped", promptProfile, err)
		return nil, err
	}

	parsed, err := parseAndValidateLLMResponse(raw)
	if err != nil {
		llmFallbackMetrics.SchemaErrors.Add(1)
		logLLMFallbackMetrics("schema_error", promptProfile, err)
		logLLMFallbackLifecycle(ctx, "skipped", promptProfile, err)
		return nil, err
	}

	pool := storage.Pool()
	inputSet := make(map[string]struct{}, len(normalizedIngredients))
	for _, ing := range normalizedIngredients {
		inputSet[ing] = struct{}{}
	}

	items := make([]SearchResultItem, 0, len(parsed.Recipes))
	for _, generated := range parsed.Recipes {
		item, err := persistGeneratedRecipeAndBuildResult(ctx, pool, generated, req, model, promptProfile, inputSet, aliases)
		if err != nil {
			logLLMFallbackLifecycle(ctx, "skipped", promptProfile, err)
			continue
		}
		items = append(items, item)
		logLLMFallbackLifecycle(ctx, "persisted", promptProfile, nil)
	}

	if len(items) > 0 {
		llmFallbackMetrics.Success.Add(1)
		logLLMFallbackMetrics("success", promptProfile, nil)
	} else {
		llmFallbackMetrics.SchemaErrors.Add(1)
		err := errors.New("no generated items persisted")
		logLLMFallbackMetrics("filtered_out", promptProfile, err)
		logLLMFallbackLifecycle(ctx, "skipped", promptProfile, err)
	}

	return items, nil
}

func callLLMForRecipes(ctx context.Context, req SearchRequest, normalizedIngredients []string) (string, string, string, error) {
	cfg, err := loadLLMRuntimeConfig()
	if err != nil {
		return "", "", "", err
	}

	promptProfile := choosePromptProfile(req, normalizedIngredients, cfg)
	prompt := buildLLMPrompt(req, normalizedIngredients, promptProfile, cfg)
	logLLMFallbackLifecycle(ctx, "provider_call", promptProfile, nil)

	raw, err := callLLMChat(ctx, cfg, prompt, 0.2, cfg.TimeoutSeconds)
	if err != nil {
		return "", "", "", err
	}

	if cfg.EnableSafetyRepair && isSafetySensitiveQuery(normalizedIngredients) {
		if _, parseErr := parseAndValidateLLMResponse(raw); parseErr != nil {
			llmFallbackMetrics.RepairsTried.Add(1)
			repairPrompt := buildLLMRepairPrompt(req, normalizedIngredients, raw, parseErr.Error(), promptProfile, cfg)
			logLLMFallbackLifecycle(ctx, "provider_call", promptProfile, nil)
			repaired, repairErr := callLLMChat(ctx, cfg, repairPrompt, 0.0, cfg.RepairTimeoutSeconds)
			if repairErr == nil {
				if _, repairedParseErr := parseAndValidateLLMResponse(repaired); repairedParseErr == nil {
					raw = repaired
					llmFallbackMetrics.RepairsOK.Add(1)
					logLLMFallbackLifecycle(ctx, "repaired", promptProfile, nil)
				}
			}
		}
	}

	return raw, cfg.Model, promptProfile, nil
}

func buildLLMPrompt(req SearchRequest, normalizedIngredients []string, promptProfile string, cfg *llmRuntimeConfig) string {
	filters := "none"
	if req.Filters != nil {
		b, _ := json.Marshal(req.Filters)
		filters = string(b)
	}

	instruction := "Generate 1-3 recipes"
	extraLines := ""
	if promptProfile == defaultComplexProfile {
		instruction = "Generate at least 2 recipes"
		extraLines = `

Case-specific requirements:
- Use staged prep + finishing flow with explicit sequencing in steps.
- Include multiple cooking techniques across the response (for example marinate, sear, roast, reduce, deglaze, or pickle).
- Keep safety notes concrete when allergens or high-risk ingredients are present.`
	}

	prefix := ""
	if cfg != nil && cfg.DisableThinkingTag {
		prefix = "/no_think\n"
	}

	return fmt.Sprintf(`%s%s for these normalized ingredients: %s.
Search mode: %s.
Filters: %s.
%s

Return JSON ONLY with this exact schema:
{
  "recipes": [
    {
      "name": "string",
      "description": "string",
      "ingredients": [
        { "name": "string", "amount": "string", "optional": false }
      ],
      "steps": ["string"],
      "prepMinutes": 30,
      "cookMinutes": 20,
      "difficulty": "easy|medium|hard",
      "cuisine": "string",
      "dietaryTags": ["string"],
      "servings": 2,
      "safetyNotes": ["string"]
    }
  ]
}

Do not include markdown, code fences, or any extra fields.`, prefix, instruction, strings.Join(normalizedIngredients, ", "), req.Mode, filters, extraLines)
}

func loadLLMRuntimeConfig() (*llmRuntimeConfig, error) {
	apiKey := strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("LLM_API_KEY is not configured")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("LLM_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}

	model := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	if model == "" {
		model = "qwen3:8b"
	}

	defaultProfile := strings.TrimSpace(os.Getenv("LLM_PROMPT_PROFILE_DEFAULT"))
	if defaultProfile == "" {
		defaultProfile = defaultPromptProfile
	}

	complexProfile := strings.TrimSpace(os.Getenv("LLM_PROMPT_PROFILE_COMPLEX"))
	if complexProfile == "" {
		complexProfile = defaultComplexProfile
	}

	timeoutSeconds := parseEnvInt("LLM_TIMEOUT_SECONDS", defaultLLMTimeoutSeconds)
	repairTimeoutSeconds := parseEnvInt("LLM_REPAIR_TIMEOUT_SECONDS", defaultRepairTimeout)
	maxTokens := parseEnvInt("LLM_MAX_TOKENS", defaultLLMMaxTokens)

	return &llmRuntimeConfig{
		APIKey:               apiKey,
		BaseURL:              baseURL,
		Model:                model,
		PromptProfileDefault: normalizePromptProfile(defaultProfile),
		PromptProfileComplex: normalizePromptProfile(complexProfile),
		EnableSafetyRepair:   parseEnvBool("LLM_ENABLE_SAFETY_REPAIR", true),
		DisableThinkingTag:   parseEnvBool("LLM_DISABLE_THINKING_TAG", true),
		MaxTokens:            maxTokens,
		TimeoutSeconds:       timeoutSeconds,
		RepairTimeoutSeconds: repairTimeoutSeconds,
	}, nil
}

func normalizePromptProfile(raw string) string {
	v := strings.TrimSpace(strings.ToLower(raw))
	if v == defaultComplexProfile {
		return defaultComplexProfile
	}
	return defaultPromptProfile
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

func choosePromptProfile(req SearchRequest, normalizedIngredients []string, cfg *llmRuntimeConfig) string {
	if req.Complex {
		return cfg.PromptProfileComplex
	}
	if isComplexLLMRequest(req, normalizedIngredients) {
		return cfg.PromptProfileComplex
	}
	return cfg.PromptProfileDefault
}

func isComplexLLMRequest(req SearchRequest, normalizedIngredients []string) bool {
	if len(normalizedIngredients) >= 10 {
		return true
	}
	if req.Filters == nil {
		return false
	}
	if req.Filters.Servings != nil && *req.Filters.Servings >= 4 {
		return true
	}
	if req.Filters.MaxPrepMinutes != nil && *req.Filters.MaxPrepMinutes >= 60 {
		return true
	}
	if req.Filters.MaxTotalMinutes != nil && *req.Filters.MaxTotalMinutes >= 90 {
		return true
	}
	return false
}

func isSafetySensitiveQuery(normalizedIngredients []string) bool {
	for _, ing := range normalizedIngredients {
		n := strings.ToLower(strings.TrimSpace(ing))
		if n == "" {
			continue
		}
		if strings.Contains(n, "chicken") || strings.Contains(n, "kidney bean") {
			return true
		}
		if strings.Contains(n, "peanut") || strings.Contains(n, "egg") || strings.Contains(n, "milk") || strings.Contains(n, "soy") || strings.Contains(n, "shellfish") {
			return true
		}
	}
	return false
}

func buildLLMRepairPrompt(req SearchRequest, normalizedIngredients []string, previousContent, parseError, promptProfile string, cfg *llmRuntimeConfig) string {
	filters := "none"
	if req.Filters != nil {
		b, _ := json.Marshal(req.Filters)
		filters = string(b)
	}
	prefix := ""
	if cfg != nil && cfg.DisableThinkingTag {
		prefix = "/no_think\n"
	}

	return fmt.Sprintf(`%sRewrite the previous model output so it is valid JSON and follows the exact schema contract.
Prompt profile: %s.
Normalized ingredients: %s.
Search mode: %s.
Filters: %s.
Previous parse error: %s.

Hard requirements:
- Output JSON only (no markdown)
- Top-level object must contain only "recipes"
- Every ingredient object must include exactly: name, amount, optional
- Keep outputs safe and practical

Previous output to repair:
%s`, prefix, promptProfile, strings.Join(normalizedIngredients, ", "), req.Mode, filters, parseError, previousContent)
}

func callLLMChat(ctx context.Context, cfg *llmRuntimeConfig, prompt string, temperature float64, timeoutSeconds int) (string, error) {
	execute := func(useJSONResponseFormat bool) (string, error) {
		body := map[string]interface{}{
			"model": cfg.Model,
			"messages": []map[string]string{
				{"role": "system", "content": "You generate safe, practical cooking recipes. Respond with valid JSON only."},
				{"role": "user", "content": prompt},
			},
			"temperature": temperature,
			"max_tokens":  cfg.MaxTokens,
		}
		if useJSONResponseFormat {
			body["response_format"] = map[string]string{"type": "json_object"}
		}
		if cfg.DisableThinkingTag {
			body["think"] = false
		}

		payload, err := json.Marshal(body)
		if err != nil {
			return "", err
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.BaseURL+"/chat/completions", bytes.NewReader(payload))
		if err != nil {
			return "", err
		}

		httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		httpReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", fmt.Errorf("LLM provider error: %d", resp.StatusCode)
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
			return "", errors.New("empty LLM response")
		}

		return providerResp.Choices[0].Message.Content, nil
	}

	content, err := execute(true)
	if err == nil {
		return content, nil
	}

	if strings.Contains(err.Error(), "empty LLM response") {
		return execute(false)
	}

	return "", err
}

func logLLMFallbackMetrics(event, profile string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	log.Printf(
		"llm_fallback event=%s profile=%s requests=%d success=%d request_errors=%d timeout_errors=%d schema_errors=%d repairs_tried=%d repairs_ok=%d err=%q",
		event,
		profile,
		llmFallbackMetrics.Requests.Load(),
		llmFallbackMetrics.Success.Load(),
		llmFallbackMetrics.RequestErrors.Load(),
		llmFallbackMetrics.TimeoutErrors.Load(),
		llmFallbackMetrics.SchemaErrors.Load(),
		llmFallbackMetrics.RepairsTried.Load(),
		llmFallbackMetrics.RepairsOK.Load(),
		errMsg,
	)
}

func logLLMFallbackLifecycle(ctx context.Context, event, profile string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	log.Printf(
		"llm_fallback_lifecycle event=%s request_id=%s profile=%s err=%q",
		event,
		middleware.RequestIDFromContext(ctx),
		profile,
		errMsg,
	)
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if os.IsTimeout(err) {
		return true
	}
	if strings.Contains(strings.ToLower(err.Error()), "timeout") {
		return true
	}
	return false
}

func parseAndValidateLLMResponse(raw string) (*llmRecipeResponse, error) {
	var parsed llmRecipeResponse
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&parsed); err != nil {
		return nil, err
	}
	if decoder.More() {
		return nil, errors.New("schema invalid: multiple JSON values")
	}

	if len(parsed.Recipes) == 0 {
		return nil, errors.New("schema invalid: recipes must contain at least one item")
	}

	for i, recipe := range parsed.Recipes {
		if strings.TrimSpace(recipe.Name) == "" {
			return nil, fmt.Errorf("schema invalid: recipes[%d].name is required", i)
		}
		if len(recipe.Ingredients) == 0 {
			return nil, fmt.Errorf("schema invalid: recipes[%d].ingredients must be non-empty", i)
		}
		if len(recipe.Steps) == 0 {
			return nil, fmt.Errorf("schema invalid: recipes[%d].steps must be non-empty", i)
		}
		if recipe.PrepMinutes < 0 || recipe.CookMinutes < 0 {
			return nil, fmt.Errorf("schema invalid: recipes[%d] prep/cook minutes must be >= 0", i)
		}
		difficulty := strings.ToLower(strings.TrimSpace(recipe.Difficulty))
		if difficulty != "easy" && difficulty != "medium" && difficulty != "hard" {
			return nil, fmt.Errorf("schema invalid: recipes[%d].difficulty invalid", i)
		}
		for j, ing := range recipe.Ingredients {
			if strings.TrimSpace(ing.Name) == "" {
				return nil, fmt.Errorf("schema invalid: recipes[%d].ingredients[%d].name is required", i, j)
			}
			if strings.TrimSpace(ing.Amount) == "" {
				return nil, fmt.Errorf("schema invalid: recipes[%d].ingredients[%d].amount is required", i, j)
			}
		}
	}

	return &parsed, nil
}

func persistGeneratedRecipeAndBuildResult(
	ctx context.Context,
	pool *pgxpool.Pool,
	generated llmRecipe,
	req SearchRequest,
	model string,
	promptProfile string,
	inputSet map[string]struct{},
	aliases *aliasLookup,
) (SearchResultItem, error) {
	type normalizedIngredient struct {
		canonical string
		amount    string
		optional  bool
	}

	recipeName := strings.TrimSpace(generated.Name)
	description := strings.TrimSpace(generated.Description)
	difficulty := strings.ToLower(strings.TrimSpace(generated.Difficulty))
	cuisine := strings.TrimSpace(generated.Cuisine)
	if cuisine == "" {
		cuisine = "global"
	}

	steps := make([]string, 0, len(generated.Steps))
	for _, step := range generated.Steps {
		s := strings.TrimSpace(step)
		if s != "" {
			steps = append(steps, s)
		}
	}
	stepsJSON, err := json.Marshal(steps)
	if err != nil {
		return SearchResultItem{}, err
	}

	prep := generated.PrepMinutes
	cook := generated.CookMinutes
	servings := generated.Servings
	if servings <= 0 {
		servings = 2
	}

	normalizedIngredients := []normalizedIngredient{}
	matched := []string{}
	missing := []string{}
	seenIng := map[string]struct{}{}

	for _, ing := range generated.Ingredients {
		canonical := normalizeRawIngredient(ing.Name)
		if canonical == "" {
			continue
		}

		if _, seen := seenIng[canonical]; seen {
			continue
		}
		seenIng[canonical] = struct{}{}

		normalizedIngredients = append(normalizedIngredients, normalizedIngredient{
			canonical: canonical,
			amount:    strings.TrimSpace(ing.Amount),
			optional:  ing.Optional,
		})

		if ing.Optional {
			continue
		}
		if _, ok := inputSet[canonical]; ok {
			matched = append(matched, canonical)
		} else {
			missing = append(missing, canonical)
		}
	}

	totalRequired := len(matched) + len(missing)
	matchPercent := 0.0
	if totalRequired > 0 {
		matchPercent = float64(len(matched)) / float64(totalRequired)
	}

	if req.Mode == "strict" && len(missing) > 0 {
		if strictGeneratedPolicy() == strictPolicyNone {
			return SearchResultItem{}, errors.New("generated recipe not strict compatible")
		}
	}

	totalMinutes := prep + cook
	grp := &recipeGroup{
		PrepMinutes: prepPtr(prep),
		CookMinutes: prepPtr(cook),
		Difficulty:  difficulty,
		Cuisine:     strPtr(cuisine),
		Servings:    servings,
		DietaryTags: generated.DietaryTags,
	}
	if !matchesFilters(grp, req.Filters, totalMinutes) {
		return SearchResultItem{}, errors.New("generated recipe filtered out")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return SearchResultItem{}, err
	}
	defer tx.Rollback(ctx)

	promptVersion := strings.TrimSpace(os.Getenv("LLM_PROMPT_VERSION"))
	if promptVersion == "" {
		promptVersion = fmt.Sprintf("%s-%s", defaultPromptVersion, promptProfile)
	}

	var recipeID string
	now := time.Now().UTC()
	err = tx.QueryRow(ctx, `
		INSERT INTO recipes (
			name, description, steps, source,
			generation_model, generation_timestamp, prompt_version,
			reviewable, quality_score,
			prep_minutes, cook_minutes, difficulty, cuisine,
			servings, dietary_tags, safety_notes
		) VALUES (
			$1, $2, $3, 'llm',
			$4, $5, $6,
			TRUE, 0.40,
			$7, $8, $9, $10,
			$11, $12, $13
		)
		RETURNING id`,
		recipeName,
		description,
		stepsJSON,
		model,
		now,
		promptVersion,
		prep,
		cook,
		difficulty,
		cuisine,
		servings,
		generated.DietaryTags,
		generated.SafetyNotes,
	).Scan(&recipeID)
	if err != nil {
		return SearchResultItem{}, err
	}

	for position, ing := range normalizedIngredients {
		resolved, err := ingredients.ResolveOrCreateForLLM(ctx, pool, ing.canonical)
		if err != nil || resolved.IngredientID == "" {
			continue
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO recipe_ingredients (recipe_id, ingredient_id, amount, unit, optional, position)
			VALUES ($1, $2, $3, '', $4, $5)`,
			recipeID,
			resolved.IngredientID,
			ing.amount,
			ing.optional,
			position,
		)
		if err != nil {
			return SearchResultItem{}, err
		}
	}

	if err := upsertRecipeQualityAnalysis(ctx, tx, recipeID, generated); err != nil {
		return SearchResultItem{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return SearchResultItem{}, err
	}

	substitutions := []Substitution{}
	if aliases != nil {
		for _, m := range missing {
			if subs, ok := aliases.canonicalToAliases[m]; ok && len(subs) > 0 {
				substitutions = append(substitutions, Substitution{Missing: m, Substitutes: subs})
			}
		}
	}

	return SearchResultItem{
		ID:                 recipeID,
		Name:               recipeName,
		Source:             "llm",
		RankingReason:      rankingReasonForGenerated(req.Mode, len(missing)),
		MatchPercent:       matchPercent,
		MatchedIngredients: matched,
		MissingIngredients: missing,
		Substitutions:      substitutions,
		PrepMinutes:        prepPtr(prep),
		CookMinutes:        prepPtr(cook),
		TotalMinutes:       totalMinutes,
		Difficulty:         difficulty,
		Cuisine:            strPtr(cuisine),
		Servings:           servings,
		DietaryTags:        generated.DietaryTags,
	}, nil
}

func strictGeneratedPolicy() string {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_STRICT_GENERATED_POLICY")))
	if raw == strictPolicyDegrade {
		return strictPolicyDegrade
	}
	return strictPolicyNone
}

func rankingReasonForGenerated(mode string, missingCount int) string {
	if strings.ToLower(strings.TrimSpace(mode)) == "strict" && missingCount > 0 && strictGeneratedPolicy() == strictPolicyDegrade {
		return "llm_fallback_generated_strict_degraded"
	}
	return "llm_fallback_generated"
}

func prepPtr(v int) *int {
	return &v
}

func strPtr(v string) *string {
	return &v
}

func upsertRecipeQualityAnalysis(ctx context.Context, tx pgx.Tx, recipeID string, generated llmRecipe) error {
	type quality struct {
		IngredientCoverage float64
		NutritionBalance   float64
		FlavourAlignment   float64
		Novelty            float64
		Overall            float64
	}

	var q quality
	err := tx.QueryRow(ctx, `
		WITH stats AS (
			SELECT
				COUNT(*)::numeric AS ingredient_count,
				COALESCE(AVG(LEAST(GREATEST(i.source_coverage, 0)::numeric / 5.0, 1.0)), 0) AS ingredient_coverage,
				COALESCE(AVG(LEAST(COALESCE(i.flavour_molecule_count, 0)::numeric / 200.0, 1.0)), 0) AS flavour_alignment,
				COALESCE(COUNT(DISTINCT NULLIF(i.category, ''))::numeric / NULLIF(COUNT(*), 0), 0) AS novelty
			FROM recipe_ingredients ri
			JOIN ingredients i ON i.id = ri.ingredient_id
			WHERE ri.recipe_id = $1
			  AND ri.optional = FALSE
		)
		SELECT
			ingredient_coverage,
			CASE
				WHEN ingredient_count >= 6 THEN 0.75
				WHEN ingredient_count >= 4 THEN 0.65
				WHEN ingredient_count >= 2 THEN 0.55
				ELSE 0.35
			END AS nutrition_balance,
			flavour_alignment,
			novelty,
			(
				ingredient_coverage * 0.40 +
				(CASE
					WHEN ingredient_count >= 6 THEN 0.75
					WHEN ingredient_count >= 4 THEN 0.65
					WHEN ingredient_count >= 2 THEN 0.55
					ELSE 0.35
				 END) * 0.25 +
				flavour_alignment * 0.20 +
				novelty * 0.15
			) AS overall
		FROM stats`, recipeID).Scan(
		&q.IngredientCoverage,
		&q.NutritionBalance,
		&q.FlavourAlignment,
		&q.Novelty,
		&q.Overall,
	)
	if err != nil {
		return err
	}

	judgeStatus := ""
	judgeNotes := ""
	if llmjudge.Enabled() {
		judgePayload := map[string]any{
			"name":        generated.Name,
			"description": generated.Description,
			"ingredients": generated.Ingredients,
			"steps":       generated.Steps,
			"safetyNotes": generated.SafetyNotes,
			"difficulty":  generated.Difficulty,
			"cuisine":     generated.Cuisine,
		}
		judgeResult, judgeErr := llmjudge.JudgeRecipeQuality(ctx, judgePayload)
		if judgeErr == nil && judgeResult != nil {
			judgeStatus = "computed"
			judgeNotesJSON, _ := json.Marshal(map[string]any{
				"judge": map[string]any{
					"overallScore":            judgeResult.OverallScore,
					"coherenceScore":          judgeResult.CoherenceScore,
					"safetyCompletenessScore": judgeResult.SafetyCompletenessScore,
					"techniqueScore":          judgeResult.TechniqueScore,
					"confidence":              judgeResult.Confidence,
					"notes":                   judgeResult.Notes,
				},
			})
			judgeNotes = string(judgeNotesJSON)
			q.Overall = calibratedOverallWithJudge(q.Overall, judgeResult.OverallScore, judgeResult.Confidence, llmjudge.MinConfidence())
		} else {
			judgeStatus = "failed"
			if judgeErr != nil {
				judgeNotes = judgeErr.Error()
			}
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO recipe_quality_analysis (
			recipe_id,
			ingredient_coverage_score,
			nutrition_balance_score,
			flavour_alignment_score,
			novelty_score,
			overall_score,
			status,
			notes,
			computed_at
		) VALUES ($1, $2, $3, $4, $5, $6, COALESCE(NULLIF($7, ''), 'computed'), NULLIF($8, ''), NOW())
		ON CONFLICT (recipe_id)
		DO UPDATE SET
			ingredient_coverage_score = EXCLUDED.ingredient_coverage_score,
			nutrition_balance_score = EXCLUDED.nutrition_balance_score,
			flavour_alignment_score = EXCLUDED.flavour_alignment_score,
			novelty_score = EXCLUDED.novelty_score,
			overall_score = EXCLUDED.overall_score,
			status = EXCLUDED.status,
			notes = EXCLUDED.notes,
			computed_at = NOW(),
			updated_at = NOW()`,
		recipeID,
		q.IngredientCoverage,
		q.NutritionBalance,
		q.FlavourAlignment,
		q.Novelty,
		q.Overall,
		judgeStatus,
		judgeNotes,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE recipes SET quality_score = $2 WHERE id = $1`, recipeID, q.Overall)
	return err
}

func calibratedOverallWithJudge(deterministicOverall, judgeOverall, judgeConfidence, minConfidence float64) float64 {
	if judgeConfidence < minConfidence {
		return deterministicOverall
	}

	return (deterministicOverall * 0.75) + (judgeOverall * 0.25)
}
