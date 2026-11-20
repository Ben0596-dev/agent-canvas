# ADR-0004: Keep the first Agent runtime in Go

- Status: accepted
- Date: 2026-11-17

## Decision

The first Agent orchestrator and product request path run in Go. Python is not a runtime dependency of the product path; it may later be added as an isolated worker for CV, OCR, embeddings, or GPU inference. Python workers return structured results through an explicit adapter and cannot mutate the canvas or bypass `Canvas Operation` directly.

## Considered Options

- **Python as the primary Agent service**: has a broad AI ecosystem, but introduces a second runtime, deployment path and cross-process failure mode before the product needs it.
- **Go as the orchestrator with optional workers**: keeps HTTP, persistence, cancellation and revision transactions in one runtime while leaving a seam for specialized workloads.

## Consequences

- Go owns Agent Run lifecycle, cancellation, retries, structured output validation and Operation proposal persistence.
- A future Python worker is an adapter behind a narrow interface; it returns data and never receives direct database credentials.
- Adding Python later is a deliberate architecture change, not an implicit dependency of the first release.
