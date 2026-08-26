package cli

import (
	"fmt"
	"path/filepath"

	"github.com/mkfeuhrer/whybase/internal/adr"
)

func freshRecord(n int, title string) adr.Record {
	return adr.Record{
		Number: n,
		Title:  title,
		Status: adr.Proposed,
	}
}

func zero4(n int) string { return fmt.Sprintf("%04d", n) }

// headingTitle renders "ADR-0007: Title" used at the top of record bodies.
func headingTitle(r adr.Record) string {
	return fmt.Sprintf("ADR-%s: %s", zero4(r.Number), r.Title)
}

func abs(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return a
}
