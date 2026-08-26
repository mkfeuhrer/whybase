package adr

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Status string

const (
	Proposed   Status = "proposed"
	Accepted   Status = "accepted"
	Superseded Status = "superseded"
	Deprecated Status = "deprecated"
)

func ValidateStatus(s string) (Status, error) {
	st := Status(s)
	switch st {
	case Proposed, Accepted, Superseded, Deprecated:
		return st, nil
	}
	return "", fmt.Errorf("invalid status %q (want proposed|accepted|superseded|deprecated)", s)
}

type Record struct {
	Number       int      `yaml:"number"`
	Title        string   `yaml:"title"`
	Status       Status   `yaml:"status"`
	Date         string   `yaml:"date"`
	Supersedes   []int    `yaml:"supersedes,omitempty"`
	SupersededBy []int    `yaml:"superseded_by,omitempty"`
	Tags         []string `yaml:"tags,omitempty"`
	Body         string   `yaml:"-"`
}

var errNoFM = errors.New("no front-matter")

func splitFrontMatter(src []byte) (meta, rest []byte, err error) {
	if !bytes.HasPrefix(src, []byte("---\n")) {
		return nil, src, errNoFM
	}
	end := bytes.Index(src[4:], []byte("\n---\n"))
	if end < 0 {
		return nil, nil, errors.New("unterminated front-matter (missing closing ---)")
	}
	return src[4 : 4+end], src[4+end+5:], nil
}

// Parse reads a record file: canonical YAML front-matter form, or a legacy
// MADR file without front-matter. The body is kept verbatim.
func Parse(src []byte) (Record, error) {
	var r Record
	meta, rest, err := splitFrontMatter(src)
	if errors.Is(err, errNoFM) {
		r.Body = string(rest)
		return parseLegacy(r)
	}
	if err != nil {
		return r, err
	}
	if uerr := yaml.Unmarshal(meta, &r); uerr != nil {
		return r, fmt.Errorf("front-matter: %w", uerr)
	}
	if r.Number <= 0 {
		return r, errors.New("front-matter: number must be > 0")
	}
	if _, serr := ValidateStatus(string(r.Status)); serr != nil {
		return r, fmt.Errorf("front-matter: %w", serr)
	}
	r.Body = string(rest)
	return r, nil
}

var (
	reLegacyTitle  = regexp.MustCompile(`(?m)^#\s*(?:ADR-)?0*(\d+)\s*[.:]\s*(.+?)\s*$`)
	reLegacyStatus = regexp.MustCompile(`(?m)^Status:\s*(\S+)`)
	reLegacyDate   = regexp.MustCompile(`(?m)^Date:\s*(\d{4}-\d{2}-\d{2})`)
)

// parseLegacy fills Number/Title/Status/Date for MADR files without front-matter.
func parseLegacy(r Record) (Record, error) {
	m := reLegacyTitle.FindStringSubmatch(r.Body)
	if m == nil {
		return r, errors.New("legacy record missing numbered title heading (# 12. Title or # ADR-0012: Title)")
	}
	fmt.Sscanf(m[1], "%d", &r.Number)
	r.Title = strings.TrimSpace(m[2])
	r.Status = Proposed
	if sm := reLegacyStatus.FindStringSubmatch(r.Body); sm != nil {
		st, serr := ValidateStatus(strings.ToLower(sm[1]))
		if serr != nil {
			return r, fmt.Errorf("legacy status: %w", serr)
		}
		r.Status = st
	}
	if dm := reLegacyDate.FindStringSubmatch(r.Body); dm != nil {
		r.Date = dm[1]
	}
	return r, nil
}
