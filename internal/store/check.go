package store

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mkfeuhrer/whybase/internal/adr"
)

// Issue is an integrity problem found by Check.
type Issue struct {
	Severity string // "error" | "warn"
	Number   int    // record it concerns (0 = repo-level)
	Message  string
}

func (i Issue) String() string {
	tag := "WARN "
	if i.Severity == "error" {
		tag = "ERROR"
	}
	if i.Number == 0 {
		return fmt.Sprintf("%s %s", tag, i.Message)
	}
	return fmt.Sprintf("%s ADR-%04d: %s", tag, i.Number, i.Message)
}

// Check validates cross-record integrity: link targets exist, supersession is
// bidirectional and status-consistent, and the alternatives section — the
// highest-value field — is present.
func (ix *Index) Check() []Issue {
	var issues []Issue
	for _, r := range ix.Ordered {
		for _, n := range append(append([]int{}, r.Supersedes...), r.SupersededBy...) {
			if _, ok := ix.ByNumber[n]; !ok {
				issues = append(issues, Issue{"error", r.Number,
					fmt.Sprintf("links to missing record %d", n)})
			}
		}
		if len(r.SupersededBy) > 0 && r.Status != adr.Superseded {
			issues = append(issues, Issue{"warn", r.Number,
				fmt.Sprintf("stale status: superseded by %s but status is %q", ints(r.SupersededBy), r.Status)})
		}
		for _, older := range r.Supersedes {
			if o, ok := ix.ByNumber[older]; ok && !containsInt(o.SupersededBy, r.Number) {
				issues = append(issues, Issue{"warn", older,
					fmt.Sprintf("missing backlink: superseded by ADR-%04d which does not list it in superseded_by", r.Number)})
			}
		}
		if !hasAlternatives(r.Body) {
			issues = append(issues, Issue{"warn", r.Number, "missing alternatives section"})
		}
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].Number < issues[j].Number })
	return issues
}

func hasAlternatives(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		t := strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(t, "## alternatives") {
			rest := body[indexOfFold(body, line)+len(line):]
			return strings.TrimSpace(rest) != ""
		}
	}
	return false
}

func indexOfFold(h, needle string) int {
	hl, nl := strings.ToLower(h), strings.ToLower(needle)
	return strings.Index(hl, nl)
}

func containsInt(xs []int, x int) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func ints(xs []int) string {
	parts := make([]string, len(xs))
	for i, v := range xs {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(parts, ",")
}

var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "for": true, "of": true, "and": true,
	"to": true, "in": true, "on": true, "with": true, "use": true, "using": true,
}

func tokens(s string) []string {
	f := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})
	out := f[:0]
	for _, w := range f {
		if len(w) > 2 && !stopWords[w] {
			out = append(out, w)
		}
	}
	return out
}

// Search returns records matching all query tokens across title, tags and
// body, best matches first, capped at 10.
func (ix *Index) Search(q string) []adr.Record {
	want := tokens(q)
	if len(want) == 0 {
		return nil
	}
	type scored struct {
		r adr.Record
		s int
	}
	var hits []scored
	for _, r := range ix.Ordered {
		hay := strings.ToLower(r.Title + " " + strings.Join(r.Tags, " ") + " " + r.Body)
		score := 0
		for _, w := range want {
			if strings.Contains(hay, w) {
				score++
				if strings.Contains(strings.ToLower(r.Title), w) {
					score++ // title hits weigh double
				}
			}
		}
		if score >= len(want) { // require every token somewhere
			hits = append(hits, scored{r, score})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].s > hits[j].s })
	if len(hits) > 10 {
		hits = hits[:10]
	}
	out := make([]adr.Record, len(hits))
	for i, h := range hits {
		out[i] = h.r
	}
	return out
}

// Governing returns the records most relevant to the given file paths so an
// agent can consult them before editing. Path segments are matched against
// record text; top 5 returned.
func (ix *Index) Governing(paths []string) []adr.Record {
	var want []string
	for _, p := range paths {
		want = append(want, tokens(p)...)
	}
	if len(want) == 0 {
		return nil
	}
	type scored struct {
		r adr.Record
		s int
	}
	var hits []scored
	for _, r := range ix.Ordered {
		title := strings.ToLower(r.Title)
		tags := strings.Join(r.Tags, " ")
		body := strings.ToLower(r.Body)
		score := 0
		for _, w := range dedupe(want) {
			switch {
			case strings.Contains(title, w):
				score += 3
			case strings.Contains(tags, w):
				score += 2
			case strings.Contains(body, w):
				score++
			}
		}
		if score > 0 {
			hits = append(hits, scored{r, score})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].s != hits[j].s {
			return hits[i].s > hits[j].s
		}
		return hits[i].r.Number > hits[j].r.Number
	})
	if len(hits) > 5 {
		hits = hits[:5]
	}
	out := make([]adr.Record, len(hits))
	for i, h := range hits {
		out[i] = h.r
	}
	return out
}

func dedupe(xs []string) []string {
	seen := map[string]bool{}
	out := xs[:0]
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}
