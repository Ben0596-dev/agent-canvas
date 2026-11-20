# Backend Development Guide

## Technology Stack

- Go 1.25
- Gin for HTTP transport
- GORM with PostgreSQL for persistence
- Explicit SQL migrations under `migrations/`; application startup never runs `AutoMigrate`
- OpenAPI 3.1 contract at `../contracts/openapi.yaml`

## Four-Layer Architecture

Keep production code under `internal/` and preserve this dependency direction:

```text
controller → service → repository → model
```

- `controller/`: register Gin routes, bind and validate transport DTOs, call services, and map errors to HTTP responses. Keep business rules out of controllers. Controllers import services and models, not repository implementations or GORM query APIs.
- `service/`: implement use cases, business validation, transaction orchestration, revision checks, and Agent workflows. Services depend on repository interfaces and models; they do not import Gin or GORM.
- `repository/`: declare persistence interfaces, shared persistence errors, and concrete query/write implementations. Put adapters in subpackages such as `repository/gorm` and `repository/memory`; repositories operate on structures from `model/`.
- `model/`: define database entity structures, GORM column tags, relationships, and `TableName` mappings. Models contain structure declarations rather than query execution or HTTP behavior. HTTP request and response DTOs belong to the controller layer.

`cmd/server/main.go` is the composition root. It may construct concrete adapters and inject them into services and controllers; business behavior does not belong there. Shared infrastructure such as database initialization stays under `internal/platform/`.

## Contract-First TDD

Update `../contracts/openapi.yaml` before changing an external HTTP or SSE interface. Implement one vertical slice at a time using red → green:

1. Add or change the OpenAPI operation and schemas.
2. Add a failing test at the public seam.
3. Implement the smallest controller, service, repository, and model changes that make it pass.
4. Run formatting, tests, vet, and contract validation.

Keep Go tests beside the package they exercise as `*_test.go`. Controller tests use `httptest` through the router; service tests use repository interfaces; GORM adapter tests require a real PostgreSQL test database when database behavior matters. Assert observable behavior rather than private functions or GORM call sequences.

## Domain and Persistence Rules

- Every canvas mutation is a validated `Canvas Operation` batch.
- Project semantic writes use `expectedRevision` and commit atomically; a mismatch maps to HTTP 409.
- Agent output is a proposal until the user confirms it. Agent code cannot write project tables directly.
- Use explicit transactions for multi-record writes and keep transaction ownership in the service/repository seam.
- Store large assets outside project JSON and reference them by stable ID.
- Keep the single-user identity behind an interface or explicit value so later authentication does not require rewriting project queries.

## Validation

Run from `backend/`:

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
```

Also run `git diff --check` from the repository root. A backend change is complete when the OpenAPI contract, migrations, implementation, and tests agree for every affected external behavior.
