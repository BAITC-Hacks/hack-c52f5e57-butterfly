# Butterfly application

The app is a modular monolith: FastAPI serves the browser workspace and an
ownership-scoped meeting API. SQLite and uploaded files live under
`BUTTERFLY_DATA_DIR` (ignored by git). Each browser session has an opaque,
HTTP-only owner cookie; knowing a meeting ID is not authorization.

The browser shows transcript segments, editable action items, evidence and job
events. PDF/DOCX exports are generated locally. Fly visualizes the same events
and navigates to the next review item. Telegram is a separately configured
notification consumer; it never receives transcript contents by default.

Provider adapters accept the endpoint and exact model supplied by the separate
Brev preparation session. Unconfigured ASR leaves audio awaiting a provider;
manual transcripts and clearly labeled synthetic demos remain usable. There is
no automatic cloud inference fallback. Configuration presence is not a health
or quality claim.

Stage gates: S2 startup/health; S3 owned meeting flow and local exports; S4
integration/error tests and provider hookup; S5 independent reviews/tests;
S6 reproducible snapshot and demo preparation. Each stage receives its own
commit. Live ASR/Nemotron, Telegram delivery and hosted deployment are recorded
as unverified until their actual checks run.
