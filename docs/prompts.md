# Prompt Orchestration

OpenReview AI treats `prompts/` as a prompt library. Prompts are selected by review profile and reviewer persona, then rendered into defensive pull request review instructions before being sent to an LLM provider.

The repository includes a small OpenReview-native curated prompt set under `prompts/curated`. Larger imported prompt collections should live under `prompts/vendor` until they are adapted and promoted into the curated set.

## Current Flow

```text
Review Profile
  -> Reviewer Personas
  -> Prompt Plan
  -> Prompt Renderer
  -> Provider Request
  -> Normalized Findings
```

## Prompt Registry

The curated MVP registry lives in `internal/prompt/manifest.go`.

Initial prompt IDs:

- `code.production-readiness`
- `security.senior-engineering`
- `security.go`
- `security.injection`
- `security.auth`
- `security.authz`
- `security.ssrf`
- `security.xss`

## Include Support

Prompt files may use:

```text
@include(shared/_rules.txt)
```

Includes are resolved relative to the current prompt file and are blocked from escaping the configured prompt base directory.

## Variables

The renderer supports both forms:

```text
{{REPO_NAME}}
{PR_TITLE}
```

Common variables include:

- `REPO_NAME`
- `PR_NUMBER`
- `PR_TITLE`
- `PR_DESCRIPTION`
- `DIFF`
- `REVIEW_PROFILE_RULES`
- `PERSONA_ID`
- `PERSONA_NAME`
- `PERSONA_INSTRUCTIONS`
- `OUTPUT_SCHEMA`

## Defensive Wrapper

Every rendered prompt is wrapped with OpenReview AI policy text:

- defensive PR review only
- no live exploitation
- no destructive testing
- no credential extraction
- exploit payloads only as minimal defensive examples

This lets us reuse strong security prompts while keeping the product aligned with automated code review rather than autonomous penetration testing.

## Output Contract

Every prompt asks the model to return JSON containing:

- `findings`
- `summary`
- `recommendation`

Provider clients should validate and normalize this JSON before comments are posted to GitHub.
