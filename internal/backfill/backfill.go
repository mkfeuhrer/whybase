package backfill

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Cluster is one proposed ADR: a decision spanning one or more commits.
type Cluster struct {
	Title   string   `json:"title"`
	Commits []string `json:"commits"` // SHAs
	Summary string   `json:"summary"` // one-paragraph decision summary
}

// Options controls a backfill run.
type Options struct {
	Since    string // git ref for `git log --since` (e.g. "6 months ago")
	Limit    int    // top-N scored commits to consider (default 10)
	Provider string // anthropic | openai | mock
}

// Result summarizes what a run produced.
type Result struct {
	Proposed []string // written proposed/ file paths
	Skipped  int
}

// Run walks git history, scores, clusters (via provider when available), and
// writes proposed/ records. No provider (mock) falls back to one record per
// top commit so the flow is dogfoodable offline.
func Run(ctx context.Context, root string, o Options, cl Clusterer) (*Result, error) {
	if o.Since == "" {
		o.Since = "1 month ago"
	}
	if o.Limit == 0 {
		o.Limit = 10
	}
	commits, err := gitLog(ctx, root, o.Since)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(commits, func(i, j int) bool {
		return ScoreCommit(commits[i]) > ScoreCommit(commits[j])
	})
	if len(commits) > o.Limit {
		commits = commits[:o.Limit]
	}

	clusters, cerr := cl.Cluster(ctx, commits)
	if cerr != nil || len(clusters) == 0 {
		// Degraded but useful: one record per top commit.
		clusters = onePerCommit(commits)
	}

	proposedDir := filepath.Join(root, "proposed")
	if err := os.MkdirAll(proposedDir, 0o755); err != nil {
		return nil, err
	}

	res := &Result{}
	used := map[string]bool{}
	for i, cl := range clusters {
		title := strings.TrimSpace(cl.Title)
		if title == "" {
			title = "Backfilled decision (commits " + strings.Join(cl.Commits, ",") + ")"
		}
		slug := slugify(title)
		if used[slug] {
			slug = fmt.Sprintf("%s-%d", slug, i+1)
		}
		used[slug] = true
		p := filepath.Join(proposedDir, fmt.Sprintf("%03d-%s.md", i+1, slug))
		md := renderProposed(title, cl.Summary, cl.Commits, o.Since)
		if err := os.WriteFile(p, []byte(md), 0o644); err != nil {
			return res, err
		}
		res.Proposed = append(res.Proposed, p)
	}
	return res, nil
}

// Clusterer turns scored commits into candidate records. The LLM providers
// implement it; mock returns one per commit.
type Clusterer interface {
	Cluster(ctx context.Context, commits []Commit) ([]Cluster, error)
}

// onePerCommit is the no-LLM fallback.
func onePerCommit(commits []Commit) []Cluster {
	out := make([]Cluster, 0, len(commits))
	for _, c := range commits {
		out = append(out, Cluster{
			Title:   strings.TrimSpace(c.Subject),
			Commits: []string{c.SHA},
			Summary: "Backfilled from commit " + shortSHA(c.SHA) + ": " + strings.TrimSpace(c.Subject),
		})
	}
	return out
}

func renderProposed(title, summary string, commits []string, since string) string {
	shas := make([]string, 0, len(commits))
	for _, s := range commits {
		shas = append(shas, shortSHA(s))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "---\nnumber: 0\ntitle: %q\nstatus: proposed\ndate: %q\ntags: [\"backfill\"]\n---\n\n", title, time.Now().Format("2006-01-02"))
	fmt.Fprintf(&b, "# ADR-0000: %s\n\n## Context\n\nBackfilled from git history since %q (commits %s).\n\n", title, since, strings.Join(shas, ", "))
	if summary != "" {
		fmt.Fprintf(&b, "## Decision\n\n%s\n\n", summary)
	}
	b.WriteString("## Alternatives considered\n\n- Not recorded (backfilled). Rejected: the original decision rationale predates whybase; this record captures the outcome, not the full debate.\n\n")
	b.WriteString("## Consequences\n\n- Backfilled record: verify against the actual commit history before accepting.\n")
	return b.String()
}

// gitLog shells out to git (no dependency) for commit metadata since a ref.
func gitLog(ctx context.Context, root, since string) ([]Commit, error) {
	args := []string{"log", "--since=" + since, "--pretty=format:%H%x00%s%x00%b%x00%an", "--numstat"}
	cmd := exec.CommandContext(ctx, "git", append(args, "--", ".")...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	return parseGitLog(string(out))
}

func parseGitLog(raw string) ([]Commit, error) {
	var commits []Commit
	var cur *Commit
	for _, line := range strings.Split(raw, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\x00")
		switch len(parts) {
		case 4: // commit header: sha, subject, body, author
			commits = append(commits, Commit{SHA: parts[0], Subject: parts[1], Body: parts[2]})
			cur = &commits[len(commits)-1]
		case 2: // numstat row: added, deleted, file
			if cur == nil {
				continue
			}
			var add, del int
			fmt.Sscanf(parts[0], "%d", &add)
			fmt.Sscanf(parts[1], "%d", &del)
			cur.Additions += add
			cur.Deletions += del
			if !strings.Contains(parts[1], "-") {
				cur.Files = append(cur.Files, parts[1])
			}
		}
	}
	return commits, nil
}

// slugify mirrors the CLI slug rule (lowercase, non-alnum → -, max 50).
func slugify(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(s) {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 50 {
		out = strings.TrimRight(out[:50], "-")
	}
	if out == "" {
		out = "untitled"
	}
	return out
}

var _ = json.Marshal // keep encoding/json imported for cluster prompt future use
