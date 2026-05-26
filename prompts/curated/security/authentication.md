# Authentication Review

Review the pull request for authentication risks.

Repository: {{REPO_NAME}}
Pull Request: #{{PR_NUMBER}} {{PR_TITLE}}

## Rules

{{REVIEW_PROFILE_RULES}}

## Diff

```diff
{{DIFF}}
```

## Focus

- missing authentication checks
- weak token validation
- missing token expiration checks
- insecure session handling
- insecure password or credential flows
- sensitive authentication data in logs
- replay risks
