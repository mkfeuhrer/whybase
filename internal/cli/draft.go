package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mkfeuhrer/whybase/internal/adr"
	"github.com/mkfeuhrer/whybase/internal/draft"
	"github.com/spf13/cobra"
)

// ErrIntegrity signals `status` found error-severity issues: exit non-zero
// without printing a redundant "error:" line.
var ErrIntegrity = errors.New("integrity check failed")

func draftCmd() *cobra.Command {
	var pr int
	var branch, provider string
	var assumeYes bool
	c := &cobra.Command{
		Use:   "draft",
		Short: "Draft an ADR from a PR or branch using AI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pr == 0 && branch == "" {
				return fmt.Errorf("need --pr <number> or --branch <name>")
			}
			var prov draft.Provider
			switch provider {
			case "mock":
				prov = draft.NewMock(mockDraftMD)
			case "anthropic", "":
				key := os.Getenv("ANTHROPIC_API_KEY")
				if key == "" {
					return fmt.Errorf("ANTHROPIC_API_KEY not set (or pass --provider mock)")
				}
				prov = draft.NewAnthropic(key, "")
			case "openai":
				key := os.Getenv("OPENAI_API_KEY")
				if key == "" {
					return fmt.Errorf("OPENAI_API_KEY not set (or pass --provider mock)")
				}
				prov = draft.NewOpenAI(key, "")
			default:
				return fmt.Errorf("unknown provider %q (anthropic|openai|mock)", provider)
			}

			fetcher := draft.NewGHFetcher("")
			ref := draft.Ref{PR: pr, Branch: branch}
			var data draft.PRData
			if provider == "mock" {
				data = draft.PRData{Number: pr, Title: "Mock pull request", Body: "offline sample", Diff: "diff --git a/main.go b/main.go"}
			} else {
				var ferr2 error
				data, ferr2 = fetcher.Fetch(cmd.Context(), ref)
				if ferr2 != nil {
					return ferr2
				}
			}
			md, err := prov.Draft(cmd.Context(), data)
			if err != nil {
				return err
			}
			rec, err := adr.Parse([]byte(md))
			if err != nil {
				return fmt.Errorf("model output unparsable as ADR: %w\n--- model said ---\n%s", err, truncate(md, 800))
			}
			rec.Status = adr.Proposed

			ix, dir, _, err := loadIndex(flagRoot)
			if err != nil {
				return err
			}
			rec.Number = ix.NextNumber()
			if !assumeYes {
				out, rerr := adr.Render(rec)
				if rerr != nil {
					return rerr
				}
				fmt.Fprintf(stderr, "%s\nWrite this record? [y/N] ", out)
				var answer string
				fmt.Fscanln(os.Stdin, &answer)
				answer = strings.ToLower(strings.TrimSpace(answer))
				if answer != "y" && answer != "yes" {
					fmt.Fprintln(stderr, "aborted; nothing written")
					return nil
				}
			}
			p, err := writeRecord(dir, rec)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", p)
			return nil
		},
	}
	c.Flags().IntVar(&pr, "pr", 0, "pull request number to draft from")
	c.Flags().StringVar(&branch, "branch", "", "branch to draft from (vs main)")
	c.Flags().StringVar(&provider, "provider", "", "anthropic | openai | mock (default from env key availability)")
	c.Flags().BoolVar(&assumeYes, "yes", false, "write without confirmation prompt")
	return c
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Report record integrity issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			ix, _, ferrs, err := loadIndex(flagRoot)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, fe := range ferrs {
				fmt.Fprintln(out, "ERROR", fe)
			}
			issues := ix.Check()
			for _, is := range issues {
				fmt.Fprintln(out, is.String())
			}
			if len(ferrs) == 0 && len(issues) == 0 {
				fmt.Fprintf(out, "all good: %d records, 0 issues\n", len(ix.Ordered))
				return nil
			}
			return ErrIntegrity
		},
	}
}

const mockDraftMD = `---
number: 0
title: "Use Redis for cache"
status: proposed
---

# ADR-0000: Use Redis for cache

## Context
Latency spikes under load; p99 breached SLO twice last quarter.

## Decision
Cache hot keys in Redis with a 5-minute TTL.

## Alternatives considered
- Memcached. Rejected: no persistence and we already run Redis elsewhere.
- Local in-memory LRU. Rejected: invalidation across replicas is unsolved.

## Consequences
- New operational surface (Redis HA).
- Stale reads up to TTL length are acceptable per product.

`

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…[truncated]"
}
