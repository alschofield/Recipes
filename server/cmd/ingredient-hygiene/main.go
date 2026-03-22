package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"recipes/pkg/ingredients"
	"recipes/pkg/storage/postgres"
)

func main() {
	ctx := context.Background()
	pool := storage.Pool()

	rows, err := pool.Query(ctx, `SELECT id, canonical_name FROM ingredients ORDER BY canonical_name`)
	if err != nil {
		fmt.Printf("Failed to load ingredients: %v\n", err)
		return
	}
	defer rows.Close()

	type ing struct {
		id   string
		name string
	}
	items := []ing{}
	for rows.Next() {
		var i ing
		if err := rows.Scan(&i.id, &i.name); err == nil {
			items = append(items, i)
		}
	}

	type pair struct {
		left  string
		right string
		score float64
	}
	candidates := []pair{}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			score := similarity(items[i].name, items[j].name)
			if score >= 0.9 {
				candidates = append(candidates, pair{left: items[i].name, right: items[j].name, score: score})
			}
		}
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })

	fmt.Printf("Found %d potential duplicates\n", len(candidates))
	for _, c := range candidates {
		fmt.Printf("- %s <> %s (%.3f)\n", c.left, c.right, c.score)
	}
}

func similarity(a, b string) float64 {
	an := ingredients.NormalizeName(strings.ToLower(a))
	bn := ingredients.NormalizeName(strings.ToLower(b))
	if an == bn {
		return 1
	}
	maxLen := len(an)
	if len(bn) > maxLen {
		maxLen = len(bn)
	}
	if maxLen == 0 {
		return 1
	}
	d := levenshtein(an, bn)
	return 1 - float64(d)/float64(maxLen)
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	m := make([][]int, len(a)+1)
	for i := range m {
		m[i] = make([]int, len(b)+1)
		m[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		m[0][j] = j
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			a1 := m[i-1][j] + 1
			b1 := m[i][j-1] + 1
			c1 := m[i-1][j-1] + cost
			m[i][j] = a1
			if b1 < m[i][j] {
				m[i][j] = b1
			}
			if c1 < m[i][j] {
				m[i][j] = c1
			}
		}
	}
	return m[len(a)][len(b)]
}
