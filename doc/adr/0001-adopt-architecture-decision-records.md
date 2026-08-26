---
number: 1
title: "Adopt architecture decision records"
status: accepted
date: "2026-08-26"
tags: ["process"]
---

# ADR-0001: Adopt architecture decision records

## Context

Whybase is built for the agent era: coding agents and teammates both need
queryable rationale. Dogfooding our own tool is the cheapest way to validate
the authoring UX and catch integrity issues early.

## Decision

We will record every architecturally significant whybase decision as MADR
markdown in `doc/adr/`, managed by the whybase CLI itself.

## Alternatives considered

- GitHub Discussions. Rejected: not versioned with code, not parseable by our
  own MCP server, invisible to offline tooling.
- No internal process; rely on commit messages. Rejected: commits record what
  changed, never what was rejected or why.

## Consequences

- Slight ceremony on notable decisions; mitigated by `whybase draft`.
- Our own repo doubles as the reference example for users.
