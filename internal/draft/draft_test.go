package draft

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fake gh binary: echoes canned JSON for `pr view`, diff for `pr diff`.
func fakeGH(t *testing.T, viewJSON, diffText string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "gh")
	script := "#!/bin/sh\ncase \"$*\" in\n  *view*) cat <<'EOF'\n" + viewJSON + "\nEOF\n;;\n  *diff*) cat <<'EOF'\n" + diffText + "\nEOF\n;;\nesac\n"
	if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestGHFetcherFetchesPR(t *testing.T) {
	gh := fakeGH(t,
		`{"title":"Use Redis","body":"Because latency","number":9}`,
		"diff --git a/cache.go b/cache.go\n+redis",
	)
	f := NewGHFetcher(gh)
	d, err := f.Fetch(context.Background(), Ref{PR: 9})
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "Use Redis" || d.Number != 9 || !strings.Contains(d.Diff, "cache.go") || !strings.Contains(d.Body, "latency") {
		t.Fatalf("bad PRData: %+v", d)
	}
}

func TestGHFetcherRejectsEmptyRef(t *testing.T) {
	f := NewGHFetcher("gh")
	if _, err := f.Fetch(context.Background(), Ref{}); err == nil {
		t.Fatal("want error for empty ref")
	}
}

func TestMockProvider(t *testing.T) {
	md := "---\nnumber: 0\ntitle: Mocked\nstatus: proposed\n---\n\n# Mocked\n\n## Alternatives considered\n- none yet\n"
	p := NewMock(md)
	out, err := p.Draft(context.Background(), PRData{Number: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "title: Mocked") || !strings.Contains(out, "status: proposed") {
		t.Fatalf("mock output wrong:\n%s", out)
	}
}
