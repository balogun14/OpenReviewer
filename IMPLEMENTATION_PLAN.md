# Implementation Plan

This plan turns OpenReview AI from the current Go MVP into a useful self-hosted GitHub App.

The guiding principle is to make the core review loop real before building dashboards or advanced workflow features.

## Current Foundation

- [x] Go service scaffold
- [x] Docker Compose setup
- [x] GitHub webhook endpoint
- [x] Webhook signature validation
- [x] Review engine
- [x] Reviewer personas
- [x] Prompt registry
- [x] Safe prompt loading
- [x] `@include(...)` prompt expansion
- [x] Prompt variable interpolation
- [x] Defensive prompt wrapper
- [x] Mock provider
- [x] Basic normalized finding model
- [x] Open-source project docs and GitHub templates

## Phase 1: Real LLM Reviews

Goal: replace mock review behavior with real provider-backed structured reviews.

- [x] Define structured LLM response schema
- [x] Add JSON response parser
- [x] Add response validation
- [x] Normalize LLM findings into internal findings
- [x] Handle malformed model responses
- [x] Add provider retry policy
- [x] Add OpenRouter provider
- [x] Add OpenAI-compatible provider
- [x] Add provider configuration from environment
- [x] Add tests for response parsing and validation
- [x] Add tests for provider request construction

Exit criteria:

- [x] A local request can run selected prompts against a real LLM provider
- [x] Provider responses become normalized findings
- [x] Invalid provider responses fail safely

## Phase 2: GitHub PR Integration

Goal: review actual pull requests and publish useful GitHub feedback.

- [ ] Add GitHub App configuration
- [ ] Implement installation token generation
- [ ] Fetch pull request metadata
- [ ] Fetch changed files and patches
- [ ] Detect supported file types
- [ ] Ignore generated files and lockfiles
- [ ] Chunk large diffs
- [ ] Map findings to changed lines
- [ ] Post inline PR comments
- [ ] Post review summary comment
- [ ] Create or update check runs
- [ ] Add idempotency for repeated webhook delivery
- [ ] Add tests for GitHub payload handling
- [ ] Add tests for changed-line mapping

Exit criteria:

- [ ] Opening or updating a PR triggers a review
- [ ] Findings appear as GitHub comments
- [ ] A summary comment includes severity counts and recommendation

## Phase 3: Repository Configuration

Goal: make repository-owned review standards the core differentiator.

- [ ] Define `openreview.yml` schema
- [ ] Load repository rules from default branch
- [ ] Load repository rules from PR branch when safe
- [ ] Support architecture rules
- [ ] Support review profiles
- [ ] Support reviewer personas
- [ ] Support prompt selection
- [ ] Support ignored paths
- [ ] Support generated-file patterns
- [ ] Support provider/model overrides
- [ ] Inject repository rules into prompts
- [ ] Add config validation errors
- [ ] Add tests for config parsing

Example target:

```yaml
project:
  language: go
  framework: net-http

review:
  profile: security-first

reviewers:
  - security-engineer
  - staff-backend-engineer

rules:
  - handlers must propagate request context
  - database queries must use parameter binding
  - external HTTP calls must have timeouts

ignore:
  - "**/*.generated.go"
  - "vendor/**"
```

Exit criteria:

- [ ] A repository can control review behavior through `openreview.yml`
- [ ] Config errors are visible and actionable
- [ ] Repository rules are included in rendered prompts

## Phase 4: Persistence and Queueing

Goal: make reviews reliable enough for production self-hosting.

- [ ] Add PostgreSQL schema
- [ ] Store organizations
- [ ] Store repositories
- [ ] Store pull requests
- [ ] Store reviews
- [ ] Store findings
- [ ] Store review profiles
- [ ] Add database migrations
- [ ] Add review job queue
- [ ] Add retries
- [ ] Add job idempotency
- [ ] Add audit logs
- [ ] Track provider latency
- [ ] Track provider token usage and cost
- [ ] Add integration tests for persistence

Exit criteria:

- [ ] Reviews survive process restarts
- [ ] Failed jobs can retry safely
- [ ] Duplicate webhooks do not create duplicate reviews or comments

## Phase 5: Prompt Library Cleanup

Goal: turn the imported prompt folder into a maintainable OpenReview-native prompt system.

- [x] Create initial `prompts/curated` OpenReview-native prompt set
- [ ] Split raw imported prompts from curated OpenReview prompts
- [ ] Move raw imported prompts to `prompts/vendor`
- [ ] Add prompt metadata files
- [ ] Add prompt ownership/category metadata
- [ ] Add prompt safety classification
- [ ] Adapt offensive security prompts into defensive review prompts
- [ ] Remove or quarantine prompts unsuitable for PR review
- [ ] Add golden prompt rendering tests
- [ ] Add sample diffs and expected finding fixtures
- [ ] Document prompt contribution rules

Target structure:

```text
prompts/
  curated/
    code/
    security/
    architecture/
    performance/
  vendor/
    shannon/
    community/
```

Exit criteria:

- [ ] Every active prompt has metadata
- [ ] Every active prompt is defensive-review safe
- [ ] Prompt changes can be tested before merge

## Phase 6: Dashboard

Goal: provide a useful UI after the backend review loop is solid.

- [ ] Add Next.js app
- [ ] Repository list
- [ ] Review history
- [ ] Review detail page
- [ ] Finding detail page
- [ ] Review profile editor
- [ ] Provider configuration view
- [ ] Installation/setup guide
- [ ] Finding feedback workflow
- [ ] False-positive tracking
- [ ] Basic usage and cost analytics

Exit criteria:

- [ ] Maintainers can inspect reviews outside GitHub
- [ ] Teams can manage review profiles without editing YAML manually
- [ ] Finding feedback can inform later review behavior

## Recommended Next Task

Start with Phase 1:

- [ ] define the structured LLM response schema
- [ ] implement parser and validator
- [ ] add OpenRouter provider

That creates the first real end-to-end AI review loop and gives the project something concrete to test against real pull request diffs.
