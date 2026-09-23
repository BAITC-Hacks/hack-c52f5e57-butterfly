#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
if ! command -v docker >/dev/null 2>&1; then
  echo "Docker CLI unavailable: container verification was not run." >&2
  exit 2
fi
docker compose config --quiet
docker info >/dev/null
service=${1:-backend}
case "$service" in
  backend) ;;
  backend-brev|backend-plugin) ;;
  *) echo "Expected backend, backend-brev or backend-plugin" >&2; exit 2 ;;
esac
docker compose build "$service"
docker compose up -d --wait "$service"
docker compose exec -T "$service" python -c 'import json, urllib.request; h=json.load(urllib.request.urlopen("http://127.0.0.1:8000/health",timeout=3)); assert h["status"]=="ok"; print("HTTP health:",h["status"],"runtime:",h["runtime"])'
