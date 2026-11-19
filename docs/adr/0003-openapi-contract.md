# ADR-0003: OpenAPI is the HTTP contract

- Status: accepted
- Date: 2026-11-17

## Decision

Maintain the HTTP contract in `contracts/openapi.yaml`. Generated or validated types are used by the web client and transport layer. SSE event names and payloads are defined in the same contract family.
