# whybase

> Every why in your codebase, one query away.

**Whybase** makes architecture decisions a first-class, queryable part of your
repository. Records are plain MADR markdown in `doc/adr/` (adr-tools
compatible), can be AI-drafted from pull requests, and are served to coding
agents — Claude Code, Cursor, Copilot — over MCP so they stop re-proposing
rejected approaches.

## Why

AI agents read your code but not your rationale. Without queryable records of
*what was rejected and why*, agents reintroduce banned libraries and reopen
settled debates. Humans have the same problem: decisions made in Slack and PR
threads evaporate. Whybase is the system of record for both audiences.

## Install

```sh
go install github.com/mkfeuhrer/whybase/cmd/whybase@latest
```

Requires Go 1.24+. A Homebrew formula will follow the first tagged release.

## Quickstart

```sh
whybase init                               # teaches your agents to consult decisions (AGENTS.md)
whybase new "Use Postgres for sessions"     # creates doc/adr/0001-use-postgres-for-sessions.md (proposed)
cd doc/adr && $EDITOR 0001-*.md             # fill in Decision + Alternatives considered
whybase list                                # NUMBER STATUS DATE TITLE
whybase supersede 1 --reason "scale"        # reverse a decision, with a trail
whybase status                              # integrity report; exit 1 on errors
```

## Draft records from PRs with AI

```sh
export ANTHROPIC_API_KEY=sk-...   # or OPENAI_API_KEY
whybase draft --pr 42 --yes       # fetches diff via gh, drafts full MADR record
```

No API key? `--provider mock` writes a sample record offline. Everything except
`draft` works fully offline, forever.

## Serve to AI agents (MCP)

```sh
claude mcp add whybase -- whybase mcp   # Claude Code
# opencode: add the same command under [mcp] in opencode.json
```

Then run `whybase init` in the repo — it writes a managed block into
`AGENTS.md` instructing agents to call `check_paths` before every edit and to
never reintroduce rejected alternatives. Without this step, agents only use
the tools when asked; with it, consultation is the default.

Tools exposed:

| Tool | What it gives the agent |
|---|---|
| `search_decisions` | keyword search over all records |
| `get_decision` | full record by number |
| `check_paths` | governing decisions (+ rejected alternatives) for files about to be edited |

## How it compares

| | whybase | adr-tools | log4brains |
|---|---|---|---|
| maintained | yes | stalled (2024) | stalled (2024) |
| AI drafting from PRs | yes | — | — |
| MCP server for agents | yes | — | — |
| integrity checking | yes | — | — |
| static site publishing | planned | — | yes |

Whybase deliberately does **not** do: enforcement gates / GitHub Apps, Slack or
Jira capture, or hosted anything (v1). Markdown in git is the database; leave
anytime by deleting `doc/adr/`.

## Development

```sh
make test    # go test ./...
make build   # ./whybase
make install # go install ./cmd/whybase
```

Design spec: [`docs/superpowers/specs/2026-08-26-whybase-v1-design.md`](docs/superpowers/specs/2026-08-26-whybase-v1-design.md) ·
Implementation plan: [`docs/superpowers/plans/2026-08-26-whybase-v1.md`](docs/superpowers/plans/2026-08-26-whybase-v1.md)

MIT © Mohit Khare
