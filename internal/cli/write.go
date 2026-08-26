package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mkfeuhrer/whybase/internal/adr"
	"github.com/mkfeuhrer/whybase/internal/store"
)

func loadIndex(root string) (*store.Index, string, []store.FileError, error) {
	cfg, err := LoadConfig(root)
	if err != nil {
		return nil, "", nil, err
	}
	dir := filepath.Join(root, cfg.ADRDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, dir, nil, err
	}
	ix, ferrs, err := store.Load(dir)
	if err != nil {
		return nil, dir, nil, err
	}
	return ix, dir, ferrs, nil
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(title string) string {
	s := strings.ToLower(title)
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = strings.TrimRight(s[:50], "-")
	}
	if s == "" {
		s = "untitled"
	}
	return s
}

func fileName(r adr.Record) string {
	return fmt.Sprintf("%04d-%s.md", r.Number, slugify(r.Title))
}

// writeRecord renders and writes a record into dir using its canonical name.
func writeRecord(dir string, r adr.Record) (string, error) {
	if r.Date == "" {
		r.Date = time.Now().Format("2006-01-02")
	}
	out, err := adr.Render(r)
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, fileName(r))
	if err := os.WriteFile(p, out, 0o644); err != nil {
		return "", err
	}
	return p, nil
}

var errSilent = errors.New("")
