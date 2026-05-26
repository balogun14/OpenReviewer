# Architecture

OpenReview AI starts as a single Go service with explicit internal package boundaries.

```text
GitHub Webhook
  -> HTTP API
  -> Review Engine
  -> Reviewer Personas
  -> Prompt Planner
  -> Prompt Renderer
  -> Provider Client
  -> Findings
```

## Current Packages

- `cmd/openreview`: process entrypoint.
- `internal/httpapi`: HTTP routes, request decoding, in-memory stores.
- `internal/github`: GitHub webhook payload and signature handling.
- `internal/review`: review orchestration, persona fan-out, severity summary, recommendation.
- `internal/prompt`: prompt manifest, safe loader, include processing, interpolation, prompt planning.
- `internal/provider`: provider request contracts and mock provider.
- `internal/profile`: review profile model and defaults.
- `internal/finding`: normalized finding model and deduplication.
- `internal/config`: environment configuration.

## Near-Term Direction

The single process should stay until the review flow is real end to end. After GitHub App auth, diff fetching, and first provider integrations are working, split background review execution behind a queue.

Recommended order:

1. GitHub App installation token support.
2. Pull request file and diff fetching.
3. Structured LLM output parser and validator.
4. OpenRouter provider.
5. Inline comments and summary comments.
6. PostgreSQL persistence.
7. Redis or PostgreSQL-backed review queue.
