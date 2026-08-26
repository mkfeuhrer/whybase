# whybase `backfill` — Design Spec

- **Date:** 2026-08-26
- **Status:** Draft for review
- **Author:** Mohit Khare (+ ox-alpha session)
- **Companion:** `docs/superpowers/specs/2026-08-26-whybase-v1-design.md` (v1), `docs/superpowers/plans/2026-08-26-whybase-v1.md`

## 1. Problem

The #1 adoption barrier for ADRs is **starting from zero**. Teams with an
established codebase (and a git history full of unrecorded decisions) see
`whybase new` as "write docs about the past" and bounce. The decisions are
already made — they live in commit messages, diffs, and PR bodies — but they
are not queryable, so agents and humans keep relitigating them.

## 2. Solution

**`whybase backfill`** — one command that mines a repo's git history for
decision-shaped changes and drafts a **`proposed`** ADR for each, into a
`proposed/` triage queue the human can accept, edit, or delete.

- **Commit scorer (offline, deterministic):** walks `git log --since`, scores
  each commit by decision-shaped signals (coverage, message verbs like
  `add|use|migrate|switch|adopt|reject|drop|replace`, churn, PR-merge
  detection, dependency-touch), returns the top-N.
- **LLM clustering pass (BYO key):** one prompt turns the top commits'
  messages+diffs into **clustered, deduplicated** candidate records — because
  a single decision often spans several commits, and `backfill` should draft
  one record per decision, not one per commit.
- **Writes `proposed/NNNN-*.md`:** separate directory, `status: proposed`,
  NOT picked up by `whybase status` / MCP until the human accepts
  (`whybase accept` moves it into `doc/adr/` and assigns the next number).
- **No key?** `--provider mock` writes a scored sample so the flow is
  dogfoodable offline.

## 3. Scope & acceptance criteria

1. `whybase backfill --since <ref> --provider mock` lists top scored commits,
   clusters them, and writes one `proposed/NNNN-*.md` per cluster.
2. `whybase status` ignores `proposed/` (staging area, not live records).
3. Accept flow: `whybase accept <file>` moves a proposed record into
   `doc/adr/` with the next number and re-rendered front-matter.
4. Commit scorer is deterministic and unit-tested (no LLM needed for the
   ranking; the LLM only clusters/drafts).
5. `go test ./...` green; dogfood on whybase itself.

## 4. Non-goals (v1)

- No auto-accept / auto-merge into live records (human gates every backfill).
- No PR-body enrichment (that's `draft --pr`'s lane; backfill works from
  commits alone so it needs no network/gh).
- No `--all` / full-history backfill (start from `--since`; add paging later).

## 5. Architecture

- `internal/backfill/` — scorer (`score.go`), clustering prompt (`prompt.go`),
  `Run()` orchestrator.
- `internal/cli/backfill.go` — cobra command: flags `--since <ref>`,
  `--provider anthropic|openai|mock`, `--limit N` (default 10), `--yes`.
- `internal/cli/accept.go` — `whybase accept <file>`.

### Commit scorer (exact signals)

```
commit score = coverage(0..1) * 40
             + verb bonus (max 30)
             + churn (0..15)
             + merge/PR bonus (10)
             + dep-touch bonus (5)
```

- **coverage** = added+deleted lines across files, clamped/log-scaled (a
  giant vendored-bump shouldn't dominate; pure-refactor commits score low).
- **verb bonus** — message first word in `add|use|migrate|switch|adopt|reject|
  drop|replace|introduce|move|remove|support|implement` → +30; `feat|fix|refactor|
  chore|docs|test` conventional prefix → +10.
- **churn** = files touched, capped at 15.
- **merge/PR bonus** = subject matches `Merge pull request|(#\d+)` → +10.
- **dep-touch bonus** = path matches `go.mod|package.json|requirements.txt|
  Cargo.toml|pom.xml` → +5.

### Clustering prompt

Inputs: top-N (commit subject, message, truncated per-commit diff). Output:
`JSON` list of clusters, each `{title, commits: [sha...], summary}`. The CLI
then drafts one ADR per cluster (context = concatenated messages, decision =
summary, alternatives = `<not recorded — backfill>` placeholder) and writes
it as `proposed/NNNN-<slug>.md`.

## 6. Risks

- **Noise → triage burden.** Mitigation: `proposed/` staging + `--limit` +
  delete-without-trace; scoring favors merges/deps, not churn.
- **LLM clustering drift.** Mitigation: prompt emits strict JSON; parse failure
  → fall back to one-record-per-top-commit (degraded but still useful).
- **`proposed/` drift.** Mitigation: `status` explicitly ignores it (tested);
  `accept` is the only promotion path.

## 7. Open questions (resolved at impl)

- Scoring weights → tuned on whybase's own history during dogfood.
- Whether `backfill` reuses `draft.Provider` → **yes**, BYO-key
  `anthropic|openai|mock`, same offline story.
