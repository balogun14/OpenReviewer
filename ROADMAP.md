# Roadmap

This roadmap is intentionally practical. The project should become useful as a self-hosted GitHub App before expanding into dashboards, analytics, or advanced learning features.

## Phase 1: Review Core

- Structured LLM JSON response parser and validator
- OpenRouter provider
- OpenAI-compatible provider
- Prompt execution retries and provider error handling
- Finding normalization and deduplication
- Review summary generation

## Phase 2: GitHub App Integration

- GitHub App authentication
- Installation token management
- Pull request file and diff fetching
- Inline PR comments
- PR summary comments
- Check run status reporting

## Phase 3: Repository Configuration

- `openreview.yml` loader
- repository rules injection
- profile selection
- generated-file detection
- file ignore rules
- language/framework detection

## Phase 4: Persistence and Queueing

- PostgreSQL persistence
- review job queue
- retry policy
- idempotency
- audit logs
- cost and latency tracking

## Phase 5: Provider Expansion

- Anthropic direct provider
- OpenAI direct provider
- Gemini provider
- Ollama provider
- LM Studio/OpenAI-compatible local provider

## Phase 6: Product Surface

- Next.js dashboard
- organization/repository views
- review profile editor
- finding feedback workflow
- false-positive tracking
- learning mode

## Not Yet

- CI/CD orchestration
- replacing SAST tools
- autonomous penetration testing
- IDE assistant features
