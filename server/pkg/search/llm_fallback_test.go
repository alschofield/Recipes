package search

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
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

	content, model, err := callLLMForRecipes(context.Background(), SearchRequest{Mode: "strict"}, []string{"tomato"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if model != "test-model" {
		t.Fatalf("expected model test-model, got %s", model)
	}
	if content == "" {
		t.Fatal("expected provider content")
	}
}
