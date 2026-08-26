package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCLI executes whybase against a temp repo root.
func runCLI(t *testing.T, root string, args ...string) error {
	t.Helper()
	r := NewRoot()
	r.SetArgs(append(args, "--dir", root))
	return r.Execute()
}

func mustCLI(t *testing.T, root string, args ...string) {
	t.Helper()
	if err := runCLI(t, root, args...); err != nil {
		t.Fatalf("whybase %v: %v", args, err)
	}
}

func seedTwo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "doc", "adr")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "0001-adopt-adrs.md"), []byte(`---
number: 1
title: "Adopt ADRs"
status: accepted
date: "2026-01-05"
---

# ADR-0001: Adopt ADRs

## Context

Decisions evaporate.

## Decision

We record decisions.

## Alternatives considered

- None. Rejected: we tried chaos.
`), 0o644)
	os.WriteFile(filepath.Join(dir, "0002-use-postgres.md"), []byte(`---
number: 2
title: "Use Postgres"
status: accepted
date: "2026-02-10"
tags: ["storage"]
---

# ADR-0002: Use Postgres

## Decision

Postgres everywhere.

## Alternatives considered

- Mongo. Rejected: relations matter.
`), 0o644)
	return root
}

func TestNewCreatesNextNumber(t *testing.T) {
	root := seedTwo(t)
	mustCLI(t, root, "new", "Use Kafka For Events")

	p := filepath.Join(root, "doc", "adr", "0003-use-kafka-for-events.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, want := range []string{"number: 3", "status: proposed", "## Alternatives considered"} {
		if !strings.Contains(s, want) {
			t.Fatalf("%s missing %q:\n%s", p, want, s)
		}
	}
}

func TestNewSlugifies(t *testing.T) {
	root := t.TempDir()
	mustCLI(t, root, "new", "Hello, World! (v2) -- it's fine")
	p := filepath.Join(root, "doc", "adr", "0001-hello-world-v2-it-s-fine.md")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("expected slug file: %v", err)
	}
}

func TestSupersedeRewritesBothSides(t *testing.T) {
	root := seedTwo(t)
	mustCLI(t, root, "supersede", "2", "--new-title", "Use Postgres v15+", "--reason", "need partitioning")

	newB, err := os.ReadFile(filepath.Join(root, "doc", "adr", "0003-use-postgres-v15.md"))
	if err != nil {
		t.Fatal(err)
	}
	nb := string(newB)
	for _, want := range []string{"supersedes:\n    - 2", "status: proposed", "Supersede reason: need partitioning"} {
		if !strings.Contains(nb, want) {
			t.Fatalf("new record missing %q:\n%s", want, nb)
		}
	}

	oldB, err := os.ReadFile(filepath.Join(root, "doc", "adr", "0002-use-postgres.md"))
	if err != nil {
		t.Fatal(err)
	}
	ob := string(oldB)
	if !strings.Contains(ob, "superseded_by:\n    - 3") || !strings.Contains(ob, "status: superseded") {
		t.Fatalf("old record not updated:\n%s", ob)
	}
}

func TestSupersedeRejectsAlreadySuperseded(t *testing.T) {
	root := seedTwo(t)
	mustCLI(t, root, "supersede", "2", "--reason", "first time")
	if err := runCLI(t, root, "supersede", "2", "--reason", "again"); err == nil {
		t.Fatal("second supersede of same record must fail")
	}
}

func TestSupersedeMissingRecordFails(t *testing.T) {
	root := seedTwo(t)
	if err := runCLI(t, root, "supersede", "99", "--reason", "x"); err == nil {
		t.Fatal("supersede of missing record must fail")
	}
}
