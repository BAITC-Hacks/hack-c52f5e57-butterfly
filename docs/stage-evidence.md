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
