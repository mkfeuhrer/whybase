---
number: 2
title: "Shell out to gh CLI instead of GitHub SDK"
status: accepted
date: "2026-08-26"
tags: ["dependencies", "cli"]
---

# ADR-0002: Shell out to gh CLI instead of GitHub SDK

## Context

`whybase draft` needs PR titles, bodies and diffs. Options were the official
GitHub Go SDK (go-github) or shelling out to the `gh` CLI most developers who
use PRs already have installed and authenticated.

## Decision

We shell out to `gh pr view` / `gh pr diff` via `exec.Command`, behind a
`Fetcher` interface so tests inject a fake binary.

## Alternatives considered

- go-github SDK. Rejected: adds a dependency + token management for data `gh`
  already authenticates; binary size and supply-chain surface grow for zero
  user-visible benefit.
- Raw REST calls with GITHUB_TOKEN. Rejected: pushes auth setup onto every
  user on day one; gh already solved it.

## Consequences

- `draft` requires `gh` installed (clear error if missing); all other commands stay dependency-free.
- Testing uses fake script binaries — no network in CI.
