package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitWritesManagedBlock(t *testing.T) {
	root := t.TempDir()
	mustCLI(t, root, "init")
	b, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "<!-- whybase:start -->") ||
		!strings.Contains(s, "<!-- whybase:end -->") ||
		!strings.Contains(s, "check_paths") {
		t.Fatalf("managed block missing:\n%s", s)
	}
}

func TestInitPreservesExistingContent(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# My project\n\nCustom notes.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustCLI(t, root, "init")
	b, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "# My project") || !strings.Contains(s, "Custom notes.") {
		t.Fatalf("existing content lost:\n%s", s)
	}
}

func TestInitIdempotent(t *testing.T) {
	root := t.TempDir()
	mustCLI(t, root, "init")
	mustCLI(t, root, "init")
	b, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if got := strings.Count(string(b), "<!-- whybase:start -->"); got != 1 {
		t.Fatalf("want exactly 1 block after double init, got %d", got)
	}
}
