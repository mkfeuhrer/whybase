package backfill

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mkfeuhrer/whybase/internal/draft"
)

// staticClusterer is a test double returning fixed clusters.
type staticClusterer struct{ clusters []Cluster }

func (s staticClusterer) Cluster(_ context.Context, _ []Commit) ([]Cluster, error) {
	return s.clusters, nil
}

func TestParseClustersFencesAndProse(t *testing.T) {
	in := "Here are the decisions:\n```json\n[{\"title\":\"Use Postgres\",\"commits\":[\"a1\"],\"summary\":\"switch\"}]\n```"
	clusters, err := parseClusters(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(clusters) != 1 || clusters[0].Title != "Use Postgres" || clusters[0].Commits[0] != "a1" {
		t.Fatalf("bad parse: %+v", clusters)
	}
}

func TestRunWritesProposed(t *testing.T) {
	root := t.TempDir()
	mustGit(t, root, "init", "-b", "master")
	mustGit(t, root, "config", "user.email", "t@t")
	mustGit(t, root, "config", "user.name", "t")
	os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644)
	mustGit(t, root, "add", ".")
	mustGit(t, root, "commit", "-m", "add session store")
	os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n\nvar cache = map[string]string{}\n"), 0o644)
	mustGit(t, root, "commit", "-am", "use in-memory cache for session reads")

	cl := staticClusterer{clusters: []Cluster{{
		Title: "Use in-memory cache for sessions", Commits: []string{"any"}, Summary: "We cache sessions in-process.",
	}}}
	res, err := Run(context.Background(), root, Options{Since: "1 year ago", Limit: 5, Provider: "mock"}, cl)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Proposed) != 1 {
		t.Fatalf("want 1 proposed, got %+v", res)
	}
	b, err := os.ReadFile(res.Proposed[0])
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, want := range []string{"number: 0", "status: proposed", "## Alternatives considered", "Use in-memory cache for sessions"} {
		if !strings.Contains(s, want) {
			t.Fatalf("proposed missing %q:\n%s", want, s)
		}
	}
}

func TestRunFallsBackToOnePerCommit(t *testing.T) {
	root := t.TempDir()
	mustGit(t, root, "init", "-b", "master")
	mustGit(t, root, "config", "user.email", "t@t")
	mustGit(t, root, "config", "user.name", "t")
	os.WriteFile(filepath.Join(root, "a.go"), []byte("package main\n"), 0o644)
	mustGit(t, root, "add", ".")
	mustGit(t, root, "commit", "-m", "add auth middleware")
	os.WriteFile(filepath.Join(root, "b.go"), []byte("package main\n"), 0o644)
	mustGit(t, root, "add", ".")
	mustGit(t, root, "commit", "-m", "add rate limiter")

	res, err := Run(context.Background(), root, Options{Since: "1 year ago", Limit: 5, Provider: "mock"}, failingClusterer{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Proposed) != 2 {
		t.Fatalf("fallback should draft one per top commit, got %d: %+v", len(res.Proposed), res.Proposed)
	}
}

type failingClusterer struct{}

func (failingClusterer) Cluster(context.Context, []Commit) ([]Cluster, error) {
	return nil, os.ErrNotExist
}

func TestLLMClustererParsesOutput(t *testing.T) {
	p := draft.NewMock("```json\n[{\"title\":\"Use Redis cache\",\"commits\":[\"x1\"],\"summary\":\"cache\"}]\n```")
	cl := NewLLMClusterer(p, 5)
	got, err := cl.Cluster(context.Background(), []Commit{{SHA: "x1", Subject: "add cache"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "Use Redis cache" {
		t.Fatalf("bad cluster: %+v", got)
	}
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2026-01-01T00:00:00Z")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
