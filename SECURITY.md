# Security Policy

OpenReview AI is security-focused software. Please report vulnerabilities responsibly.

## Supported Versions

The project is pre-1.0. Security fixes are applied to the main branch until versioned releases begin.

## Reporting a Vulnerability

Do not open a public issue for vulnerabilities.

Until a dedicated security email is available, report privately to the repository owner or maintainer through GitHub.

Please include:

- affected component
- impact
- reproduction steps
- relevant logs or payloads, if safe to share
- suggested fix, if known

## Scope

Security reports are especially useful for:

- GitHub webhook signature validation
- GitHub App authentication
- provider API key handling
- prompt injection risks
- unsafe prompt includes or path traversal
- secret leakage in logs or findings
- repository data exposure
- unsafe GitHub comment posting

## Project Safety Rules

OpenReview AI must remain defensive by default.

- Do not add autonomous exploitation behavior.
- Do not add destructive testing behavior.
- Do not add credential extraction behavior.
- Do not send repository data to providers unless explicitly configured.
- Do not log secrets, provider keys, webhook secrets, or raw credentials.

Security review prompts may discuss exploitability only to explain defensive impact and remediation.
