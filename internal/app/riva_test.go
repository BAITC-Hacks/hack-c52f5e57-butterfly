package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBrevHandoffAliasesAndRequest(t *testing.T) {
	t.Setenv("NEMOTRON_BASE_URL", "")
	t.Setenv("NEMOTRON_MODEL", "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["model"] != "nvidia/nemotron-3.5-lightning-30b-a3b" || object(body["chat_template_kwargs"])["enable_thinking"] != false || object(body["response_format"])["type"] != "json_object" {
			t.Error("request differs from tested Brev contract", body)
		}
		payload, _ := json.Marshal(protocolPayload())
		writeJSON(w, 200, map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"content": string(payload)}}}})
	}))
	defer server.Close()
	t.Setenv("LOCAL_LLM_BASE_URL", server.URL+"/v1")
	t.Setenv("LOCAL_LLM_MODEL", "nvidia/nemotron-3.5-lightning-30b-a3b")
	_, actions, err := extract(context.Background(), sourceTranscript(), "2026-09-23")
	if err != nil || len(actions) != 1 {
		t.Fatalf("ready Brev configuration rejected: %v actions=%v", err, actions)
	}
}

func TestRivaPreservesTimesAndNeverInventsSpeakers(t *testing.T) {
	segments, err := parseRivaOutput([]byte(`{"segments":[{"text":"Анна подготовит отчёт.","start":61.25,"end":63.8}],"diarization":false}`))
	if err != nil || len(segments) != 1 {
		t.Fatal(segments, err)
	}
	segment := object(segments[0])
	if segment["start"] != 61.25 || segment["end"] != 63.8 || segment["speaker"] != "Спикер не определён" {
		t.Fatal(segment)
	}
	for _, raw := range []string{`{`, `{"segments":[]}`, `{"segments":[{"text":"x","start":4,"end":2}]}`, `{"segments":[{"text":"x","start":0,"end":1}],"diarization":true}`, `{"segments":[{"text":"x","start":-1,"end":2}]}`} {
		if _, err := parseRivaOutput([]byte(raw)); err == nil {
			t.Fatalf("accepted invalid helper output %s", raw)
		}
	}
}

func TestNemotronRejectsTruncatedJSONEvenWhenParseable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := json.Marshal(protocolPayload())
		writeJSON(w, 200, map[string]any{"choices": []any{map[string]any{"finish_reason": "length", "message": map[string]any{"content": string(payload)}}}})
	}))
	defer server.Close()
	t.Setenv("NEMOTRON_BASE_URL", server.URL+"/v1")
	t.Setenv("NEMOTRON_MODEL", "exact-model")
	if _, _, err := extract(context.Background(), sourceTranscript(), ""); err == nil {
		t.Fatal("accepted truncated model output")
	}
}

func TestRivaDoesNotAllowEndpointURLs(t *testing.T) {
	t.Setenv("RIVA_ASR_ENDPOINT", "https://user:password@host/")
	if _, err := transcribe(context.Background(), "unread-file.mp3", "ru"); err == nil {
		t.Fatal("accepted malformed gRPC target")
	}
}
