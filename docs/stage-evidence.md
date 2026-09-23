# Build evidence

Build started 2026-09-23 12:18 UTC in an isolated checkout, branch
`codex/butterfly-build`. The other session owns Brev setup. No live model,
Telegram delivery or public deployment is implied by local test results.

## S2 — application skeleton

- FastAPI starts through `PORT=8765 sh scripts/run.sh`.
- `GET /health` returns HTTP 200, with both provider configuration flags false
  and `live_verified=false` when no endpoints are supplied.
- `pytest tests/test_bootstrap.py -q`: **2 passed**. Tests cover provider-free
  startup and rejection of wildcard CORS with credentials.
- Python dependencies locked in `requirements.txt`; runtime data and secrets ignored.
- Environment note: sandbox TestClient event-loop operations hang; the same
  isolated local tests pass outside the sandbox. Starlette emits a non-failing
  deprecation warning about its httpx-based TestClient.
- Exit gate passed for the application skeleton. Live provider hookup is deferred
  to S4 by the user's instruction; Brev is being configured separately.

## S3 — Go vertical slice

User requested reuse of plugin MCP/Fly instead of a Python server. HTTP, session
ownership, file storage, review and provider adapters are Go. Python remains only
for local PDF/DOCX and the existing Riva SDK. Frontend follows the supplied PDF.

Fresh release smoke against the running Go server passed: session → synthetic
meeting → draft PDF blocked → per-action review → confirm → PDF/DOCX/JSON exports
→ second-session 404 → authenticated ready-plugin Fly HTML → deletion.
MCP system.context initialized the actual installed process; real reports.create
and routing.fly_feedback produced an anonymous audit artifact and Fly log events.
Export tests: 10 passed. node --check: passed. Go API tests include concurrent
confirmation and no duplicate callback. Real ASR/LLM is not inferred from demo.

Deadline changed by user to 18:00 Almaty (13:00 UTC), with 15 minutes remaining.
The user also explicitly requested publishing this build to main. Remaining time
is reserved for integration verification, release documentation and push; model
reinstallation and optional features are out of scope for this cut.

## S4 — integration hardening

Independent performance review preceded architecture review (docs/final-review.md).
Seven release tests cover Telegram ownership/configuration/confirmation, concurrent
send deduplication, uncertain delivery, Fly route allowlist, CORS/body bounds and
literal environment parsing. No live Telegram message was sent by tests.
`go test -race ./...` and `go vet ./...` passed after provider integration and
isolation fixes. Hosted AI keys and POSTGRES_DSN are cleared in plugin children.
Three.js 0.160.0 and its license are now local assets for the existing Fly view.

Coverage remains below the full default 80% gate; exact measurements are in the
review. A public HTTPS Brev backend and Vercel E2E are not verified. The supplied
other-session Riva/Nemotron smoke is preserved as historical evidence, not relabeled
as a new application run. S3 was published to main as ac6f07c with fast-forward push.

## S5 / S6 — deadline release

Formal judge scoring was cut by the new hard deadline; no score was invented.
Clean checkout of e22518b passed Go build/tests, 10 export tests, JS syntax and the
same real-server synthetic review/export/ownership/Fly scenario. Source archive
SHA256: 1245b966d4a797e6c580f054a4e32bd69c1d6b105b922ca0d31c5b6682e023a9.
Both S3 and S4 commits reached GitHub main with fast-forward pushes.

The original S6 public Vercel+Brev E2E gate remains failed/unverified. This release
is runnable locally; it is not a claim that a public Brev deployment, diarization,
KK/mixed quality or production readiness has been completed. The supplied VM
smoke belongs to the separate model setup session.

Final deadline additions: Dockerfile/compose.yaml with persistent data and optional
existing Brev network/plugin profiles; Docker CLI unavailable, so image build and
container startup are explicitly unverified. scripts/check.sh passed all Go race,
build/vet, 10 Python export tests and frontend syntax. Regression eval against the
running clean-checkout server: 15 passed, 0 failed, 0 skipped; model accuracy and
judge scores remain unverified. GitHub Actions workflow added (remote run not yet
observed). Fly vendored Three.js fetch/import check passed.
