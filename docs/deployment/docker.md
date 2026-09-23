# Docker runtime

The image runs the Go HTTP server as UID 10001. Python is limited to local PDF/DOCX export and the optional Riva SDK adapter. No models are downloaded or reinstalled by Compose. Meeting data persists in the private `butterfly-data` volume; model ports are not published.

## Local manual and synthetic demo

```sh
docker compose config --quiet
docker compose up -d --build --wait backend
sh scripts/docker-smoke.sh backend
```

Open http://127.0.0.1:8000. A synthetic demo or manual transcript works without provider credentials. The service binds only to host loopback. Use an HTTPS reverse proxy for remote browser access; set `BUTTERFLY_PUBLIC_URL`, exact `BUTTERFLY_CORS_ORIGINS`, and secure cookies for that deployment. A configured provider is not proof of successful live processing.

## Reuse the running Brev stack

Run these commands **on the Brev VM**. Read `nemotron-integration-handoff.md` first and discover the existing model container's Docker network. Set `BREV_DOCKER_NETWORK` to its exact name; Compose will not create it. Stop the local demo backend if it owns the same port.

```sh
export BREV_DOCKER_NETWORK=YOUR_EXISTING_MODEL_NETWORK
docker compose stop backend
docker compose --profile brev up -d --build --wait backend-brev
sh scripts/docker-smoke.sh backend-brev
```

The `brev` profile installs `nvidia-riva-client==2.27.0` (version confirmed from PyPI metadata), reaches `nemo-speech-multilingual:50051` and `http://nvidia-llm-vllm:8000/v1`, and selects `nvidia/nemotron-3.5-lightning-30b-a3b`. Override the corresponding environment variables only if those actual addresses differ. It uses Riva gRPC and ffmpeg PCM16 mono 16 kHz conversion. Recognition by speaker is not implemented.

## Existing HackAlem MCP and Fly demo

Mount an existing compatible Linux HackAlem checkout with executable `bin/hackalem-mcp` and `bin/hackalem-demo-api`; the checkout is read-only. A separate private runtime directory is created in the data volume.

```sh
export HACKALEM_PLUGIN_PATH=/absolute/path/to/hackalem
docker compose stop backend
docker compose --profile plugin up -d --build --wait backend-plugin
```

To combine the Brev connection and plugin mount in one foreground container:

```sh
docker compose stop backend backend-brev backend-plugin
docker compose --profile brev run --rm --service-ports \
  -v "$HACKALEM_PLUGIN_PATH:/opt/hackalem:ro" \
  -e HACKALEM_HOME=/opt/hackalem backend-brev
```

Only start one backend service at a time: they share a port and storage. Telegram is enabled only when the existing MCP and explicit token/chat allowlist are configured. Secrets can be supplied through the local `.env` or environment; `.env` is excluded from the build context and never copied into the image.

## Verification status

This development host has no Docker CLI. Docker image build, engine startup, container health, and live model access are therefore **unverified here**. `scripts/docker-smoke.sh` exits 2 with an explicit message when Docker is absent. The documented native fallback is `sh scripts/run.sh` with the export dependencies installed. Do not treat a static YAML parse as a running container check.
