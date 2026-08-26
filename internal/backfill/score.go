// Package backfill mines a repo's git history for decision-shaped commits and
// drafts proposed ADRs into a triage queue (proposed/), to be accepted by the
// human before they enter doc/adr/.
package backfill

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strings"
)

// Commit is a single git commit, with scoring inputs attached.
type Commit struct {
	SHA     string
	Subject string
	Body    string
	Files   []string
	// Additions/Deletions are line counts across the commit's diff.
	Additions, Deletions int
	IsMerge              bool
}

// Score weights (spec §5). Sum of maxima = 100.
const (
	wCoverage   = 40
	wVerb       = 30
	wChurn      = 15
	wMerge      = 10
	wDepTouch   = 5
	maxChurnPts = 15
)

var (
	// strongVerbs are first-word signals that a commit makes a choice.
	strongVerbs = map[string]bool{
		"add": true, "use": true, "migrate": true, "switch": true,
		"adopt": true, "reject": true, "drop": true, "replace": true,
		"introduce": true, "move": true, "remove": true, "support": true,
		"implement": true,
	}
	conventionalPrefix = regexp.MustCompile(`^(feat|fix|refactor|chore|docs|test)(\(.+\))?:`)
	depFileRe          = regexp.MustCompile(`(^|/)(go\.mod|package\.json|requirements\.txt|Cargo\.toml|pom\.xml)$`)
)

// ScoreCommit ranks how "decision-shaped" a commit is. Pure score; no IO.
func ScoreCommit(c Commit) int {
	if c.IsMerge {
		return wMerge
	}
	score := 0

	// Coverage: log-scaled line churn so one big vendored bump doesn't dominate.
	lines := c.Additions + c.Deletions
	cov := 0.0
	if lines > 0 {
		cov = logScale(float64(lines))
	}
	score += int(cov * wCoverage)

	// Verb bonus from the first word of the subject.
	if fields := strings.Fields(strings.TrimSpace(c.Subject)); len(fields) > 0 {
		first := strings.ToLower(fields[0])
		if strongVerbs[first] {
			score += wVerb
		} else if conventionalPrefix.MatchString(c.Subject) {
			score += 10
		}
	}

	// Churn: files touched, capped.
	churn := len(c.Files)
	if churn > maxChurnPts {
		churn = maxChurnPts
	}
	score += churn

	// Dependency-file touch.
	for _, f := range c.Files {
		if depFileRe.MatchString(filepath.ToSlash(f)) {
			score += wDepTouch
			break
		}
	}

	return score
}

// logScale maps 1..huge to 0..1 with a knee at ~400 lines (log2(400)≈8.64),
// so small commits score near-zero and a vendored megabump saturates at 1.
func logScale(lines float64) float64 {
	if lines <= 0 {
		return 0
	}
	v := math.Log2(lines) / 8.64
	if v > 1 {
		v = 1
	}
	return v
}

// describe renders a one-line summary of a scored commit (status output).
func describe(c Commit) string {
	files := strings.Join(c.Files, ", ")
	if files == "" {
		files = "-"
	}
	return fmt.Sprintf("%s %s (%d+/%d-) [%s]", shortSHA(c.SHA), c.Subject, c.Additions, c.Deletions, files)
}

func shortSHA(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}
