# Whybase v1 — Design Spec

- **Date:** 2026-08-26
- **Status:** Draft for review
- **Author:** Mohit Khare (+ ox-alpha session)
- **Repo (planned):** `github.com/mkfeuhrer/whybase`
- **Companion research artifacts:** `~/AI/artifacts/2026-08-26-taski-reborn-decision-tooling.html`, `~/AI/artifacts/2026-08-26-taski-option-a-deepdive.html`

## 1. Problem

Three failures compound in 2026:

1. **Decisions evaporate.** Engineering choices are made in Slack threads, PR reviews, and AI-agent sessions — then scroll away. Six months later teams relitigate settled questions with worse context.
2. **AI agents repeat mistakes.** Coding agents read code but not rationale. Without queryable records of *what was rejected and why*, agents re-propose banned libraries and reopen dead debates.
3. **The tooling gap.** Incumbents are unmaintained (`adr-tools` ★5.6k — last push Apr 2024; `log4brains` — Dec 2024). New entrants are tiny (<★150) and single-purpose (enforcement-only or capture-only). The one integrated system (Decispher) is cloud SaaS that OSS projects and security-conscious orgs cannot adopt.

**Who has this pain:** platform/staff engineers and eng managers at teams using Claude Code/Cursor/Copilot daily; OSS maintainers; solo devs losing their own context across agent sessions.

## 2. Solution

**Whybase** — one Go binary that makes architecture decisions a first-class, queryable part of every repo.

- **Markdown-native:** MADR-format records in `doc/adr/` (adr-tools-compatible layout). No lock-in; exit cost is deleting `.taski/`… i.e. nothing.
- **Author fast:** `whybase new|status|supersede` manage numbering, statuses, supersession links, graph integrity.
- **Draft with AI:** `whybase draft --pr N | --branch B` pulls diff + commit messages via `gh`, drafts a complete record (context, decision, *alternatives considered* — the highest-value field), writes it as `proposed`. BYO API key (`ANTHROPIC_API_KEY` / `OPENAI_API_KEY`); no key → everything except `draft` still works.
- **Serve to agents:** `whybase mcp` exposes stdio MCP tools (`search_decisions`, `get_decision`, `check_paths`) so coding agents pull relevant decisions before writing code.
- **Free OSS forever.** MIT. Hosted/cross-repo features are a possible later layer, never a requirement.

**Positioning line:** *"Every why in your codebase, one query away."*

## 3. Why this wins (market evidence, Aug 2026)

- Write-layer incumbents abandoned → "successor to adr-tools" is a launchable narrative with inherited mindshare.
- Demand validated by velocity: `keep-the-why` went 0→★144 in ~7 weeks (created Jul 2026); multiple MCP-ADR micro-projects appearing monthly.
- Empty cell: **full loop (capture→integrity→serve-agents), OSS-first/local**. Nobody holds it.
- Adjacent category heat: Cursor ~$50B valuation talks, Claude Code ~$2.5B ARR, Cognee ★30k — agent memory/context is a funded battleground; curated decision records are the defensible slice platforms won't build.

## 4. v1 Scope & acceptance criteria

1. Commands work end-to-end on a real repo: `new`, `draft`, `status`, `list`, `supersede`, `mcp`.
2. Round-trips standard MADR markdown without mutation; precise file:line errors on malformed front-matter.
3. A coding agent queries decisions via MCP and receives governing records in <1s on a 500-record repo.
4. Integrity report catches: dangling supersedes, stale statuses, missing alternatives section.
5. Installable via `go install ./cmd/whybase@latest`; Homebrew formula follows launch.
6. `go test ./...` green; golden-file tests for markdown round-trip; e2e: new→draft(mock LLM)→supersede→mcp query against fixture repo.

## 5. Architecture

```
cmd/whybase          cobra entrypoint only
internal/adr         MADR markdown parse/write + front-matter (pure)
internal/store       discovery, in-memory index, integrity graph checks
internal/draft       gh CLI diff fetch → LLM Provider iface (Anthropic/OpenAI adapters)
internal/mcpserver   stdio MCP server over store (official Go SDK)
```

- Storage: files + optional generated index cache; no database (YAGNI).
- Config: `.whybase.yml`, all defaults overridable (dir, template, model).

## 6. Non-goals (v1)

Enforcement gates / GitHub App · Slack/Jira capture · hosted service · static-site generation (stay log4brains-compatible instead) · Windows support claims.

## 7. Validation plan (staged, each stage independently useful)

| # | Test | Kill criterion |
|---|---|---|
| 1 | Dogfood 2 weeks on own repos | Not still using after week 1 |
| 2 | Launch post: "adr-tools is unmaintained — building the successor" | <200 pts HN AND <50 stars week 1 |
| 3 | 10 staff/platform engineers, 90-sec demo video | <4 unprompted "my team would install" |
| 4 | Ship MCP server standalone first, watch installs | No organic issues/PRs from strangers |

Stages gate further investment — designed explicitly against the Taski-1.0 failure mode (months of heads-down before any external signal).

## 8. Risks

- **Field velocity:** expect 3–5 new entrants before year-end; window is months, not years.
- **Platform absorption:** if Cursor/Claude Code ship native repo memory, standalone value compresses. Counter: cross-tool, cross-repo, human-governed, exportable records stay outside their incentives.
- **OSS grind:** revenue distant by design; success metrics for v1 are adoption signals (stars, contributors, dogfood retention), not money.

## 9. Naming decision record

`precedent` (metaphor) rejected as not self-explanatory; availability-checked shortlist → **whybase**: 0 Go repos, no brew collision, states both artifact (*why*) and value (*queryable base*).
