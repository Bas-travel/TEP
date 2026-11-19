---
description: "Review documentation, verify Git safety, and execute the Travel Super App MVP sprint workflow"
name: "Travel Super App MVP"
argument-hint: "Optional phase, such as Phase 0, documentation review, or Sprint 1"
agent: "agent"
---

Work directly in the current repository and use repository-first development for the Travel Super App.

## Inputs

- Optional requested phase: `${input:phase:Phase 0 project audit}`
- Existing architecture documents, source code, migrations, tests, and UI assets.
- Do not rely on assumptions when repository evidence is available.

## Required workflow

1. Audit Git before editing: current branch, remotes, working tree, changed files, and recent commits.
2. Inspect repository structure, technology stack, architecture, documentation, API contracts, database schema, tests, and current sprint state.
3. Review OpenAPI, ERD, MVP scope, naming conventions, authentication, error responses, migrations, acceptance criteria, and UI assets for consistency.
4. Preserve existing work. Do not overwrite unrelated changes, use force push, commit secrets, or modify `.env` files containing secrets.
5. Record documentation defects and implementation gaps in clear findings before implementation.
6. For documentation-only changes, commit a meaningful baseline only after reviewing the diff; obtain explicit authorization before adding code or remote operations if the task is scoped that way.
7. Before pushing, verify branch identity, remote, diff, secret exposure, commit contents, and upstream status. Push only to the appropriate branch.
8. Execute Sprint 1 dependencies in order: Docker Compose, Go API structure, Flutter app, mock HTTP integration, migration/seed baseline, health checks, developer documentation, and acceptance testing.
9. Integrate the supplied UIUX reference as a professional bilingual MVP: reuse existing Flutter widgets and design tokens, implement real navigation and states, and avoid placeholder controls.
10. Validate real behavior with commands and tests. Run static analysis, API tests, Flutter tests, migrations, and available end-to-end checks. Do not claim success without output.

## Mandatory progress artifacts

Maintain or update when applicable:

- `docs/DEVELOPMENT_PROGRESS.md`
- `docs/SYSTEM_ARCHITECTURE.md`
- `docs/DATABASE_DESIGN.md`
- `docs/API_DOCUMENTATION.md`
- `docs/SECURITY_AUDIT.md`
- `docs/TEST_REPORT.md`
- `docs/DEPLOYMENT_GUIDE.md`
- `docs/KNOWN_ISSUES.md`

## Output

Report concise, evidence-based results covering:

- Files created or modified.
- Git and branch status before and after changes.
- Documentation inconsistencies found and fixes applied.
- Sprint implementation completed.
- Exact validation commands and actual results.
- Acceptance criteria status.
- Blockers requiring credentials, services, or business decisions.
- Next actionable items.

Do not mark an unverified task as completed.
