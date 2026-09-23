# Provider contract

The current implementation is Go; its existing Brev integration is documented in
[the adapter runbook](deployment/butterfly-brev-adapter.md). Source VM parameters
and the other session's measured smoke are retained in
[runtime-handoff.json](deployment/runtime-handoff.json).

The prepared ASR uses **Riva gRPC**, not an HTTP audio/transcriptions endpoint.
Set `RIVA_ASR_ENDPOINT` and `RIVA_PYTHON` to the interpreter with the installed
Riva SDK. A bounded local helper decodes PCM16 mono 16 kHz with ffmpeg and calls
Riva in chunks. Go still owns HTTP, authorization, storage and workflow.
`LOCAL_LLM_BASE_URL` and `LOCAL_LLM_MODEL` configure the existing Nemotron HTTP
endpoint; `NEMOTRON_*` aliases take precedence. Requests specify json_object,
enable_thinking=false, temperature=0 and require finish_reason=stop.

Optional generic self-hosted ASR endpoints use `ASR_BASE_URL`, `ASR_MODEL` and
`ASR_API_KEY`, with multipart `/audio/transcriptions`. They are not the configured
Brev speech service. Known hosted OpenAI/NVIDIA API addresses are rejected.

Every action must reference a real source segment and exact quote. Unknown owner
or ISO date stays empty; relative deadlines require human review. Riva speaker
identity is explicitly unknown. No diarization or mixed-language claim is made.
Manual transcripts and the clearly labeled synthetic demo work without models.

A browser session owns meetings, files and exports. PATCH edits invalidate prior
completion. Confirm requires every action to be reviewed. PDF/DOCX require a
completed meeting; JSON remains available for drafts. Telegram only sends a generic
status after an explicit click on a completed, owned meeting.

Mock tests prove adapter contracts. The other session's 30-second ASR→LLM smoke
proves the models in that VM. Neither proves the newly built application's public
upload→Brev→review→download path; that remains a deployment acceptance check.
