package app

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sourceTranscript() []any {
	return []any{map[string]any{"id": "s1", "text": "Анна подготовит отчёт к 2026-09-25."}}
}
func protocolPayload() map[string]any {
	return map[string]any{"summary": "Согласовали отчёт.", "actions": []any{map[string]any{"text": "Подготовить отчёт", "owner": "Анна", "deadline": "2026-09-25", "evidence": "Анна подготовит отчёт к 2026-09-25.", "segment_id": "s1"}}}
}
func TestNemotronGroundedPendingActions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Error(r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["model"] != "exact-model" || object(body["response_format"])["type"] != "json_object" {
			t.Error(body)
		}
		payload, _ := json.Marshal(protocolPayload())
		writeJSON(w, 200, map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"content": string(payload)}}}})
	}))
	defer server.Close()
	t.Setenv("NEMOTRON_BASE_URL", server.URL+"/v1")
	t.Setenv("NEMOTRON_MODEL", "exact-model")
	summary, actions, err := extract(context.Background(), sourceTranscript(), "2026-09-23")
	if err != nil {
		t.Fatal(err)
	}
	action := object(actions[0])
	if summary == "" || action["owner"] != "Анна" || action["deadline"] != "2026-09-25" || action["status"] != "pending" {
		t.Fatal(summary, actions)
	}
}
func TestRejectHallucinationsAndUnknownFields(t *testing.T) {
	payload := protocolPayload()
	action := object(entries(payload["actions"])[0])
	action["owner"] = "Борис"
	action["deadline"] = "2026-10-01"
	_, actions, err := validateExtraction(payload, sourceTranscript())
	if err != nil || object(actions[0])["owner"] != nil || object(actions[0])["deadline"] != nil {
		t.Fatal(actions, err)
	}
	action["evidence"] = "Борис всё сделает."
	if _, _, err = validateExtraction(payload, sourceTranscript()); err == nil {
		t.Fatal("hallucinated evidence accepted")
	}
	payload = protocolPayload()
	payload["extra"] = "untrusted"
	if _, _, err = validateExtraction(payload, sourceTranscript()); err == nil {
		t.Fatal("unknown output field accepted")
	}
}
func TestSelfHostedASRPreservesUnknownSpeaker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/transcriptions" {
			t.Error(r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Error(err)
		}
		if r.FormValue("model") != "exact-asr" || r.FormValue("language") != "kk" || r.FormValue("response_format") != "verbose_json" {
			t.Error(r.Form)
		}
		writeJSON(w, 200, map[string]any{"segments": []any{map[string]any{"text": "Сәлем, команда.", "start": 0, "end": 2.5}}})
	}))
	defer server.Close()
	t.Setenv("ASR_BASE_URL", server.URL+"/v1")
	t.Setenv("ASR_MODEL", "exact-asr")
	path := filepath.Join(t.TempDir(), "audio.wav")
	_ = os.WriteFile(path, []byte("RIFFaudio"), 0600)
	segments, err := transcribe(context.Background(), path, "kk")
	if err != nil {
		t.Fatal(err)
	}
	if object(segments[0])["speaker"] != "Спикер не определён" || object(segments[0])["end"] != 2.5 {
		t.Fatal(segments)
	}
}
func TestBlockedCloudHostsAndMissingModel(t *testing.T) {
	for _, base := range []string{"https://api.openai.com/v1", "https://integrate.api.nvidia.com/v1", "https://ai.api.nvidia.com/v1", "https://API.OPENAI.COM./v1", "https://user:secret@private.test/v1"} {
		if _, err := providerEndpoint(base, "/chat/completions"); err == nil {
			t.Fatal("accepted", base)
		}
	}
	t.Setenv("NEMOTRON_BASE_URL", "http://unreachable.invalid/v1")
	t.Setenv("NEMOTRON_MODEL", "")
	if _, _, err := extract(context.Background(), sourceTranscript(), "2026-09-23"); err == nil || !strings.Contains(err.Error(), "NEMOTRON_MODEL") {
		t.Fatal(err)
	}
}
func TestProviderErrorsDoNotLeakBodyOrToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte("secret-token confidential transcript"))
	}))
	defer server.Close()
	_, err := providerRequest(context.Background(), server.URL, "/chat/completions", "secret-token", "Nemotron", "application/json", bytes.NewBufferString("{}"))
	if err == nil || strings.Contains(err.Error(), "secret-token") || strings.Contains(err.Error(), "confidential") {
		t.Fatal(err)
	}
}
func TestInvalidTimestamps(t *testing.T) {
	for _, value := range []any{math.NaN(), math.Inf(1), -1.0, "zero", true, 86401.0} {
		if _, err := numeric(value); err == nil {
			t.Fatal("invalid timestamp accepted", value)
		}
	}
}
