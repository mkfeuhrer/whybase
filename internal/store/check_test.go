package store

import (
	"testing"

	"github.com/mkfeuhrer/whybase/internal/adr"
)

func rec(n int, title, status string, supersedes, supersededBy []int, body string) adr.Record {
	return adr.Record{Number: n, Title: title, Status: adr.Status(status),
		Date: "2026-01-01", Supersedes: supersedes, SupersededBy: supersededBy,
		Tags: []string{"storage"}, Body: body, Path: "/x"}
}

const fullBody = "\n# T\n\n## Context\nx\n\n## Decision\ny\n\n## Alternatives considered\n- Redis. Rejected: cost.\n"
const noAltBody = "\n# T\n\n## Context\nx\n\n## Decision\ny\n"

func index(records ...adr.Record) *Index {
	ix := &Index{ByNumber: map[int]adr.Record{}}
	for _, r := range records {
		ix.ByNumber[r.Number] = r
		ix.Ordered = append(ix.Ordered, r)
	}
	return ix
}

func TestCheckClean(t *testing.T) {
	ix := index(
		rec(1, "A", "accepted", nil, nil, fullBody),
		rec(2, "B", "superseded", nil, []int{3}, fullBody),
		rec(3, "C", "accepted", []int{2}, nil, fullBody),
	)
	if issues := ix.Check(); len(issues) != 0 {
		t.Fatalf("clean repo should have no issues: %+v", issues)
	}
}

func TestCheckDanglingLinks(t *testing.T) {
	ix := index(rec(1, "A", "accepted", []int{99}, nil, fullBody))
	issues := ix.Check()
	if len(issues) != 1 || issues[0].Severity != "error" || issues[0].Number != 1 {
		t.Fatalf("want error issue on 1, got %+v", issues)
	}
}

func TestCheckStaleStatus(t *testing.T) {
	ix := index(
		rec(1, "A", "accepted", nil, []int{2}, fullBody), // marked replaced but still accepted
		rec(2, "B", "accepted", []int{1}, nil, fullBody),
	)
	found := false
	for _, is := range ix.Check() {
		if is.Number == 1 && is.Severity == "warn" && containsStr(is.Message, "stale status") {
			found = true
		}
	}
	if !found {
		t.Fatal("missing stale-status warning")
	}
}

func TestCheckMissingBacklink(t *testing.T) {
	ix := index(
		rec(1, "A", "accepted", nil, nil, fullBody),
		rec(2, "B", "accepted", []int{1}, nil, fullBody), // supersedes 1 but 1 has no backlink
	)
	found := false
	for _, is := range ix.Check() {
		if is.Number == 1 && containsStr(is.Message, "missing backlink") {
			found = true
		}
	}
	if !found {
		t.Fatal("missing backlink warning")
	}
}

func TestCheckMissingAlternatives(t *testing.T) {
	ix := index(rec(1, "A", "proposed", nil, nil, noAltBody))
	found := false
	for _, is := range ix.Check() {
		if is.Number == 1 && containsStr(is.Message, "missing alternatives") {
			found = true
		}
	}
	if !found {
		t.Fatal("missing-alternatives warning")
	}
}

func TestSearch(t *testing.T) {
	ix := index(
		rec(1, "Use Postgres for sessions", "accepted", nil, nil, fullBody+"sessions live in pg."),
		rec(2, "Use Kafka for events", "accepted", nil, nil, fullBody),
		rec(3, "Cache with Redis", "superseded", nil, []int{4}, fullBody),
	)
	got := ix.Search("postgres sessions")
	if len(got) != 1 || got[0].Number != 1 {
		t.Fatalf("search postgres sessions → %+v", got)
	}
	got = ix.Search("redis")
	if len(got) == 0 || got[0].Number != 3 {
		t.Fatalf("search redis should rank 3 first → %+v", got)
	}
}

func TestGoverning(t *testing.T) {
	ix := index(
		rec(1, "Postgres for session storage", "accepted", nil, nil, fullBody+"applies to internal/sessions."),
		rec(2, "Kafka for events", "accepted", nil, nil, fullBody),
	)
	got := ix.Governing([]string{"internal/sessions/store.go"})
	if len(got) != 1 || got[0].Number != 1 {
		t.Fatalf("governing(internal/sessions/store.go) → %+v", got)
	}
}

// helpers
func containsStr(h, n string) bool {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return true
		}
	}
	return false
}
