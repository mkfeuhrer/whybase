package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	var tags []string
	c := &cobra.Command{
		Use:   "new <title>",
		Short: "Create a proposed decision record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ix, dir, _, err := loadIndex(flagRoot)
			if err != nil {
				return err
			}
			r := freshRecord(ix.NextNumber(), args[0])
			r.Tags = tags
			p, err := writeRecord(dir, r)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", p)
			return nil
		},
	}
	c.Flags().StringSliceVar(&tags, "tag", nil, "tags for the record")
	return c
}

func listCmd() *cobra.Command {
	var status string
	c := &cobra.Command{
		Use:   "list",
		Short: "List decision records",
		RunE: func(cmd *cobra.Command, args []string) error {
			ix, _, ferrs, err := loadIndex(flagRoot)
			if err != nil {
				return err
			}
			for _, fe := range ferrs {
				fmt.Fprintln(stderr, "warning:", fe)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "NUMBER  STATUS       DATE        TITLE")
			for _, r := range ix.Ordered {
				if status != "" && string(r.Status) != status {
					continue
				}
				fmt.Fprintf(out, "%06d  %-11s  %-10s  %s\n", r.Number, r.Status, r.Date, r.Title)
			}
			return nil
		},
	}
	c.Flags().StringVar(&status, "status", "", "filter by status")
	return c
}

func supersedeCmd() *cobra.Command {
	var reason, newTitle string
	c := &cobra.Command{
		Use:   "supersede <number>",
		Short: "Replace a decision with a new record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if reason == "" {
				return fmt.Errorf("--reason is required (future readers need to know why)")
			}
			var old int
			if _, err := fmt.Sscanf(args[0], "%d", &old); err != nil {
				return fmt.Errorf("bad number %q", args[0])
			}
			ix, dir, _, err := loadIndex(flagRoot)
			if err != nil {
				return err
			}
			prev, ok := ix.Get(old)
			if !ok {
				return fmt.Errorf("no record %d", old)
			}
			if prev.Status == "superseded" {
				return fmt.Errorf("ADR-%04d is already superseded by %v", old, prev.SupersededBy)
			}
			title := newTitle
			if title == "" {
				title = "Supersede " + prev.Title
			}
			next := ix.NextNumber()
			repl := freshRecord(next, title)
			repl.Supersedes = []int{old}
			repl.Body = "\n# " + headingTitle(repl) + "\n\n## Context\n\n> Supersede reason: " + reason + "\n\nThis record replaces ADR-" + zero4(old) + ".\n"
			pNew, err := writeRecord(dir, repl)
			if err != nil {
				return err
			}
			prev.Status = "superseded"
			prev.SupersededBy = []int{next}
			pOld, err := writeRecord(dir, prev)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s\nupdated %s\n", pNew, pOld)
			return nil
		},
	}
	c.Flags().StringVar(&reason, "reason", "", "why this decision is being reversed")
	c.Flags().StringVar(&newTitle, "new-title", "", "title for the replacement record")
	return c
}
