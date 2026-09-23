# Architecture

Go net/http serves the static frontend and ownership-scoped meeting API.
Opaque HTTP-only session cookies isolate records; atomic JSON storage and private
uploads live in BUTTERFLY_DATA_DIR. One process owns the data directory.
Mutation serialization prevents duplicate confirmations and conflicting writes.

The provider layer calls self-hosted Nemotron HTTP and Riva gRPC via a narrow
local SDK subprocess. It validates source quotes and keeps unknown owners/dates
empty for human review. There is no hosted AI fallback. Speaker labels remain
unknown unless a provider supplies them; no diarization claim is inferred.

Confirmed meetings export through a bounded local Python PDF/DOCX helper with
bundled Cyrillic fonts. The HTTP service and storage remain Go. JSON stays usable
without the export helper. Frontend is static HTML, CSS and JavaScript.

Existing HackAlem binaries run against a dedicated private data root. The stdio
MCP bridge exposes an explicit tool allowlist, caps messages, serializes calls and
clears hosted AI keys. Fly receives generic workflow metadata, not transcripts.
Its existing MB 3D view is proxied through authenticated visualization routes;
plugin control and artifact routes are not exposed. Browser Three.js is vendored.
The fuller fly body/mesh demo requires assets absent from the installed plugin.

Telegram uses the existing MCP sender with its allowlist. A user click on a
completed owned meeting reserves a durable send record before contacting Telegram.
Uncertain delivery is not automatically retried. Messages contain no recording,
transcript, action text or participant details. This is a single-team configured
chat, not a multi-tenant bot or subscription product.

MCP failure does not prevent manual review or export. Its bounded event queue
can skip audit events when full, reported in health. The monolith and its JSON
store are a hackathon deployment, not a multi-instance production queue.

Current gates distinguish local synthetic evidence, mocked provider contracts,
the separate session's real VM smoke, and the still unverified public E2E.
