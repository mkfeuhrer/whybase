# whybase agent-compliance A/B — run protocol

> Money-shot validation: does an AI agent that can query recorded decisions
> stop re-proposing rejected approaches?
> Source: `~/AI/artifacts/2026-08-26-whybase-validation-showcase.html` (Stage 3).

## Setup (one-time)

```sh
cd ~/Desktop/projects/whybase-demo
git pull                                   # ensure AGENTS.md init block present
claude mcp add whybase -- whybase mcp      # register MCP server
# verify: whybase status → all good: 5 records, 0 issues
```

## Run A (control) ×5 — no MCP

Fresh Claude Code session in `whybase-demo`, **no** MCP registered.
Prompt (vary phrasing slightly each run):

> "Add session caching to this service."

Score each run:
- Did the agent cite any ADR / existing precedent?
- Did it propose Redis (the recorded rejection)?

## Run B (MCP) ×5

Same prompt, but MCP registered. Expected:
- Agent calls `check_paths` on `sessions/store.go`
- Cites ADR-0002's rejection rationale (Redis) and/or ADR-0005 (read replica)
- Implements within precedent

## Scoring table

| Run | Control (no MCP): Redis? cited ADR? | With MCP: check_paths? cited ADR? avoided Redis? |
|---|---|---|
| 1 | | |
| 2 | | |
| 3 | | |
| 4 | | |
| 5 | | |

**Verdict rules:**
- Control proposes Redis ≥4/5 and MCP avoids it ≥4/5 → the contrast is the launch hero.
- Any MCP run that *violates* a recorded decision → bug; fix before launch.

## Publishing guidance (playbook)

Even partial results are publishable: "4/5 runs the agent cited the decision
and avoided Redis; control proposed Redis 5/5" is a front-page HN sentence.
Capture transcripts side-by-side (asciinema or screenshots) — that contrast
is the hero image.
