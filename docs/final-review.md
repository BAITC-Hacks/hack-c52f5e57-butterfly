# Release review

Reviewed the Go HTTP entry point, persistence and provider path, MCP bridge, Fly proxy and explicit Telegram notification path. Provider integration was being completed concurrently; these results describe the local test run, not a verified public Brev deployment.

## Performance review, then architecture review

The provider semaphore bounds model work to one meeting at a time. Provider requests have bounded response size and timeouts; upload size is capped. MCP operations use a bounded queue and serialized, time-limited protocol calls. PDF/DOCX generation uses a bounded subprocess lifetime. These limits suit the single-user demonstration.

The state store serializes and syncs the whole JSON state on every mutation. Meeting lists clone full transcripts before returning at most 100 results. Retry currently holds the global mutation lock while contacting providers, so another meeting's edit/confirm/delete may wait for that call. These are explicit scaling limitations, not measured throughput claims; no speculative optimization was added during release preparation.

Application state and ownership remain in `internal/app`; provider access stays behind server-side configuration. `internal/integrations` composes the installed MCP and Fly demo with the app, and the browser cannot choose arbitrary MCP tools. Fly forwards only five visualization routes and requires a valid session. Anonymous workflow metadata reaches Fly; transcript content is excluded from the review audit.

Telegram requires an owned, confirmed meeting and configured channel. An exclusive persistent reservation permits only one send even for concurrent clicks. An uncertain delivery result remains reserved and is not retried automatically. Release tests use a mock caller: no Telegram message was sent by the test suite.

## Evidence

- `GOCACHE=/tmp/butterfly-go-cache go test -race ./...`: all four Go packages passed.
- `GOCACHE=/tmp/butterfly-go-cache go vet ./...`: passed.
- `GOCACHE=/tmp/butterfly-go-cache go test -cover ./...`: command 56.2%, app 73.1%, integrations 29.4%, MCP bridge 81.0%. The full 80% coverage gate is **not satisfied**.
- Added seven release tests covering notification ownership, confirmation/configuration requirements, concurrent deduplication, uncertain delivery, Fly route/session restrictions, cross-origin mutations, request limits and literal dotenv parsing.
- Existing app tests cover private audio/export access, recovery, review/confirmation idempotency, concurrent confirmation, unknown ownership/deadline and document export after confirmation.

## Deployment checks that remain separate

Set `BUTTERFLY_CORS_ORIGINS` to the public HTTPS frontend origin when TLS terminates at a proxy; set secure cookie configuration explicitly. Review found a CDN dependency in the plugin page; the release now vendors Three.js 0.160.0 with its license and rewrites that import in the authenticated proxy. A configured provider flag does not establish live verification. Public HTTPS upload → Riva → Nemotron → review → PDF/DOCX remains a deployment smoke check, distinct from mocked provider tests.

Review found inherited `POSTGRES_DSN` in the MCP child; the release now clears it together with hosted-model keys, matching isolation of the Fly process. Provider health now recognizes the supplied Riva and LOCAL_LLM aliases. Race tests and vet were repeated after these fixes.
