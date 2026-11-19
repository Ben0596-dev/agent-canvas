# ADR-0004: Keep the first Agent runtime in Go

- Status: accepted
- Date: 2026-11-17

## Decision

The first Agent orchestrator runs in Go. Python is not a runtime dependency of the product path; it may later be added as an isolated worker for CV, OCR, embeddings, or GPU inference. Python workers return structured results and cannot mutate the canvas directly.
