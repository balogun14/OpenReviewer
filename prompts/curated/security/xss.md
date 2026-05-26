# XSS Review

Review the pull request for cross-site scripting risks.

Repository: {{REPO_NAME}}
Pull Request: #{{PR_NUMBER}} {{PR_TITLE}}

## Rules

{{REVIEW_PROFILE_RULES}}

## Diff

```diff
{{DIFF}}
```

## Focus

- unsafe HTML rendering
- user-controlled script, style, URL, or event-handler contexts
- missing output encoding
- unsafe markdown rendering
- unsafe template helpers
- CSP regressions
- DOM XSS risks
