# ADR-0001: Establish the engineering baseline

- Status: accepted
- Date: 2026-11-17

## Context

The repository contains a Next.js skeleton and an empty Go backend. The first milestone is engineering readiness for an Agent plus infinite canvas product, not feature breadth.

## Decision

Use one repository with independently deployable `web` and `backend` modules. Keep all Node dependencies, tooling configuration, and the only Node lockfile inside `web`; the repository root is not a Node package. Standardize local and CI checks around Docker Compose, strict TypeScript, ESLint, Prettier, tests, Go tests, and Conventional Commits. Enforce the same checks through Husky hooks and CI.

## Consequences

Contributors get one repeatable workflow and a single contract/documentation location. Local hooks add a small commit-time cost, while CI remains the authoritative gate.
