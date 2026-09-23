from fastapi.testclient import TestClient
from butterfly.main import create_app


def test_health_is_available_without_gpu_or_provider_secrets(tmp_path, monkeypatch):
    monkeypatch.delenv("NEMOTRON_BASE_URL", raising=False)
    monkeypatch.delenv("ASR_BASE_URL", raising=False)
    with TestClient(create_app(tmp_path)) as client:
        health = client.get("/health")
        assert health.status_code == 200
        assert health.json()["provider"] == {
            "llm_configured": False, "asr_configured": False, "live_verified": False,
        }
        assert client.get("/").status_code == 200


def test_cors_requires_explicit_origin(tmp_path, monkeypatch):
    import pytest
    monkeypatch.setenv("BUTTERFLY_CORS_ORIGINS", "*")
    with pytest.raises(ValueError, match="explicit CORS"):
        create_app(tmp_path)
