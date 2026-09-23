# DevOps report - Butterfly Nemotron launch preparation

## Decision and inspected stack

2026-09-23: owner selected the ready Nemotron Voice Agent Launchable and prioritizes speed. Official source inspected at 13801d9963aea20ac663318060824a78b28809fa (notebooks/brev_launchable.ipynb, docker-compose.yml, multilingual config and resolvers). Intended profile: multilingual-assistant/single-gpu; self-hosted NeMo-Speech.cpp and vLLM. Brev card showed RTX PRO Server 6000 96 GB, AWS, 8 CPU, 64 GiB RAM, 500 GiB storage and $3.43/hr. Owner subsequently reported a created Running/Built VM nemotron-voice-agent-402b8f with the same GPU/CPU/RAM/disk and an actual displayed instance price of $4.12/hr. This supersedes the preview price. Runtime readiness is still unverified.

## Configuration and checks

Updated docs/deployment/brev-gpu.md and README.md replace the earlier L4 recommendation. HF_TOKEN is entered privately on Brev for model downloads; NVIDIA_API_KEY must be cleared from process environment and .env to avoid external fallback. PIPELINE_TLS=false applies only behind Brev HTTPS; RESET_VOLUMES=False. Required runtime checks: Docker Compose config, nvidia-smi, application /health on port7860, local vLLM /health and /v1/models on18000. GPU/Docker preflight ran successfully in Jupyter: nvidia-smi reported RTX PRO 6000 Blackwell Server Edition,97887MiB,driver595.91.07; Docker29.8.1; Composev5.5.1; cloned HEAD matched13801d9963aea20ac663318060824a78b28809fa. All subprocess checks exited0. Model/service health checks have not run. No plugin or application source code changed, so a Go quality gate would not verify this deployment.

## Test data

Two MP3 recordings and two reference protocols were downloaded from user-provided Google links into gitignored data/test/meeting-minutes. Sizes and SHA256 hashes are in the local manifest. Approximate audio lengths:274.32 and206.09seconds. Protocol language is Russian; audio language is not yet verified. References contain10 and6action items, and five distinct speaker labels each. Meeting dates and alignment timestamps are absent. Personal contents and recordings are not included in public Git.

## Evidence and gaps

git check-ignore data/test/meeting-minutes/meeting-1.mp3 data/test/meeting-minutes/meeting-1-reference.txt exited0 and returned both paths. The owner created the VM and completed NVIDIA sign-in. Jupyter access works; repository cloned and multilingual profile selected. Cell2b is waiting for private HF_TOKEN entry. Model downloads, service health and inference have not been verified. Russian is present in ASR but may be excluded by stock voice UI language intersection; Kazakh needs a supplemental ASR path. Upstream stock voice agent has no meeting upload/minutes workflow; meeting adapter, five-speaker diarization, exports, Fly, Telegram and final Vercel deployment remain implementation work.

## Sources

https://github.com/NVIDIA-AI-Blueprints/nemotron-voice-agent/tree/13801d9963aea20ac663318060824a78b28809fa
https://brev.nvidia.com/launchable/deploy?launchableID=env-3GGgfIXCnso410yatOZD24ufRJ2
https://github.com/NVIDIA/NeMo-Speech.cpp/blob/main/docs/cli.md
https://huggingface.co/nvidia/nemotron-3.5-asr-streaming-0.6b
