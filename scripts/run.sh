#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
PYTHON_BIN=${BUTTERFLY_PYTHON:-.venv/bin/python}
if [ -f .env ]; then
  exec "$PYTHON_BIN" -m uvicorn butterfly.main:app --env-file .env --host "${BUTTERFLY_HOST:-127.0.0.1}" --port "${PORT:-8000}"
fi
exec "$PYTHON_BIN" -m uvicorn butterfly.main:app --host "${BUTTERFLY_HOST:-127.0.0.1}" --port "${PORT:-8000}"
