package adr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderFrontmatter(t *testing.T) {
	r := Record{
		Number: 7, Title: "Use Postgres for sessions", Status: Accepted,
		Date: "2026-03-14", Supersedes: []int{3}, Tags: []string{"storage"},
		Body: "\n# ADR-0007: Use Postgres for sessions\n\n## Context\nSessions outgrew memory.\n",
	}
	out, err := Render(r)
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	for _, want := range []string{
		"---\n", "number: 7", `title: Use Postgres for sessions`, "status: accepted",
		"date: \"2026-03-14\"", "supersedes:\n    - 3", "tags:\n    - storage",
		"## Alternatives considered",
	} {
		if !contains(got, want) {
			t.Fatalf("render missing %q:\n%s", want, got)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	in := []byte(fmDoc)
	r, err := Parse(in)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := Parse(MustRender(t, r))
	if err != nil {
		t.Fatal(err)
	}
	if r.Number != r2.Number || r.Title != r2.Title || r.Status != r2.Status ||
		r.Date != r2.Date || !intsEq(r.Supersedes, r2.Supersedes) ||
		!strsEq(r.Tags, r2.Tags) || r.Body != r2.Body {
		t.Fatalf("round trip mismatch:\n%+v\n%+v", r, r2)
	}
}

func TestRenderAddsSkeleton(t *testing.T) {
	r := Record{Number: 1, Title: "Adopt ADRs", Status: Proposed, Date: "2026-08-26"}
	out := string(MustRender(t, r))
	if !contains(out, "# ADR-0001: Adopt ADRs") ||
		!contains(out, "## Context") || !contains(out, "## Decision") ||
		!contains(out, "## Alternatives considered") || !contains(out, "## Consequences") {
		t.Fatalf("skeleton sections missing:\n%s", out)
	}
}

func TestGoldenFiles(t *testing.T) {
	fixtures := []struct {
		name string
		rec  Record
	}{
		{"simple", Record{Number: 1, Title: "Adopt ADRs", Status: Proposed, Date: "2026-08-26"}},
		{"superseded", Record{Number: 3, Title: "Use Redis", Status: Superseded,
			Date: "2025-01-09", SupersededBy: []int{7}}},
		{"full", Record{Number: 7, Title: "Use Postgres for sessions", Status: Accepted,
			Date: "2026-03-14", Supersedes: []int{3}, Tags: []string{"storage", "sessions"}}},
	}
	update := os.Getenv("GOLDEN_UPDATE") == "1"
	for _, f := range fixtures {
		golden := filepath.Join("testdata", "golden", f.name+".md")
		out, err := Render(f.rec)
		if err != nil {
			t.Fatal(err)
		}
		if update {
			if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
				t.Fatal(err)
			}
			os.WriteFile(golden, out, 0o644)
			continue
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("golden missing (%s): run GOLDEN_UPDATE=1 go test ./internal/adr/", golden)
		}
		if string(out) != string(want) {
			t.Errorf("golden mismatch for %s:\n--- got ---\n%s\n--- want ---\n%s", f.name, out, want)
		}
	}
}

// helpers

func MustRender(t *testing.T, r Record) []byte {
	t.Helper()
	b, err := Render(r)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func contains(h, n string) bool { return indexOfStr(h, n) >= 0 }

func indexOfStr(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}

func intsEq(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func strsEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
