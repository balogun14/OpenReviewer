# Contributor Workflows

This document covers common development workflows for contributors.

## Basic Checks

Run before opening a PR:

```powershell
gofmt -w cmd internal
go test ./...
go build ./cmd/openreview
```

Or:

```powershell
make check
```

## Per-File Commits

The project has a helper for committing every changed file separately:

```powershell
.\scripts\commit-each-file.ps1 -WhatIf
.\scripts\commit-each-file.ps1
```

With make:

```powershell
make commit-each-file
```

## Testing GitHub Webhook Logic

Most GitHub behavior should be tested without real GitHub network calls.

Use:

- `httptest.Server` for GitHub API client tests
- fake implementations of the `httpapi.gitHubClient` interface for webhook tests
- real webhook payload fixtures when adding new event handling

Relevant tests:

- `internal/github/client_test.go`
- `internal/github/diff_test.go`
- `internal/httpapi/server_test.go`

## Testing Providers

Provider tests should not call real LLM APIs.

Use `httptest.Server` and assert:

- request path
- auth headers
- model selection
- response format
- prompt body
- retry behavior
- malformed response handling

Relevant tests:

- `internal/provider/openai_compatible_test.go`
- `internal/provider/retry_test.go`
- `internal/llm/response_test.go`

## Adding GitHub App Features

When changing GitHub App behavior, update:

- `docs/github-app.md`
- `IMPLEMENTATION_PLAN.md`, if the roadmap changes
- tests in `internal/github` or `internal/httpapi`

Security-sensitive changes should explain:

- what GitHub permission is needed
- what repository data is accessed
- what gets posted back to GitHub
- how duplicate webhook delivery is handled

## Adding Provider Features

Provider integrations should live behind the provider interface.

Avoid provider-specific behavior leaking into the review engine. The review engine should know about prompts and findings, not API-specific payloads.

## Adding Prompt Features

Prompt changes should preserve defensive behavior.

Do not add prompts that instruct the model to:

- perform live exploitation
- extract credentials
- run destructive tests
- generate weaponized exploit chains

Use minimal payload examples only when they help explain defensive impact.
