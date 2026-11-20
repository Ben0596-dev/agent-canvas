# ADR-0002: Use Gin and GORM for the Go application

- Status: accepted
- Date: 2026-11-17

## Context

The backend needs a conventional HTTP layer and repository seams while the domain is still evolving.

## Decision

Use Gin for HTTP transport and GORM for persistence. Organize backend business code as `controller → service → repository → model`: controllers adapt Gin and HTTP DTOs, services own use cases and transactions, repositories implement database execution, and models define GORM entity structures and table mappings. Database schema changes remain explicit migration files; GORM AutoMigrate is forbidden in application startup.

## Consequences

The team gets familiar routing and productive persistence with a predictable four-layer layout. Explicit migrations, repository tests, and transaction tests are mandatory to avoid implicit schema drift. Controller DTOs remain separate from database models so persistence-only fields are not exposed over HTTP.
