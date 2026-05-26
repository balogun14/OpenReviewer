# Go Security Review

Perform a defensive Go security review for this pull request.

Repository: {{REPO_NAME}}
Pull Request: #{{PR_NUMBER}} {{PR_TITLE}}

## Rules

{{REVIEW_PROFILE_RULES}}

## Diff

```diff
{{DIFF}}
```

## Go-Specific Checks

- context propagation and cancellation
- ignored errors
- unsafe file path handling
- SQL or command injection
- SSRF via outbound HTTP calls
- goroutine leaks
- race-prone shared state
- insecure cryptography
- secret leakage in code or logs
- missing request size limits or timeouts
