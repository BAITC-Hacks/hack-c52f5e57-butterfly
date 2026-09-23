FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/butterfly ./cmd/butterfly

FROM python:3.12-slim-bookworm AS runtime
ARG INSTALL_RIVA=0
ARG RIVA_CLIENT_VERSION=2.27.0
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 \
    BUTTERFLY_HOST=0.0.0.0 PORT=8000 BUTTERFLY_DATA_DIR=/app/data/runtime \
    BUTTERFLY_EXPORT_PYTHON=/usr/local/bin/python RIVA_PYTHON=/usr/local/bin/python
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid 10001 butterfly \
    && useradd --uid 10001 --gid butterfly --no-create-home butterfly
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt \
    && if [ "$INSTALL_RIVA" = 1 ]; then pip install --no-cache-dir "nvidia-riva-client==$RIVA_CLIENT_VERSION"; fi
COPY --from=build /out/butterfly /usr/local/bin/butterfly
COPY frontend ./frontend
COPY butterfly ./butterfly
COPY scripts/export_protocol.py scripts/riva_transcribe.py ./scripts/
COPY demo ./demo
RUN mkdir -p /app/data/runtime && chown -R 10001:10001 /app/data
USER 10001:10001
EXPOSE 8000
HEALTHCHECK --interval=15s --timeout=3s --start-period=15s --retries=3 \
    CMD python -c "import json,urllib.request; assert json.load(urllib.request.urlopen('http://127.0.0.1:8000/health',timeout=2))['status']=='ok'"
ENTRYPOINT ["/usr/local/bin/butterfly"]
