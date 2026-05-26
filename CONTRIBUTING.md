# Contributing

Thanks for helping build OpenReview AI.

This project is early, so the highest-value contributions are focused, well-tested changes that improve the review pipeline without locking the project into one provider, hosting model, or workflow.

## Development Setup

Requirements:

- Go 1.25+
- Docker, optional

Run checks:

```powershell
gofmt -w cmd internal
go test ./...
go build ./cmd/openreview
```

Or, if `make` is available:

```powershell
make check
```

Run the service:

```powershell
go run ./cmd/openreview
```

## Contribution Workflow

1. Open or comment on an issue for anything non-trivial.
2. Keep PRs focused. Small PRs are easier to review and merge.
3. Add or update tests for behavior changes.
4. Update docs when changing configuration, endpoints, prompts, or review behavior.
5. Run formatting and tests before opening the PR.

## Engineering Principles

- Prefer boring, explicit Go over clever abstractions.
- Keep provider integrations behind interfaces.
- Keep prompt orchestration data-driven where possible.
- Treat security review as defensive code review, not autonomous exploitation.
- Do not add network calls, background workers, persistence, or external dependencies without a clear product reason.
- Preserve self-hosting and local-model paths.

## Prompt Contributions

Prompt changes are product changes. They affect review quality, false positives, and user trust.

When adding or changing prompts:

- keep them defensive
- require structured JSON output where possible
- include concrete severity guidance
- avoid instructions for live exploitation
- avoid leaking secrets or asking models to extract credentials
- add tests for prompt loading, include expansion, or rendering when applicable
- update [docs/prompts.md](docs/prompts.md)

## Security-Sensitive Contributions

Changes to webhook handling, prompt safety, provider clients, authentication, authorization, storage, or GitHub comment posting should include threat-model thinking in the PR description.

At minimum, explain:

- what trust boundary is affected
- what user or repository data is handled
- how secrets are protected
- what happens on malformed input or provider failure

## Commit Style

Use clear, imperative commit messages:

```text
Add prompt renderer include validation
Fix GitHub webhook signature parsing
Document review profile format
```

For changes that should be committed one file at a time:

```powershell
.\scripts\commit-each-file.ps1
```

Preview the commits without writing them:

```powershell
.\scripts\commit-each-file.ps1 -WhatIf
```

Or, if `make` is available:

```powershell
make commit-each-file
```

## Pull Request Checklist

- Tests pass with `go test ./...`
- Code is formatted with `gofmt`
- New behavior is documented
- Security implications are described when relevant
- The change is scoped to one concern
