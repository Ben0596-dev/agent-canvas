# ADR-0002: Use Gin and GORM for the Go application

- Status: accepted
- Date: 2026-11-17

## Context

The backend needs a conventional HTTP layer and repository seams while the domain is still evolving.

## Decision

Use Gin for HTTP transport and GORM for persistence. Database schema changes remain explicit migration files; GORM AutoMigrate is forbidden in application startup. Repository interfaces hide GORM from application modules.

## Consequences

The team gets familiar routing and productive persistence. Explicit migrations, repository tests, and transaction tests are mandatory to avoid implicit schema drift and ORM leakage.
