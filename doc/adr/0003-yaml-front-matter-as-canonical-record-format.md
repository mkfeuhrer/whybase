---
number: 3
title: "YAML front-matter as canonical record format"
status: accepted
date: "2026-08-26"
tags: ["format", "compatibility"]
---

# ADR-0003: YAML front-matter as canonical record format

## Context

Records must be machine-parseable (MCP serving, integrity checks, drafting)
but also readable by humans and compatible with the adr-tools ecosystem our
target users already have.

## Decision

Canonical writer output is YAML front-matter (`number/title/status/date/
supersedes/superseded_by/tags`) followed by MADR body sections. The parser
also accepts legacy front-matter-less MADR files and never rewrites them on
read; round-trip is golden-file tested.

## Alternatives considered

- Pure MADR headings, no front-matter. Rejected: regex-parsing status/links
  from prose is fragile and supersession graphs need exact structured fields.
- SQLite index database. Rejected: files are the database; a second source of
  truth drifts. YAGNI until repos exceed ~10k records.

## Consequences

- New records carry a small YAML block some purists may dislike.
- Zero-migration path for existing adr-tools/log4brains repos.
