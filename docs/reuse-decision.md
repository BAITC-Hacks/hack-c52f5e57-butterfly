# Runtime reuse decision

User correction: use the existing HackAlem Fly demo and ready MCP capabilities.

The application HTTP runtime is Go (`net/http`). Its meeting endpoints are a
thin domain adapter because the plugin has no meeting upload/review API.
The existing HackAlem MCP process is reused over stdio JSON-RPC, with a narrow
tool allowlist. The existing demo-api serves Fly; only its visualization routes
are proxied, not control-plane sessions, filesystem listings or arbitrary calls.

Checked source: HackAlem 0.16.66 `cmd/demo-api`, `cmd/hackalem-mcp`,
`cmd/hackalem-rehearse` and `internal/tools`, 2026-09-23. The plugin remains a
separately installed dependency, with a dedicated runtime root for this product.

- Adopt: `routing.fly_recommend`, `routing.fly_feedback`, `reports.create`,
  `telegram.send`, and the existing Fly demo/SSE implementation.
- Build: meeting ownership, review state, self-hosted provider adapters and the
  product-facing API; these are absent from the MCP capability set.
- Keep only as an export helper: Python ReportLab/python-docx, because the
  existing reports capability produces Markdown/HTML, not PDF/DOCX.
- Exclude: MCP `audio.transcribe`, whose current implementation uses hosted
  OpenAI. Nemotron and ASR come from the separately configured Brev endpoint.

S2's Python HTTP prototype is a historical checkpoint and will be removed from
the active runtime when the Go replacement passes the same behavior gates.
No shared history is rewritten. No speculative brain learning or ASR quality
claim is made by embedding Fly.
