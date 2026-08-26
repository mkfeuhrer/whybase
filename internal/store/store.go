package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mkfeuhrer/whybase/internal/adr"
)

// FileError is a per-file problem that must not abort loading the rest of the
// repository (one broken record should not hide the other ninety).
type FileError struct {
	Path string
	Err  error
}

func (fe FileError) Error() string { return fmt.Sprintf("%s: %v", fe.Path, fe.Err) }

// Number returns the record number parsed from the filename prefix, or 0.
func (fe FileError) Number() int {
	var n int
	fmt.Sscanf(filepath.Base(fe.Path), "%d", &n)
	return n
}

type Index struct {
	ByNumber map[int]adr.Record
	Ordered  []adr.Record // ascending by number
}

// Load walks dir for *.md files and parses each into an index.
func Load(dir string) (*Index, []FileError, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	ix := &Index{ByNumber: map[int]adr.Record{}}
	paths := map[int]string{}
	var ferrs []FileError
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		p := filepath.Join(dir, e.Name())
		src, rerr := os.ReadFile(p)
		if rerr != nil {
			ferrs = append(ferrs, FileError{p, rerr})
			continue
		}
		rec, perr := adr.Parse(src)
		if perr != nil {
			ferrs = append(ferrs, FileError{p, perr})
			continue
		}
		if first, dup := paths[rec.Number]; dup {
			ferrs = append(ferrs, FileError{p, fmt.Errorf("duplicate number %d (already loaded from %s)", rec.Number, filepath.Base(first))})
			continue
		}
		rec.Path = p
		paths[rec.Number] = p
		ix.ByNumber[rec.Number] = rec
		ix.Ordered = append(ix.Ordered, rec)
	}
	sort.Slice(ix.Ordered, func(i, j int) bool { return ix.Ordered[i].Number < ix.Ordered[j].Number })
	return ix, ferrs, nil
}

func (ix *Index) NextNumber() int {
	max := 0
	for n := range ix.ByNumber {
		if n > max {
			max = n
		}
	}
	return max + 1
}

func (ix *Index) Get(n int) (adr.Record, bool) {
	r, ok := ix.ByNumber[n]
	return r, ok
}
