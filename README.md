# OpenReview AI

OpenReview AI is an open-source, self-hostable GitHub App for automated pull request reviews using configurable LLM providers and repository-specific review standards.

The goal is not to replace human reviewers. OpenReview AI is designed to act as a senior engineer doing the first review pass: catching security risks, correctness issues, architectural drift, and production-readiness problems before a maintainer spends time on the PR.

## Why OpenReview AI

Most AI review tools are generic, closed, or tied to one vendor. OpenReview AI is built around a different set of assumptions:

- teams should own their review standards
- security findings should be first-class, not an afterthought
- review prompts should be inspectable and versioned
- providers should be swappable
- self-hosting should be normal
- local models should be possible for privacy-sensitive teams

## Current Status

This repository is early-stage. The current Go service contains the foundation for:

- GitHub webhook handling
- reviewer personas
- prompt planning and rendering
- safe prompt include expansion
- defensive prompt wrapping
- normalized finding models
- in-memory review/profile storage
- mock provider execution

The next major implementation milestones are real GitHub App authentication, PR diff fetching, structured LLM output parsing, and OpenRouter/OpenAI-compatible provider support.

## Architecture

```text
GitHub Pull Request
  -> Webhook API
  -> Review Engine
  -> Reviewer Personas
  -> Prompt Planner
  -> Prompt Renderer
  -> LLM Provider
  -> Findings Normalizer
  -> GitHub Comments and Summary
```

See [docs/architecture.md](docs/architecture.md) and [docs/prompts.md](docs/prompts.md) for more detail.

## Quick Start

Requirements:

- Go 1.25+
- Docker, optional

Run locally:

```powershell
go run ./cmd/openreview
```

The service listens on `:8080` by default.

```powershell
$env:OPENREVIEW_ADDR=":8081"
$env:GITHUB_WEBHOOK_SECRET="dev-secret"
go run ./cmd/openreview
```

Run with Docker:

```powershell
docker compose up --build
```

## Endpoints

- `GET /healthz`
- `POST /webhooks/github`
- `GET /reviews/{id}`
- `GET /repositories`
- `POST /review-profiles`
- `PUT /review-profiles/{id}`

## Review Profiles

Review profiles let teams define standards and reviewer personas.

```json
{
  "id": "security-first",
  "name": "Security First",
  "rules": [
    "prioritize security",
    "flag SQL injection risks",
    "detect authorization issues"
  ],
  "reviewers": [
    "security-engineer",
    "staff-backend-engineer"
  ]
}
```

Profiles will later be loaded from repository-level `openreview.yml` files.

## Prompt Library

OpenReview AI treats `prompts/` as a prompt library. The application selects a curated set of prompts based on review profile and reviewer persona, then wraps them with defensive PR-review policy before sending them to a model.

Current prompt orchestration supports:

- prompt registry
- `@include(...)` expansion
- path traversal protection
- variable interpolation
- JSON output contract injection

## Roadmap

See [ROADMAP.md](ROADMAP.md).

For the step-by-step build plan and implementation checklist, see [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md).

## Contributing

Contributions are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md), then check open issues or propose a focused change.

Useful commands:

```powershell
gofmt -w cmd internal
go test ./...
go build ./cmd/openreview
```

## Security

Please do not report vulnerabilities through public issues. See [SECURITY.md](SECURITY.md).

OpenReview AI contains security review prompts, but the application itself must remain defensive by default. Prompts and provider integrations should not enable autonomous exploitation, credential extraction, or destructive testing.

## License

OpenReview AI is licensed under the [MIT License](LICENSE).
