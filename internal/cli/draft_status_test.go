package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDraftWithMockProviderWritesProposed(t *testing.T) {
	root := seedTwo(t)
	mustCLI(t, root, "draft", "--pr", "42", "--provider", "mock", "--yes")

	p := filepath.Join(root, "doc", "adr", "0003-use-redis-for-cache.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "number: 3") || !strings.Contains(s, "status: proposed") {
		t.Fatalf("draft not normalized to next number/proposed:\n%s", s)
	}
}

func TestStatusCleanExitsZero(t *testing.T) {
	root := seedTwo(t)
	if err := runCLI(t, root, "status"); err != nil {
		t.Fatalf("clean repo should exit 0: %v", err)
	}
}

func TestStatusFailsOnBrokenLink(t *testing.T) {
	root := seedTwo(t)
	dir := filepath.Join(root, "doc", "adr")
	os.WriteFile(filepath.Join(dir, "0003-broken.md"), []byte(`---
number: 3
title: "Bad links"
status: accepted
date: "2026-05-01"
supersedes: [99]
---

# ADR-0003: Bad links

## Alternatives considered

- none
`), 0o644)

	err := runCLI(t, root, "status")
	if err == nil || !strings.Contains(err.Error(), "integrity") {
		t.Fatalf("broken repo must exit non-zero with integrity error, got %v", err)
	}
}

func TestFullFlowNewDraftSupersedeStatus(t *testing.T) {
	root := seedTwo(t)
	mustCLI(t, root, "new", "Adopt RFC process")
	mustCLI(t, root, "draft", "--pr", "7", "--provider", "mock", "--yes")
	mustCLI(t, root, "supersede", "1", "--reason", "process changed")
	if err := runCLI(t, root, "status"); err != nil {
		t.Fatalf("flow end state should be clean: %v", err)
	}
}
