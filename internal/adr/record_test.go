package adr

import (
	"strings"
	"testing"
)

const fmDoc = `---
number: 7
title: "Use Postgres for sessions"
status: accepted
date: 2026-03-14
supersedes: [3]
tags: ["storage"]
---

# ADR-0007: Use Postgres for sessions

## Context
Sessions outgrew the in-memory store.

## Decision
We will use Postgres for sessions.

## Alternatives considered
- Redis. Rejected: persistence story weak for our needs.
`

func TestParseFrontmatter(t *testing.T) {
	r, err := Parse([]byte(fmDoc))
	if err != nil {
		t.Fatal(err)
	}
	if r.Number != 7 || r.Title != "Use Postgres for sessions" ||
		r.Status != Accepted || r.Date != "2026-03-14" ||
		len(r.Supersedes) != 1 || r.Supersedes[0] != 3 ||
		len(r.Tags) != 1 || r.Tags[0] != "storage" {
		t.Fatalf("bad record: %+v", r)
	}
	if !strings.Contains(r.Body, "## Decision") || !strings.HasPrefix(r.Body, "\n# ADR-0007") {
		t.Fatalf("body lost or mangled: %q", r.Body[:30])
	}
}

func TestParseBadStatus(t *testing.T) {
	src := strings.Replace(fmDoc, "status: accepted", "status: approved", 1)
	_, err := Parse([]byte(src))
	if err == nil {
		t.Fatal("want error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseUnterminatedFrontmatter(t *testing.T) {
	src := "---\nnumber: 1\n"
	_, err := Parse([]byte(src))
	if err == nil || !strings.Contains(err.Error(), "unterminated") {
		t.Fatalf("want unterminated error, got %v", err)
	}
}

func TestParseZeroNumber(t *testing.T) {
	src := strings.Replace(fmDoc, "number: 7", "number: 0", 1)
	_, err := Parse([]byte(src))
	if err == nil || !strings.Contains(err.Error(), "number must be > 0") {
		t.Fatalf("want number error, got %v", err)
	}
}

func TestValidateStatus(t *testing.T) {
	for _, s := range []string{"proposed", "accepted", "superseded", "deprecated"} {
		if _, err := ValidateStatus(s); err != nil {
			t.Fatalf("%s should be valid", s)
		}
	}
	if _, err := ValidateStatus("done"); err == nil {
		t.Fatal("done should be invalid")
	}
}
