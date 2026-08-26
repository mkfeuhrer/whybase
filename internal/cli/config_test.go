package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	root := t.TempDir()
	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ADRDir != "doc/adr" {
		t.Fatalf("want doc/adr, got %q", cfg.ADRDir)
	}
}

func TestLoadConfigOverride(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".whybase.yml"), []byte("adr_dir: decisions\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ADRDir != "decisions" {
		t.Fatalf("want decisions, got %q", cfg.ADRDir)
	}
}
