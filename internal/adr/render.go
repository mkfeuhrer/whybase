package adr

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Render serializes a Record to canonical front-matter form. If the body does
// not start with a numbered heading, a MADR skeleton (Context, Decision,
// Alternatives considered, Consequences) is prepended so new records are
// complete from birth.
func Render(r Record) ([]byte, error) {
	if r.Number <= 0 {
		return nil, fmt.Errorf("render: number must be > 0")
	}
	if _, err := ValidateStatus(string(r.Status)); err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	var fm bytes.Buffer
	fm.WriteString("---\n")
	meta, err := yaml.Marshal(struct {
		Number       int      `yaml:"number"`
		Title        string   `yaml:"title"`
		Status       Status   `yaml:"status"`
		Date         string   `yaml:"date"`
		Supersedes   []int    `yaml:"supersedes,omitempty"`
		SupersededBy []int    `yaml:"superseded_by,omitempty"`
		Tags         []string `yaml:"tags,omitempty"`
	}{r.Number, r.Title, r.Status, r.Date, r.Supersedes, r.SupersededBy, r.Tags})
	if err != nil {
		return nil, err
	}
	fm.Write(meta)
	fm.WriteString("---\n")

	body := r.Body
	if !hasNumberedHeading(body) {
		body = skeleton(r) + strings.TrimLeft(body, "\n")
	}
	if !hasSection(body, "## Alternatives considered") {
		body += "\n## Alternatives considered\n\n- Option. Rejected: why.\n"
	}
	body = strings.TrimLeft(body, "\n")
	out := append(fm.Bytes(), '\n') // exactly one blank line after ---
	out = append(out, []byte(strings.TrimRight(body, "\n")+"\n")...)
	return out, nil
}

func hasSection(body, heading string) bool {
	for _, line := range strings.Split(body, "\n") {
		if strings.TrimSpace(line) == heading {
			return true
		}
	}
	return false
}

func hasNumberedHeading(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "# ") {
			return true
		}
	}
	return false
}

func skeleton(r Record) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# ADR-%04d: %s\n\n", r.Number, r.Title)
	b.WriteString("## Context\n\nWhat forced this decision?\n\n")
	b.WriteString("## Decision\n\nOne unambiguous sentence.\n\n")
	b.WriteString("## Alternatives considered\n\n- Option. Rejected: why.\n\n")
	b.WriteString("## Consequences\n\n- Trade-offs we accept.\n")
	return b.String()
}
