# GitHub App Setup

This guide explains how to run OpenReview AI as a GitHub App.

OpenReview AI is still early-stage. The current GitHub App path can receive pull request webhooks, fetch PR patches, run the review engine, post inline comments, post a summary comment, and create a check run.

## 1. Create a GitHub App

In GitHub, go to:

```text
Settings -> Developer settings -> GitHub Apps -> New GitHub App
```

Use:

- **GitHub App name:** `OpenReview AI` or your own deployment name
- **Homepage URL:** your project or deployment URL
- **Webhook URL:** `https://YOUR_DOMAIN/webhooks/github`
- **Webhook secret:** a long random secret

For local development, use a tunnel such as ngrok, Cloudflare Tunnel, or GitHub's local webhook tooling:

```text
https://YOUR_TUNNEL_URL/webhooks/github
```

## 2. App Permissions

Set these repository permissions:

| Permission | Access | Why |
| --- | --- | --- |
| Contents | Read-only | Fetch pull request files and patches |
| Pull requests | Read and write | Read PR metadata and post inline review comments |
| Issues | Read and write | Post PR summary comments |
| Checks | Read and write | Create OpenReview AI check runs |
| Metadata | Read-only | Required by GitHub Apps |

## 3. Subscribe To Events

Subscribe to:

- Pull request

OpenReview AI currently reviews these pull request actions:

- `opened`
- `synchronize`
- `reopened`
- `ready_for_review`

## 4. Generate A Private Key

After creating the app, generate a private key from the GitHub App settings page.

Save the `.pem` file somewhere safe. Do not commit it.

## 5. Install The App

Install the GitHub App on a repository or organization.

The app must be installed on any repository you want OpenReview AI to review.

## 6. Configure OpenReview AI

Required environment variables:

```powershell
$env:GITHUB_APP_ID="123456"
$env:GITHUB_APP_PRIVATE_KEY_PATH="C:\path\to\github-app-private-key.pem"
$env:GITHUB_WEBHOOK_SECRET="your-webhook-secret"
```

Provider configuration:

```powershell
$env:OPENREVIEW_PROVIDER="openrouter"
$env:OPENROUTER_API_KEY="..."
$env:OPENREVIEW_MODEL="anthropic/claude-sonnet-4"
```

Or use a local/OpenAI-compatible provider:

```powershell
$env:OPENREVIEW_PROVIDER="openai-compatible"
$env:OPENREVIEW_PROVIDER_BASE_URL="http://localhost:1234/v1"
$env:OPENREVIEW_PROVIDER_API_KEY="local-key"
$env:OPENREVIEW_MODEL="local-model"
```

For quick local testing without a real LLM:

```powershell
$env:OPENREVIEW_PROVIDER="mock"
```

## 7. Run The Server

```powershell
go run ./cmd/openreview
```

By default the server listens on:

```text
http://localhost:8080
```

Override the address:

```powershell
$env:OPENREVIEW_ADDR=":8081"
go run ./cmd/openreview
```

## 8. Local Webhook Testing

Expose the local server:

```powershell
ngrok http 8080
```

Set the GitHub App webhook URL to:

```text
https://YOUR_NGROK_DOMAIN/webhooks/github
```

If GitHub posts to `/` instead, ngrok will show `POST /`. OpenReview AI accepts this as a fallback, but the canonical webhook URL is still `/webhooks/github`.

Then open or update a pull request in an installed repository.

Expected behavior:

1. GitHub sends a pull request webhook.
2. OpenReview AI validates the webhook signature.
3. OpenReview AI creates an installation token.
4. OpenReview AI fetches PR metadata and changed files.
5. OpenReview AI reviews the diff.
6. OpenReview AI posts eligible inline comments.
7. OpenReview AI posts a summary comment.
8. OpenReview AI creates a check run.

## Troubleshooting

### Webhook returns `invalid_signature`

Check:

- `GITHUB_WEBHOOK_SECRET` matches the GitHub App webhook secret
- the tunnel forwards the raw request body unchanged

### Webhook returns `prepare_review_failed`

Check:

- `GITHUB_APP_ID` is set
- `GITHUB_APP_PRIVATE_KEY_PATH` points to a readable `.pem` file
- the app is installed on the repository
- app permissions include Contents, Pull requests, Issues, and Checks

### Review runs but there are no inline comments

Inline comments are only posted when findings map to changed lines in reviewable files. Findings without a file/line, or findings on unchanged lines, appear only in the summary.

### Provider errors

For OpenRouter:

```powershell
$env:OPENREVIEW_PROVIDER="openrouter"
$env:OPENROUTER_API_KEY="..."
$env:OPENREVIEW_MODEL="anthropic/claude-sonnet-4"
```

For OpenAI-compatible APIs, make sure the base URL does not include `/chat/completions`; OpenReview AI appends that path automatically.
