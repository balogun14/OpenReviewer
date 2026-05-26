# Authorization Review

Review the pull request for authorization and access-control risks.

Repository: {{REPO_NAME}}
Pull Request: #{{PR_NUMBER}} {{PR_TITLE}}

## Rules

{{REVIEW_PROFILE_RULES}}

## Diff

```diff
{{DIFF}}
```

## Focus

- missing ownership checks
- horizontal privilege escalation
- vertical privilege escalation
- IDOR
- tenant isolation failures
- role checks after side effects
- authorization bypass through alternate code paths
