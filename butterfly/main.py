"""HTTP entry point. Provider setup is independent of application startup."""

import os
from pathlib import Path

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware


def create_app(data_dir: str | Path | None = None) -> FastAPI:
    app = FastAPI(title="Butterfly", version="0.1.0")
    app.state.data_dir = Path(data_dir or os.getenv("BUTTERFLY_DATA_DIR", "data/runtime"))
    app.state.data_dir.mkdir(parents=True, exist_ok=True)
    origins = [v.strip() for v in os.getenv("BUTTERFLY_CORS_ORIGINS", "").split(",") if v.strip()]
    if origins:
        if "*" in origins:
            raise ValueError("Use explicit CORS origins when credentials are enabled")
        app.add_middleware(CORSMiddleware, allow_origins=origins, allow_credentials=True,
                           allow_methods=["GET", "POST", "PATCH", "DELETE"],
                           allow_headers=["Content-Type", "Authorization"])

    @app.get("/health")
    def health():
        return {"status": "ok", "version": "0.1.0", "provider": {
            "llm_configured": bool(os.getenv("NEMOTRON_BASE_URL")),
            "asr_configured": bool(os.getenv("ASR_BASE_URL")),
            "live_verified": False,
        }}

    @app.get("/")
    def index():
        return {"app": "Butterfly", "stage": "S2", "health": "/health"}

    return app


app = create_app()
