package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestShouldTriggerLLMFallback(t *testing.T) {
	tests := []struct {
		name    string
		dbOnly  bool
		results []SearchResultItem
		want    bool
	}{
		{
			name:   "dbOnly disables fallback",
			dbOnly: true,
			results: []SearchResultItem{
				{MatchPercent: 0.1},
			},
			want: false,
		},
		{
			name:    "no db results triggers fallback",
			dbOnly:  false,
			results: nil,
			want:    true,
		},
		{
			name: "low top score and few results triggers fallback",
			results: []SearchResultItem{
				{MatchPercent: 0.44},
				{MatchPercent: 0.3},
			},
			want: true,
		},
		{
			name: "enough results prevents fallback",
			results: []SearchResultItem{
				{MatchPercent: 0.44},
				{MatchPercent: 0.4},
				{MatchPercent: 0.39},
				{MatchPercent: 0.38},
				{MatchPercent: 0.37},
			},
			want: false,
		},
		{
			name: "high confidence top result prevents fallback",
			results: []SearchResultItem{
				{MatchPercent: 0.45},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldTriggerLLMFallback(tc.dbOnly, tc.results)
			if got != tc.want {
				t.Fatalf("shouldTriggerLLMFallback() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseAndValidateLLMResponse(t *testing.T) {
	t.Run("accepts valid schema", func(t *testing.T) {
		raw := `{
  "recipes": [
    {
      "name": "Tomato Pasta",
      "description": "Simple pasta",
      "ingredients": [
        { "name": "tomato", "amount": "2", "optional": false }
      ],
      "steps": ["Boil pasta", "Cook sauce"],
      "prepMinutes": 10,
      "cookMinutes": 15,
      "difficulty": "easy",
      "cuisine": "italian",
      "dietaryTags": ["vegetarian"],
      "servings": 2,
      "safetyNotes": ["Handle hot water carefully"]
    }
  ]
}`

		_, err := parseAndValidateLLMResponse(raw)
		if err != nil {
			t.Fatalf("expected valid payload, got error: %v", err)
		}
	})

	t.Run("rejects unknown fields", func(t *testing.T) {
		raw := `{
  "recipes": [
    {
      "name": "Tomato Pasta",
      "description": "Simple pasta",
      "ingredients": [
        { "name": "tomato", "amount": "2", "optional": false }
      ],
      "steps": ["Boil pasta"],
      "prepMinutes": 10,
      "cookMinutes": 15,
      "difficulty": "easy",
      "cuisine": "italian",
      "dietaryTags": [],
      "servings": 2,
      "safetyNotes": [],
      "extraField": "not allowed"
    }
  ]
}`

		_, err := parseAndValidateLLMResponse(raw)
		if err == nil {
			t.Fatal("expected unknown field validation error")
		}
	})

	t.Run("rejects invalid difficulty", func(t *testing.T) {
		raw := `{
  "recipes": [
    {
      "name": "Tomato Pasta",
      "description": "Simple pasta",
      "ingredients": [
        { "name": "tomato", "amount": "2", "optional": false }
      ],
      "steps": ["Boil pasta"],
      "prepMinutes": 10,
      "cookMinutes": 15,
      "difficulty": "expert",
      "cuisine": "italian",
      "dietaryTags": [],
      "servings": 2,
      "safetyNotes": []
    }
  ]
}`

		_, err := parseAndValidateLLMResponse(raw)
		if err == nil {
			t.Fatal("expected invalid difficulty validation error")
		}
	})
}

func TestCallLLMForRecipesWithMockedProvider(t *testing.T) {
	var seenUserPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var body struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed decoding request body: %v", err)
		}
		for _, m := range body.Messages {
			if m.Role == "user" {
				seenUserPrompt = m.Content
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
  "choices": [
    {
      "message": {
        "content": "{\"recipes\":[{\"name\":\"Mocked\",\"description\":\"Test\",\"ingredients\":[{\"name\":\"tomato\",\"amount\":\"1\",\"optional\":false}],\"steps\":[\"Cook\"],\"prepMinutes\":5,\"cookMinutes\":10,\"difficulty\":\"easy\",\"cuisine\":\"global\",\"dietaryTags\":[],\"servings\":2,\"safetyNotes\":[]}]}"
      }
    }
  ]
}`))
	}))
	defer server.Close()

	t.Setenv("LLM_API_KEY", "test-key")
	t.Setenv("LLM_BASE_URL", server.URL)
	t.Setenv("LLM_MODEL", "test-model")
	t.Setenv("LLM_DISABLE_THINKING_TAG", "true")

	content, model, profile, err := callLLMForRecipes(context.Background(), SearchRequest{Mode: "strict"}, []string{"tomato"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if model != "test-model" {
		t.Fatalf("expected model test-model, got %s", model)
	}
	if profile != defaultPromptProfile {
		t.Fatalf("expected prompt profile %s, got %s", defaultPromptProfile, profile)
	}
	if content == "" {
		t.Fatal("expected provider content")
	}
	if !strings.HasPrefix(seenUserPrompt, "/no_think\n") {
		t.Fatalf("expected user prompt to start with /no_think tag, got %q", seenUserPrompt)
	}
}

func TestCallLLMForRecipesRepairsInvalidSafetySensitivePayload(t *testing.T) {
	var reqCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&reqCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if current == 1 {
			_, _ = w.Write([]byte(`{
  "choices": [
    {
      "message": {
        "content": "{\"recipes\":[{\"name\":\"Broken\",\"description\":\"Invalid\",\"ingredients\":[{\" \":\"chicken\",\"amount\":\"1\",\"optional\":false}],\"steps\":[\"Cook\"],\"prepMinutes\":5,\"cookMinutes\":10,\"difficulty\":\"easy\",\"cuisine\":\"global\",\"dietaryTags\":[],\"servings\":2,\"safetyNotes\":[]}]}"
      }
    }
  ]
}`))
			return
		}
		_, _ = w.Write([]byte(`{
  "choices": [
    {
      "message": {
        "content": "{\"recipes\":[{\"name\":\"Fixed\",\"description\":\"Valid\",\"ingredients\":[{\"name\":\"chicken thigh\",\"amount\":\"1\",\"optional\":false}],\"steps\":[\"Cook to 74C/165F\"],\"prepMinutes\":5,\"cookMinutes\":10,\"difficulty\":\"easy\",\"cuisine\":\"global\",\"dietaryTags\":[],\"servings\":2,\"safetyNotes\":[\"Cook chicken to 74C/165F before serving\"]}]}"
      }
    }
  ]
}`))
	}))
	defer server.Close()

	t.Setenv("LLM_API_KEY", "test-key")
	t.Setenv("LLM_BASE_URL", server.URL)
	t.Setenv("LLM_MODEL", "test-model")
	t.Setenv("LLM_ENABLE_SAFETY_REPAIR", "true")
	t.Setenv("LLM_TIMEOUT_SECONDS", "10")
	t.Setenv("LLM_REPAIR_TIMEOUT_SECONDS", "10")

	content, _, _, err := callLLMForRecipes(context.Background(), SearchRequest{Mode: "inclusive"}, []string{"chicken thigh", "salt"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, parseErr := parseAndValidateLLMResponse(content); parseErr != nil {
		t.Fatalf("expected repaired payload to be valid, got %v", parseErr)
	}

	if got := atomic.LoadInt32(&reqCount); got != 2 {
		t.Fatalf("expected 2 provider calls (initial + repair), got %d", got)
	}
}

func TestChoosePromptProfileUsesComplexProfile(t *testing.T) {
	servings := 4
	cfg := &llmRuntimeConfig{
		PromptProfileDefault: defaultPromptProfile,
		PromptProfileComplex: defaultComplexProfile,
	}
	profile := choosePromptProfile(
		SearchRequest{Filters: &SearchFilters{Servings: &servings}},
		[]string{"a", "b", "c"},
		cfg,
	)
	if profile != defaultComplexProfile {
		t.Fatalf("expected %s, got %s", defaultComplexProfile, profile)
	}
}

func TestChoosePromptProfileUsesComplexProfileForLargeIngredientSet(t *testing.T) {
	cfg := &llmRuntimeConfig{
		PromptProfileDefault: defaultPromptProfile,
		PromptProfileComplex: defaultComplexProfile,
	}
	ingredients := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	profile := choosePromptProfile(SearchRequest{}, ingredients, cfg)
	if profile != defaultComplexProfile {
		t.Fatalf("expected %s, got %s", defaultComplexProfile, profile)
	}
}

func TestChoosePromptProfileHonorsComplexHint(t *testing.T) {
	cfg := &llmRuntimeConfig{
		PromptProfileDefault: defaultPromptProfile,
		PromptProfileComplex: defaultComplexProfile,
	}
	profile := choosePromptProfile(
		SearchRequest{Complex: true},
		[]string{"egg", "onion"},
		cfg,
	)
	if profile != defaultComplexProfile {
		t.Fatalf("expected %s, got %s", defaultComplexProfile, profile)
	}
}

func TestHandleLLMHealthConfigured(t *testing.T) {
	t.Setenv("LLM_API_KEY", "local-not-used")
	t.Setenv("LLM_BASE_URL", "http://ollama:11434/v1")
	t.Setenv("LLM_MODEL", "qwen3:8b")
	t.Setenv("LLM_PROMPT_PROFILE_DEFAULT", "schema_first")
	t.Setenv("LLM_PROMPT_PROFILE_COMPLEX", "safety_complex_first")
	t.Setenv("LLM_ENABLE_SAFETY_REPAIR", "true")
	t.Setenv("LLM_DISABLE_THINKING_TAG", "true")
	t.Setenv("LLM_TIMEOUT_SECONDS", "90")
	t.Setenv("LLM_REPAIR_TIMEOUT_SECONDS", "45")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/recipes/health/llm", nil)
	HandleLLMHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", body["status"])
	}
	runtime, _ := body["runtime"].(map[string]interface{})
	if runtime["configured"] != true {
		t.Fatalf("expected configured true, got %v", runtime["configured"])
	}
	if runtime["disableThinkingTag"] != true {
		t.Fatalf("expected disableThinkingTag true, got %v", runtime["disableThinkingTag"])
	}
}

func TestHandleLLMHealthDegradedWhenMissingKey(t *testing.T) {
	t.Setenv("LLM_API_KEY", "")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/recipes/health/llm", nil)
	HandleLLMHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body["status"] != "degraded" {
		t.Fatalf("expected status degraded, got %v", body["status"])
	}
}
