package draft

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Fetcher retrieves change context from the repo hosting service.
type Fetcher interface {
	Fetch(ctx context.Context, ref Ref) (PRData, error)
}

// Provider turns PRData into a drafted record.
type Provider interface {
	Draft(ctx context.Context, p PRData) (draftedMarkdown string, err error)
}

type ghFetcher struct{ bin string }

// NewGHFetcher shells out to the gh CLI (no SDK dependency).
func NewGHFetcher(bin string) Fetcher {
	if bin == "" {
		bin = "gh"
	}
	return ghFetcher{bin}
}

func (g ghFetcher) run(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, g.bin, args...).Output()
	if err != nil {
		return "", fmt.Errorf("gh %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

func (g ghFetcher) Fetch(ctx context.Context, ref Ref) (PRData, error) {
	if ref.PR == 0 && ref.Branch == "" {
		return PRData{}, fmt.Errorf("empty ref: need --pr or --branch")
	}
	if ref.PR != 0 {
		viewJSON, err := g.run(ctx, "pr", "view", fmt.Sprint(ref.PR), "--json", "number,title,body")
		if err != nil {
			return PRData{}, err
		}
		var v struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Body   string `json:"body"`
		}
		if jerr := json.Unmarshal([]byte(viewJSON), &v); jerr != nil {
			return PRData{}, fmt.Errorf("parsing gh output: %w", jerr)
		}
		diff, derr := g.run(ctx, "pr", "diff", fmt.Sprint(ref.PR))
		if derr != nil {
			return PRData{}, derr
		}
		return PRData{Number: v.Number, Title: v.Title, Body: v.Body, Diff: diff}, nil
	}
	log, err := g.run(ctx, "log", "--merges", "--oneline", "-5", ref.Branch)
	if err != nil {
		return PRData{}, err
	}
	diff, err := g.run(ctx, "diff", "--stat", "main..."+ref.Branch)
	if err != nil {
		return PRData{}, err
	}
	return PRData{Title: "Branch " + ref.Branch, Body: log, Diff: diff}, nil
}

type mockProvider struct{ md string }

// NewMock returns a provider that always emits fixed markdown (tests + offline dogfooding).
func NewMock(markdown string) Provider { return mockProvider{md: markdown} }

func (m mockProvider) Draft(_ context.Context, _ PRData) (string, error) { return m.md, nil }
