package cli

import (
	"fmt"
	"os"

	"github.com/mkfeuhrer/whybase/internal/adr"
	"github.com/mkfeuhrer/whybase/internal/backfill"
	"github.com/mkfeuhrer/whybase/internal/draft"
	"github.com/spf13/cobra"
)

func backfillCmd() *cobra.Command {
	var since, provider string
	var limit int
	c := &cobra.Command{
		Use:   "backfill",
		Short: "Mine git history for decisions and draft proposed ADRs",
		RunE: func(cmd *cobra.Command, args []string) error {
			prov := clusterProvider(provider)
			if prov == nil {
				return fmt.Errorf("unknown provider %q (anthropic|openai|mock)", provider)
			}
			root := abs(flagRoot)
			res, err := backfill.Run(cmd.Context(), root, backfill.Options{
				Since: since, Limit: limit, Provider: provider,
			}, prov)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(res.Proposed) == 0 {
				fmt.Fprintln(out, "no decision-shaped commits found since", since)
				return nil
			}
			for _, p := range res.Proposed {
				fmt.Fprintf(out, "proposed %s\n", p)
			}
			fmt.Fprintf(out, "\n%d proposed record(s) written to proposed/.\n", len(res.Proposed))
			fmt.Fprintf(out, "Review them, then promote with: whybase accept <file>\n")
			return nil
		},
	}
	c.Flags().StringVar(&since, "since", "1 month ago", "git ref for commit range (e.g. \"6 months ago\", \"v1.0\")")
	c.Flags().StringVar(&provider, "provider", "", "anthropic | openai | mock (default: env key availability)")
	c.Flags().IntVar(&limit, "limit", 10, "max scored commits to consider")
	return c
}

func acceptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "accept <file>",
		Short: "Promote a proposed record into doc/adr/",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ix, dir, _, err := loadIndex(flagRoot)
			if err != nil {
				return err
			}
			src := args[0]
			b, rerr := os.ReadFile(src)
			if rerr != nil {
				return rerr
			}
			rec, perr := adr.Parse(b)
			if perr != nil {
				return fmt.Errorf("proposed record unparsable: %w", perr)
			}
			if rec.Status != adr.Proposed {
				return fmt.Errorf("%s is %q, not proposed", src, rec.Status)
			}
			rec.Number = ix.NextNumber()
			rec.Status = adr.Accepted
			p, werr := writeRecord(dir, rec)
			if werr != nil {
				return werr
			}
			if rerr := os.Remove(src); rerr != nil && !os.IsNotExist(rerr) {
				return rerr
			}
			fmt.Fprintf(cmd.OutOrStdout(), "accepted %s → %s\n", src, p)
			return nil
		},
	}
}

// clusterProvider builds the backfill Clusterer from the CLI provider flag,
// reusing the same BYO-key draft.Provider plumbing.
func clusterProvider(name string) backfill.Clusterer {
	switch name {
	case "mock":
		return backfill.NewLLMClusterer(draft.NewMock(backfillMockClusters), 10)
	case "anthropic", "":
		key := os.Getenv("ANTHROPIC_API_KEY")
		if key != "" {
			return backfill.NewLLMClusterer(draft.NewAnthropic(key, ""), 10)
		}
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key != "" {
			return backfill.NewLLMClusterer(draft.NewOpenAI(key, ""), 10)
		}
	}
	return nil
}

const backfillMockClusters = "```json\n" +
	`[{"title":"Backfilled decision","commits":[],"summary":"Sample cluster from mock provider."}]` +
	"\n```"
