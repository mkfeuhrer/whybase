package backfill

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mkfeuhrer/whybase/internal/draft"
)

// llmClusterer adapts a draft.Provider (BYO-key Anthropic/OpenAI) to the
// Clusterer interface by asking it to group commits into decisions and emit
// strict JSON.
type llmClusterer struct {
	p   draft.Provider
	max int
}

// NewLLMClusterer wraps a draft provider; max caps commits per cluster prompt.
func NewLLMClusterer(p draft.Provider, max int) Clusterer {
	if max <= 0 {
		max = 10
	}
	return &llmClusterer{p: p, max: max}
}

func (l *llmClusterer) Cluster(ctx context.Context, commits []Commit) ([]Cluster, error) {
	if len(commits) > l.max {
		commits = commits[:l.max]
	}
	prompt := BuildClusterPrompt(commits)
	md, err := l.p.Draft(ctx, draft.PRData{Title: "backfill clustering", Diff: prompt})
	if err != nil {
		return nil, err
	}
	clusters, perr := parseClusters(md)
	if perr != nil {
		return nil, fmt.Errorf("parsing cluster output: %w", perr)
	}
	return clusters, nil
}

// BuildClusterPrompt asks the LLM to group the given commits into distinct
// decisions and emit JSON only.
func BuildClusterPrompt(commits []Commit) string {
	var b strings.Builder
	b.WriteString(`You are mining git history to recover architecture decisions.

Group the commits below into clusters, where each cluster is ONE decision that
may span several commits. Use commit subjects/messages as evidence.

Return ONLY a JSON array (no prose, no markdown fences) where each element is:
{"title": "<short decision title>", "commits": ["<full sha>", ...], "summary": "<one sentence: what was decided and why>"}

Rules:
- Merge commits and dependency bumps are usually part of a larger decision, not their own cluster.
- Prefer 1-3 clusters over many; dedupe repeated subjects.
- Title should read like an ADR title: "Use Postgres for session storage".
`)

	fmt.Fprintf(&b, "\nCommits:\n")
	for i, c := range commits {
		fmt.Fprintf(&b, "%d. %s %s\n   files: %s\n", i+1, shortSHA(c.SHA), c.Subject, strings.Join(c.Files, ", "))
		if strings.TrimSpace(c.Body) != "" {
			fmt.Fprintf(&b, "   body: %s\n", firstLine(c.Body))
		}
	}
	return b.String()
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// parseClusters tolerates fences and leading prose; returns clusters.
func parseClusters(md string) ([]Cluster, error) {
	s := md
	// strip ```json ... ``` fences if the model wrapped them
	if i := strings.Index(s, "```"); i >= 0 {
		rest := s[i+3:]
		if j := strings.Index(rest, "\n"); j >= 0 {
			rest = rest[j+1:]
		}
		if j := strings.Index(rest, "```"); j >= 0 {
			rest = rest[:j]
		}
		s = rest
	}
	// trim to first [ and last ]
	i, j := strings.Index(s, "["), strings.LastIndex(s, "]")
	if i >= 0 && j > i {
		s = s[i : j+1]
	}
	var clusters []Cluster
	if err := json.Unmarshal([]byte(s), &clusters); err != nil {
		return nil, err
	}
	return clusters, nil
}
