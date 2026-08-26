package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackfillWritesProposedAndAcceptPromotes(t *testing.T) {
	root := t.TempDir()
	mustGitCLI(t, root, "init", "-b", "master")
	mustGitCLI(t, root, "config", "user.email", "t@t")
	mustGitCLI(t, root, "config", "user.name", "t")
	os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644)
	mustGitCLI(t, root, "add", ".")
	mustGitCLI(t, root, "commit", "-m", "add session store")
	os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n\nvar cache = map[string]string{}\n"), 0o644)
	mustGitCLI(t, root, "commit", "-am", "use in-memory cache for session reads")

	mustCLI(t, root, "backfill", "--since", "1 year ago", "--limit", "5", "--provider", "mock")

	proposedDir := filepath.Join(root, "proposed")
	entries, err := os.ReadDir(proposedDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no proposed records written")
	}
	// Proposed files are NOT live records: status must ignore them.
	mustCLI(t, root, "status")

	// Accept the first proposed record → becomes ADR-0001 in doc/adr/.
	first := filepath.Join(proposedDir, entries[0].Name())
	mustCLI(t, root, "accept", first)
	live, err := os.ReadFile(filepath.Join(root, "doc", "adr", "0001-backfilled-decision.md"))
	if err != nil {
		t.Fatalf("accepted record not in doc/adr: %v", err)
	}
	if !strings.Contains(string(live), "status: accepted") || !strings.Contains(string(live), "number: 1") {
		t.Fatalf("accepted record wrong:\n%s", live)
	}
	if _, err := os.Stat(first); !os.IsNotExist(err) {
		t.Fatalf("proposed file should be removed after accept, still exists: %v", err)
	}
	mustCLI(t, root, "status")
}

func TestAcceptRejectsNonProposed(t *testing.T) {
	root := seedTwo(t)
	live := filepath.Join(root, "doc", "adr", "0002-use-postgres.md")
	if err := runCLI(t, root, "accept", live); err == nil {
		t.Fatal("accept of an accepted record must fail")
	}
}

func mustGitCLI(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2026-01-01T00:00:00Z")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
