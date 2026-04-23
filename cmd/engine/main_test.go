package engine

import (
	"strings"
	"testing"
)

func TestParseQueryTermsAND(t *testing.T) {
	terms := parseQueryTerms("go AND search and engine")
	if len(terms) != 3 {
		t.Fatalf("expected 3 terms, got %d", len(terms))
	}
	if terms[0] != "go" || terms[1] != "search" || terms[2] != "engine" {
		t.Fatalf("unexpected terms: %#v", terms)
	}
}

func TestIntersectSortedDocIDs(t *testing.T) {
	a := []uint64{1, 2, 5, 9}
	b := []uint64{2, 3, 5, 8}

	got := intersectSortedDocIDs(a, b)
	if len(got) != 2 || got[0] != 2 || got[1] != 5 {
		t.Fatalf("unexpected intersection: %#v", got)
	}
}

func TestParseConfigRequiresInputInIngestMode(t *testing.T) {
	_, err := parseConfig([]string{"--mode", "ingest"})
	if err == nil {
		t.Fatalf("expected input validation error")
	}
	if !strings.Contains(err.Error(), "input or input-dir is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseConfigAllowsInputDir(t *testing.T) {
	config, err := parseConfig([]string{"--mode", "ingest", "--input-dir", "sample_inputs/varied"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.inputDir != "sample_inputs/varied" {
		t.Fatalf("unexpected inputDir: %q", config.inputDir)
	}
}

func TestParseConfigRejectsInvalidMode(t *testing.T) {
	_, err := parseConfig([]string{"--mode", "invalid", "--input", "file.json"})
	if err == nil {
		t.Fatalf("expected invalid mode error")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSearchModeRequiresQuery(t *testing.T) {
	err := Run([]string{"--mode", "search", "--dict", "data/dict-final.sst", "--post", "data/post-final.sst"})
	if err == nil {
		t.Fatalf("expected query validation error")
	}
	if !strings.Contains(err.Error(), "query is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
