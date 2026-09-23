# Reproducible evaluation

Run against a disposable, running Butterfly instance. The suite creates synthetic meetings in new sessions, checks them, then deletes them. It does not send Telegram notifications. With no providers configured it also checks honest unavailable-provider behavior.

```bash
node scripts/eval.mjs --base http://127.0.0.1:8000 --out data/results/eval.json
```

The JSON report contains individual pass/fail/skip results, elapsed times, timestamps and the client checkout revision. Exit code is 1 if any check fails. Runtime reports are ignored by Git. The server's revision is not attested by the client revision field; launch the server from the checkout being evaluated. If a provider is configured, the provider-isolation checks are explicitly skipped to avoid unrequested inference requests.

Coverage: explicit synthetic labels; evidence linked to source segments; unknown owner/date preservation; review gating; idempotent confirmation; cross-session read/edit/export/delete isolation; PDF, DOCX and JSON exports; manual transcript preservation; unavailable ASR without fabricated results; cleanup.

Recorded local run: **15 passed, 0 failed, 0 skipped** against the clean-checkout server at `http://127.0.0.1:8767`. Output: `data/results/eval.json`. PDF and DOCX signature checks prove export generation, not a complete visual layout review. Other export tests cover multilingual content.

## Model quality and judging

This is an API regression suite, not a claim of ASR accuracy or action-extraction quality. WER, action precision and action recall remain **unverified**. The labelled synthetic fixture is in `demo/eval/manual-unknown.json`; it is never presented as real model output.

Live validation must run the documented `scripts/nemotron_smoke.py` on the Brev VM and retain audio, reference transcript, actual ASR output and extracted actions. A defensible quality score additionally requires a consented, labelled real meeting corpus, independent reference annotations, and a stated matching policy for action text/owner/deadline. Existing VM smoke timings alone do not establish accuracy, diarization or end-to-end web-upload readiness.

The HackAlem judge round is **not run**: no three isolated judges scored one frozen bundle during this build. Missing judge reports are not zero scores. Freeze a stable bundle first, then run three isolated judges and the plugin's aggregate/calibrate/rank/compare protocol before making a judging or readiness score claim.
