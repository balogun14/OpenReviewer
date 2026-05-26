# SSRF Review

Review the pull request for server-side request forgery risks.

Repository: {{REPO_NAME}}
Pull Request: #{{PR_NUMBER}} {{PR_TITLE}}

## Rules

{{REVIEW_PROFILE_RULES}}

## Diff

```diff
{{DIFF}}
```

## Focus

- user-controlled URLs
- redirects to internal networks
- missing host allowlists
- access to metadata services
- unsafe webhook fetchers
- missing timeouts and response limits
- DNS rebinding exposure
