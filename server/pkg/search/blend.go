package search

import (
	"encoding/json"
	"hash/fnv"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const (
	defaultBlendMinGenerated = 1
	defaultBlendMaxShare     = 0.40
)

func blendSearchResults(req SearchRequest, normalized []string, dbResults, generated []SearchResultItem) []SearchResultItem {
	for i := range dbResults {
		if dbResults[i].RankingReason == "" {
			dbResults[i].RankingReason = "db_ranked"
		}
	}

	if len(generated) == 0 {
		return dbResults
	}

	for i := range generated {
		if generated[i].RankingReason == "" {
			generated[i].RankingReason = "llm_fallback_generated"
		}
	}

	if len(dbResults) == 0 {
		return generated
	}

	minGenerated := parseBlendMinGenerated()
	maxShare := parseBlendMaxShare()
	allowedGenerated := allowedGeneratedCount(len(dbResults), len(generated), minGenerated, maxShare)
	if allowedGenerated <= 0 {
		return dbResults
	}
	if allowedGenerated < len(generated) {
		generated = generated[:allowedGenerated]
	}

	seed := blendSeed(buildSearchBlendSeedKey(normalized, req))
	shuffledGenerated := append([]SearchResultItem(nil), generated...)
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(shuffledGenerated), func(i, j int) {
		shuffledGenerated[i], shuffledGenerated[j] = shuffledGenerated[j], shuffledGenerated[i]
	})

	return interleaveWithSeed(dbResults, shuffledGenerated, seed)
}

func interleaveWithSeed(dbResults, generated []SearchResultItem, seed int64) []SearchResultItem {
	if len(generated) == 0 {
		return dbResults
	}

	if len(dbResults) == 0 {
		return generated
	}

	k := int(math.Ceil(float64(len(dbResults)) / float64(len(generated))))
	if k <= 0 {
		k = 1
	}

	startOffset := 0
	if k > 1 {
		startOffset = int(seed % int64(k))
	}

	out := make([]SearchResultItem, 0, len(dbResults)+len(generated))
	dbIndex := 0
	genIndex := 0
	dbSinceInsert := startOffset

	for dbIndex < len(dbResults) {
		out = append(out, dbResults[dbIndex])
		dbIndex++
		dbSinceInsert++

		if genIndex < len(generated) && dbSinceInsert >= k {
			item := generated[genIndex]
			item.RankingReason = "blend_interleave_generated"
			out = append(out, item)
			genIndex++
			dbSinceInsert = 0
		}
	}

	for genIndex < len(generated) {
		item := generated[genIndex]
		item.RankingReason = "blend_tail_generated"
		out = append(out, item)
		genIndex++
	}

	return out
}

func annotateBlendMetadata(items []SearchResultItem) {
	for i := range items {
		items[i].BlendSlot = i + 1
		if strings.TrimSpace(items[i].RankingReason) != "" {
			continue
		}
		if strings.EqualFold(items[i].Source, "llm") {
			items[i].RankingReason = "llm_fallback_generated"
		} else {
			items[i].RankingReason = "db_ranked"
		}
	}
}

func allowedGeneratedCount(dbCount, generatedCount, minGenerated int, maxShare float64) int {
	if generatedCount <= 0 {
		return 0
	}
	if dbCount <= 0 {
		return generatedCount
	}

	if maxShare <= 0 {
		if minGenerated > 0 {
			return min(minGenerated, generatedCount)
		}
		return 0
	}

	maxByShare := int(math.Floor((maxShare / (1.0 - maxShare)) * float64(dbCount)))
	if maxByShare < minGenerated {
		maxByShare = minGenerated
	}
	if maxByShare < 0 {
		maxByShare = 0
	}

	if maxByShare > generatedCount {
		maxByShare = generatedCount
	}

	return maxByShare
}

func parseBlendMinGenerated() int {
	raw := strings.TrimSpace(os.Getenv("SEARCH_BLEND_MIN_GENERATED"))
	if raw == "" {
		return defaultBlendMinGenerated
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return defaultBlendMinGenerated
	}
	return v
}

func parseBlendMaxShare() float64 {
	raw := strings.TrimSpace(os.Getenv("SEARCH_BLEND_MAX_GENERATED_SHARE"))
	if raw == "" {
		return defaultBlendMaxShare
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 || v >= 1 {
		return defaultBlendMaxShare
	}
	return v
}

func buildSearchBlendSeedKey(normalized []string, req SearchRequest) string {
	payload := struct {
		Ingredients []string       `json:"ingredients"`
		Mode        string         `json:"mode"`
		DBOnly      bool           `json:"dbOnly"`
		Complex     bool           `json:"complex"`
		Filters     *SearchFilters `json:"filters,omitempty"`
	}{
		Ingredients: append([]string(nil), normalized...),
		Mode:        req.Mode,
		DBOnly:      req.DBOnly,
		Complex:     req.Complex,
		Filters:     normalizeFiltersForCache(req.Filters),
	}

	raw, _ := json.Marshal(payload)
	return string(raw)
}

func blendSeed(key string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.TrimSpace(os.Getenv("SEARCH_BLEND_SEED"))))
	_, _ = h.Write([]byte("|"))
	_, _ = h.Write([]byte(key))
	return int64(h.Sum64())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
