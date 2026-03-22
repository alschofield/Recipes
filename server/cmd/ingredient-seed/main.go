package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"recipes/pkg/storage/postgres"
)

type seedRow struct {
	CanonicalName       string
	Category            string
	NaturalSource       string
	FlavourMoleculeText string
	SourceCoverageText  string
	QualityScoreText    string
	AnalysisStatus      string
	AnalysisNotes       string
	Metadata            string
}

func main() {
	ctx := context.Background()
	pool := storage.Pool()
	defer pool.Close()

	seedPath := strings.TrimSpace(os.Getenv("CANONICAL_INGREDIENT_SEED"))
	if seedPath == "" {
		seedPath = filepath.FromSlash("server/lib/derived/canonical_ingredient_seed_v1.csv")
	}

	handle, err := os.Open(seedPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open canonical seed CSV (%s): %v\n", seedPath, err)
		os.Exit(1)
	}
	defer handle.Close()

	reader := csv.NewReader(handle)
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read canonical seed CSV: %v\n", err)
		os.Exit(1)
	}
	if len(rows) < 2 {
		fmt.Printf("canonical seed CSV has no data rows: %s\n", seedPath)
		return
	}

	headers := map[string]int{}
	for idx, col := range rows[0] {
		headers[strings.TrimSpace(col)] = idx
	}

	required := []string{
		"canonical_name",
		"category",
		"natural_source",
		"flavour_molecule_count",
		"source_coverage",
		"quality_score",
		"analysis_status",
		"analysis_notes",
		"metadata",
	}
	for _, col := range required {
		if _, ok := headers[col]; !ok {
			fmt.Fprintf(os.Stderr, "missing required column %q in %s\n", col, seedPath)
			os.Exit(1)
		}
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to begin transaction: %v\n", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	inserted := 0
	updated := 0

	for i := 1; i < len(rows); i++ {
		raw := rows[i]
		row := seedRow{
			CanonicalName:       get(raw, headers, "canonical_name"),
			Category:            get(raw, headers, "category"),
			NaturalSource:       get(raw, headers, "natural_source"),
			FlavourMoleculeText: get(raw, headers, "flavour_molecule_count"),
			SourceCoverageText:  get(raw, headers, "source_coverage"),
			QualityScoreText:    get(raw, headers, "quality_score"),
			AnalysisStatus:      get(raw, headers, "analysis_status"),
			AnalysisNotes:       get(raw, headers, "analysis_notes"),
			Metadata:            get(raw, headers, "metadata"),
		}

		if row.CanonicalName == "" {
			continue
		}

		flavourMolecule := parseIntOrNil(row.FlavourMoleculeText)
		sourceCoverage := parseIntOrZero(row.SourceCoverageText)
		qualityScore := parseFloatOrZero(row.QualityScoreText)

		var ingredientID string
		var insertedRow bool
		err := tx.QueryRow(ctx, `
			INSERT INTO ingredients (
				canonical_name,
				category,
				natural_source,
				flavour_molecule_count,
				source_coverage,
				quality_score,
				analysis_status,
				analysis_notes,
				last_analyzed_at,
				metadata
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),$9::jsonb)
			ON CONFLICT (canonical_name)
			DO UPDATE SET
				category = COALESCE(NULLIF(EXCLUDED.category, ''), ingredients.category),
				natural_source = COALESCE(NULLIF(EXCLUDED.natural_source, ''), ingredients.natural_source),
				flavour_molecule_count = COALESCE(EXCLUDED.flavour_molecule_count, ingredients.flavour_molecule_count),
				source_coverage = GREATEST(ingredients.source_coverage, EXCLUDED.source_coverage),
				quality_score = GREATEST(ingredients.quality_score, EXCLUDED.quality_score),
				analysis_status = EXCLUDED.analysis_status,
				analysis_notes = EXCLUDED.analysis_notes,
				last_analyzed_at = NOW(),
				metadata = COALESCE(ingredients.metadata, '{}'::jsonb) || COALESCE(EXCLUDED.metadata, '{}'::jsonb)
			RETURNING id, (xmax = 0)`,
			row.CanonicalName,
			row.Category,
			row.NaturalSource,
			flavourMolecule,
			sourceCoverage,
			qualityScore,
			defaultAnalysisStatus(row.AnalysisStatus),
			row.AnalysisNotes,
			defaultMetadata(row.Metadata),
		).Scan(&ingredientID, &insertedRow)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to upsert ingredient %q: %v\n", row.CanonicalName, err)
			continue
		}

		_, _ = tx.Exec(ctx, `
			INSERT INTO ingredient_aliases (ingredient_id, alias)
			VALUES ($1, $2)
			ON CONFLICT (alias) DO NOTHING`, ingredientID, row.CanonicalName)

		if insertedRow {
			inserted++
		} else {
			updated++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to commit ingredient seed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("ingredient canonical seed complete (%d inserted-ish, %d updated-ish)\n", inserted, updated)
}

func get(raw []string, headers map[string]int, key string) string {
	idx := headers[key]
	if idx < 0 || idx >= len(raw) {
		return ""
	}
	return strings.TrimSpace(raw[idx])
}

func parseIntOrNil(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return v
}

func parseIntOrZero(raw string) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	return v
}

func parseFloatOrZero(raw string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0
	}
	return v
}

func defaultAnalysisStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "enriched", "pending", "review_required":
		return status
	default:
		return "pending"
	}
}

func defaultMetadata(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "{}"
	}
	return raw
}
