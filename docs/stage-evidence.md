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
