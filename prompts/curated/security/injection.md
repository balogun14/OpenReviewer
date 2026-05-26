# Injection Review

Perform defensive source-to-sink review for injection risks.

Repository: {{REPO_NAME}}
Pull Request: #{{PR_NUMBER}} {{PR_TITLE}}

## Rules

{{REVIEW_PROFILE_RULES}}

## Diff

```diff
{{DIFF}}
```

## Focus

Look for untrusted input reaching:

- SQL queries
- shell commands
- file paths
- template expressions
- unsafe deserialization

For each finding, identify the source, sink, missing or mismatched defense, and remediation. Do not provide exploitation steps beyond minimal defensive examples.
