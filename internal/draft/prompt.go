package draft

import (
	"fmt"
	"strings"
)

// PRData is everything the drafter knows about the change under discussion.
type PRData struct {
	Number int
	Title  string
	Body   string
	Diff   string
}

// Ref identifies what to draft from: a pull request number or a branch name.
type Ref struct {
	PR     int
	Branch string
}

const maxDiffChars = 60_000

// BuildPrompt renders the system prompt that instructs an LLM to emit one
// complete ADR in front-matter form. The alternatives section is mandatory —
// it is the field that stops rejected ideas coming back.
func BuildPrompt(p PRData) (string, error) {
	if p.Title == "" && p.Diff == "" {
		return "", fmt.Errorf("nothing to draft: no title or diff")
	}
	diff := p.Diff
	truncated := false
	if len(diff) > maxDiffChars {
		half := maxDiffChars / 2
		diff = diff[:half] + "\n... [diff truncated for length] ...\n" + diff[len(diff)-half:]
		truncated = true
	}
	var b strings.Builder
	b.WriteString(`You are drafting an Architecture Decision Record (ADR) for a software team.

Output EXACTLY ONE markdown document with YAML front-matter, nothing else.
Front-matter keys: number, title, status, date. Set "number: 0" (assigned later)
and "status: proposed".

The body MUST contain these sections:
## Context      - the specific pressure that forced a decision now
## Decision     - one unambiguous imperative sentence
## Alternatives considered - at least one alternative, each marked "Rejected: <reason>"
## Consequences - costs and trade-offs we accept, including bad ones

Write for engineers six months from now. Be concrete; name technologies.
`)
	fmt.Fprintf(&b, "\nPull request #%d: %s\n", p.Number, p.Title)
	if strings.TrimSpace(p.Body) != "" {
		fmt.Fprintf(&b, "\nDescription:\n%s\n", p.Body)
	}
	if truncated {
		b.WriteString("\nThe diff below was truncated to fit context limits.\n")
	}
	if strings.TrimSpace(diff) != "" {
		fmt.Fprintf(&b, "\nDiff:\n```diff\n%s\n```\n", diff)
	}
	return b.String(), nil
}
