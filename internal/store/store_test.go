package store

import (
	"os"
	"path/filepath"
	"testing"
)

func fixtureRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const rec1 = `---
number: 1
title: "Adopt ADRs"
status: accepted
date: "2026-01-05"
---

# ADR-0001: Adopt ADRs

## Decision

We record decisions.
`

const rec7 = `---
number: 7
title: "Use Kafka"
status: accepted
date: "2026-03-01"
---

# ADR-0007: Use Kafka
`

const rec3 = `---
number: 3
title: "Use Postgres"
status: proposed
date: "2026-02-10"
tags: ["storage"]
---

# ADR-0003: Use Postgres

## Context

Need durable storage.
`

func TestLoadOrdersAndMaps(t *testing.T) {
	dir := fixtureRepo(t, map[string]string{
		"0001-adopt-adrs.md":   rec1,
		"0003-use-postgres.md": rec3,
	})
	ix, ferrs, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ferrs) != 0 {
		t.Fatalf("unexpected file errors: %v", ferrs)
	}
	if len(ix.Ordered) != 2 || ix.Ordered[0].Number != 1 || ix.Ordered[1].Number != 3 {
		t.Fatalf("bad order: %+v", ix.Ordered)
	}
	if r, ok := ix.Get(3); !ok || r.Title != "Use Postgres" {
		t.Fatalf("Get(3) failed: %+v ok=%v", r, ok)
	}
}

func TestNextNumberWithGaps(t *testing.T) {
	dir := fixtureRepo(t, map[string]string{"0001-a.md": rec1, "0007-b.md": rec7})
	ix, _, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if ix.NextNumber() != 8 {
		t.Fatalf("want 8, got %d", ix.NextNumber())
	}
}

func TestEmptyDir(t *testing.T) {
	ix, ferrs, err := Load(t.TempDir())
	if err != nil || len(ferrs) != 0 {
		t.Fatalf("clean empty dir should load: %v %v", err, ferrs)
	}
	if ix.NextNumber() != 1 || len(ix.Ordered) != 0 {
		t.Fatalf("empty index wrong: %+v", ix)
	}
}

func TestMissingDirIsError(t *testing.T) {
	_, _, err := Load(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("want error for missing dir")
	}
}

func TestDuplicateNumbersReported(t *testing.T) {
	dir := fixtureRepo(t, map[string]string{
		"0001-a.md":      rec1,
		"0001-b-copy.md": rec1,
		"0002-b.md":      rec3,
	})
	ix, ferrs, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ferrs) != 1 || ferrs[0].Number() != 1 {
		t.Fatalf("want one duplicate-number FileError for 1, got %+v", ferrs)
	}
	if r, exists := ix.Get(1); !exists || filepath.Base(r.Path) != "0001-a.md" {
		t.Fatalf("first occurrence should win: %+v exists=%v", r, exists)
	}
	if len(ix.Ordered) != 2 {
		t.Fatalf("want records 1 and 2 indexed, got %d", len(ix.Ordered))
	}
}
