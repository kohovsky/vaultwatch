package monitor

import (
	"testing"
)

func TestFilter_NoRules_ReturnsAll(t *testing.T) {
	paths := []string{"secret/a", "secret/b", "kv/x"}
	got := Filter(paths, FilterConfig{})
	if len(got) != len(paths) {
		t.Fatalf("expected %d paths, got %d", len(paths), len(got))
	}
}

func TestFilter_IncludePrefix(t *testing.T) {
	paths := []string{"secret/a", "kv/b", "secret/c"}
	got := Filter(paths, FilterConfig{IncludePrefixes: []string{"secret/"}})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(got), got)
	}
	for _, p := range got {
		if p != "secret/a" && p != "secret/c" {
			t.Errorf("unexpected path %q", p)
		}
	}
}

func TestFilter_ExcludePrefix(t *testing.T) {
	paths := []string{"secret/a", "kv/b", "secret/c"}
	got := Filter(paths, FilterConfig{ExcludePrefixes: []string{"kv/"}})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(got), got)
	}
	for _, p := range got {
		if p == "kv/b" {
			t.Errorf("excluded path %q still present", p)
		}
	}
}

func TestFilter_IncludeAndExclude(t *testing.T) {
	paths := []string{"secret/prod/db", "secret/staging/db", "kv/other"}
	cfg := FilterConfig{
		IncludePrefixes: []string{"secret/"},
		ExcludePrefixes: []string{"secret/staging/"},
	}
	got := Filter(paths, cfg)
	if len(got) != 1 || got[0] != "secret/prod/db" {
		t.Fatalf("expected [secret/prod/db], got %v", got)
	}
}

func TestFilter_EmptyInput(t *testing.T) {
	got := Filter([]string{}, FilterConfig{IncludePrefixes: []string{"secret/"}})
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestFilter_MultipleIncludePrefixes(t *testing.T) {
	paths := []string{"secret/a", "kv/b", "pki/c", "other/d"}
	got := Filter(paths, FilterConfig{IncludePrefixes: []string{"secret/", "pki/"}})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(got), got)
	}
}
