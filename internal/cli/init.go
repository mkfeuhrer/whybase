package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	fbStart = "<!-- whybase:start -->"
	fbEnd   = "<!-- whybase:end -->"
)

func managedBlock() string {
	return fbStart + `
## Architecture decisions (precedent) — managed by whybase

This repository records its architecture decisions with whybase
(https://github.com/mkfeuhrer/whybase): ADR markdown under doc/adr/,
served to coding agents over MCP.

**Mandatory workflow for any code change:**
1. BEFORE creating or editing any file, call the "check_paths" tool (whybase
   MCP server) with the paths you are about to touch.
2. If a governing decision exists: follow it. NEVER reintroduce an alternative
   listed as "Rejected" there — that rationale was already evaluated and lost.
3. If your task contradicts a record, or a significant new choice has no
   record, use "search_decisions" / "get_decision" to confirm, then surface
   the conflict to the human instead of silently violating precedent.
4. Tools: search_decisions (keywords), get_decision (by number),
   check_paths (before edits).
` + fbEnd
}

func initCmd() *cobra.Command {
	var alsoClaude bool
	c := &cobra.Command{
		Use:   "init",
		Short: "Write agent instructions into AGENTS.md so agents consult decisions unprompted",
		RunE: func(cmd *cobra.Command, args []string) error {
			targets := []string{"AGENTS.md"}
			if alsoClaude {
				targets = append(targets, "CLAUDE.md")
			}
			for _, name := range targets {
				p := filepath.Join(flagRoot, name)
				existing, err := os.ReadFile(p)
				if err != nil && !os.IsNotExist(err) {
					return err
				}
				s := string(existing)
				switch {
				case strings.Contains(s, fbStart) && strings.Contains(s, fbEnd):
					i := strings.Index(s, fbStart)
					j := strings.Index(s, fbEnd) + len(fbEnd)
					s = s[:i] + managedBlock() + s[j:]
				case s == "":
					s = "# " + filepath.Base(abs(flagRoot)) + "\n\n" + managedBlock() + "\n"
				default:
					s = strings.TrimRight(s, "\n") + "\n\n" + managedBlock() + "\n"
				}
				if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "updated %s\n", p)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nRegister the MCP server so agents can act on this:\n  claude mcp add whybase -- whybase mcp\n")
			return nil
		},
	}
	c.Flags().BoolVar(&alsoClaude, "claude", false, "also manage CLAUDE.md")
	return c
}
