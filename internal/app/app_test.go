package app

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func testServer(t *testing.T) (*App, *http.ServeMux) {
	t.Helper()
	a, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	a.Register(mux)
	return a, mux
}
func request(t *testing.T, mux http.Handler, method, path string, body io.Reader, cookie *http.Cookie, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, body)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}
func session(t *testing.T, mux http.Handler) *http.Cookie {
	w := request(t, mux, "POST", "/api/session", nil, nil, "")
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	return w.Result().Cookies()[0]
}
func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	out := map[string]any{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err, w.Body.String())
	}
	return out
}
func upload(t *testing.T, mux http.Handler, cookie *http.Cookie, name string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("title", "Реальная встреча")
	part, _ := writer.CreateFormFile("file", name)
	_, _ = part.Write(content)
	_ = writer.Close()
	return request(t, mux, "POST", "/api/meetings", &body, cookie, writer.FormDataContentType())
}
func TestUploadWithoutASRAndOwnerIsolation(t *testing.T) {
	t.Setenv("ASR_BASE_URL", "")
	_, mux := testServer(t)
	alice, bob := session(t, mux), session(t, mux)
	w := upload(t, mux, alice, "../../meeting.wav", []byte("RIFFaudio"))
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	meeting := decode(t, w)
	if meeting["status"] != "awaiting_provider" || meeting["mode"] != "live" || len(entries(meeting["transcript"])) != 0 || len(entries(meeting["actions"])) != 0 {
		t.Fatal(meeting)
	}
	id := text(meeting["id"])
	for _, suffix := range []string{"", "/audio", "/export?format=json"} {
		if got := request(t, mux, "GET", "/api/meetings/"+id+suffix, nil, bob, ""); got.Code != 404 {
			t.Fatal(got.Code, got.Body.String())
		}
	}
	if _, ok := meeting["_audio_file"]; ok {
		t.Fatal("private storage path exposed")
	}
}
func TestDemoReviewAndConfirm(t *testing.T) {
	_, mux := testServer(t)
	cookie := session(t, mux)
	w := request(t, mux, "POST", "/api/demo", nil, cookie, "")
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	meeting := decode(t, w)
	id := text(meeting["id"])
	if meeting["mode"] != "demo" || len(entries(meeting["actions"])) != 3 {
		t.Fatal(meeting)
	}
	if request(t, mux, "POST", "/api/meetings/"+id+"/confirm", nil, cookie, "").Code != 409 {
		t.Fatal("unreviewed confirmation accepted")
	}
	for _, item := range entries(meeting["actions"]) {
		object(item)["status"] = "confirmed"
	}
	patch, _ := json.Marshal(map[string]any{"summary": "Проверено", "actions": meeting["actions"]})
	if got := request(t, mux, "PATCH", "/api/meetings/"+id, bytes.NewReader(patch), cookie, "application/json"); got.Code != 200 {
		t.Fatal(got.Code, got.Body.String())
	}
	first := request(t, mux, "POST", "/api/meetings/"+id+"/confirm", nil, cookie, "")
	second := request(t, mux, "POST", "/api/meetings/"+id+"/confirm", nil, cookie, "")
	if first.Code != 200 || second.Code != 200 || first.Body.String() != second.Body.String() {
		t.Fatal("confirmation must be idempotent", first.Body.String(), second.Body.String())
	}
}
func TestEmptyAndUnsupportedInput(t *testing.T) {
	_, mux := testServer(t)
	cookie := session(t, mux)
	if got := request(t, mux, "POST", "/api/meetings", strings.NewReader("title=Empty"), cookie, "application/x-www-form-urlencoded"); got.Code != 422 {
		t.Fatal(got.Code)
	}
	if got := upload(t, mux, cookie, "attack.html", []byte("hello")); got.Code != 415 {
		t.Fatal(got.Code)
	}
}

func TestManualTranscriptPersistsWithoutFabricatedInference(t *testing.T) {
	t.Setenv("NEMOTRON_BASE_URL", "")
	a, mux := testServer(t)
	cookie := session(t, mux)
	values := url.Values{"title": {"Текст"}, "language": {"mixed"}, "transcript": {"Анна: подготовлю план.\nДанияр: келісемін."}}
	w := request(t, mux, "POST", "/api/meetings", strings.NewReader(values.Encode()), cookie, "application/x-www-form-urlencoded")
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	meeting := decode(t, w)
	if meeting["mode"] != "manual" || meeting["status"] != "needs_review" || len(entries(meeting["actions"])) != 0 || len(entries(meeting["transcript"])) != 2 {
		t.Fatal(meeting)
	}
	restarted, err := New(a.dataDir)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(cookie)
	owner, err := restarted.Owner(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = restarted.Get(owner, text(meeting["id"])); err != nil {
		t.Fatal(err)
	}
}
func TestUnknownOwnerDeadlineRequireReviewAndEditsReopen(t *testing.T) {
	_, mux := testServer(t)
	cookie := session(t, mux)
	meeting := decode(t, request(t, mux, "POST", "/api/demo", nil, cookie, ""))
	id := text(meeting["id"])
	actions := entries(meeting["actions"])
	if object(actions[2])["owner"] != nil || object(actions[2])["deadline"] != nil {
		t.Fatal("demo uncertainty hidden")
	}
	for _, item := range actions {
		object(item)["status"] = "confirmed"
	}
	body, _ := json.Marshal(map[string]any{"actions": actions})
	request(t, mux, "PATCH", "/api/meetings/"+id, bytes.NewReader(body), cookie, "application/json")
	request(t, mux, "POST", "/api/meetings/"+id+"/confirm", nil, cookie, "")
	edited := decode(t, request(t, mux, "PATCH", "/api/meetings/"+id, strings.NewReader(`{"summary":"Исправлено"}`), cookie, "application/json"))
	if edited["status"] != "needs_review" {
		t.Fatal(edited)
	}
}
func TestDeleteAudioAndInvalidUploadCleanup(t *testing.T) {
	t.Setenv("ASR_BASE_URL", "")
	a, mux := testServer(t)
	cookie := session(t, mux)
	if w := upload(t, mux, cookie, "empty.wav", nil); w.Code != 422 {
		t.Fatal(w.Code)
	}
	files, _ := os.ReadDir(filepath.Join(a.dataDir, "uploads"))
	if len(files) != 0 {
		t.Fatal("failed upload left file")
	}
	meeting := decode(t, upload(t, mux, cookie, "audio.wav", []byte("audio")))
	if w := request(t, mux, "DELETE", "/api/meetings/"+text(meeting["id"]), nil, cookie, ""); w.Code != 204 {
		t.Fatal(w.Code)
	}
	files, _ = os.ReadDir(filepath.Join(a.dataDir, "uploads"))
	if len(files) != 0 {
		t.Fatal("deleted audio remains")
	}
}
func TestCookieAttributesAndRecovery(t *testing.T) {
	t.Setenv("BUTTERFLY_COOKIE_SAMESITE", "none")
	a, mux := testServer(t)
	cookie := session(t, mux)
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteNoneMode {
		t.Fatal(cookie)
	}
	meeting := newMeeting("Interrupted", "ru", "2026-09-23", "live")
	if err := a.create("owner", meeting); err != nil {
		t.Fatal(err)
	}
	restarted, err := New(a.dataDir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := restarted.Get("owner", text(meeting["id"]))
	if err != nil || got["status"] != "failed" {
		t.Fatal(got, err)
	}
}
func TestPrivateJSONExportAndCallbacks(t *testing.T) {
	a, mux := testServer(t)
	created := 0
	reviewed := 0
	a.OnCreated = func(map[string]any) { created++ }
	a.OnReview = func(map[string]any, map[string]any) { reviewed++ }
	cookie := session(t, mux)
	meeting := decode(t, request(t, mux, "POST", "/api/demo", nil, cookie, ""))
	id := text(meeting["id"])
	if created != 1 {
		t.Fatal(created)
	}
	for _, item := range entries(meeting["actions"]) {
		object(item)["status"] = "confirmed"
	}
	patch, _ := json.Marshal(map[string]any{"actions": meeting["actions"]})
	request(t, mux, "PATCH", "/api/meetings/"+id, bytes.NewReader(patch), cookie, "application/json")
	request(t, mux, "POST", "/api/meetings/"+id+"/confirm", nil, cookie, "")
	request(t, mux, "POST", "/api/meetings/"+id+"/confirm", nil, cookie, "")
	if reviewed != 1 {
		t.Fatal("duplicate review callback", reviewed)
	}
	w := request(t, mux, "GET", "/api/meetings/"+id+"/export?format=json", nil, cookie, "")
	if w.Code != 200 || decode(t, w)["status"] != "completed" {
		t.Fatal(w.Code, w.Body.String())
	}
	a.Export = func(_ map[string]any, format string) ([]byte, error) { return []byte("export " + format), nil }
	w = request(t, mux, "GET", "/api/meetings/"+id+"/export?format=pdf", nil, cookie, "")
	if w.Code != 200 || w.Body.String() != "export pdf" {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestMutationsAndAudioStayPrivate(t *testing.T) {
	t.Setenv("ASR_BASE_URL", "")
	_, mux := testServer(t)
	alice, bob := session(t, mux), session(t, mux)
	meeting := decode(t, upload(t, mux, alice, "private.mp3", []byte("audio")))
	id := text(meeting["id"])
	for _, tc := range []struct{ method, path, body string }{{"PATCH", "/api/meetings/" + id, `{"summary":"stolen"}`}, {"POST", "/api/meetings/" + id + "/retry", ""}, {"POST", "/api/meetings/" + id + "/confirm", ""}, {"DELETE", "/api/meetings/" + id, ""}} {
		if got := request(t, mux, tc.method, tc.path, strings.NewReader(tc.body), bob, "application/json"); got.Code != 404 {
			t.Fatal(tc.method, got.Code)
		}
	}
	if got := request(t, mux, "GET", "/api/meetings", nil, nil, ""); got.Code != 401 {
		t.Fatal(got.Code)
	}
	if meetings := entries(decode(t, request(t, mux, "GET", "/api/meetings", nil, bob, ""))["meetings"]); len(meetings) != 0 {
		t.Fatal(meetings)
	}
}
func TestConcurrentConfirmationEmitsOneCallback(t *testing.T) {
	a, mux := testServer(t)
	var calls atomic.Int32
	a.OnReview = func(map[string]any, map[string]any) { calls.Add(1) }
	cookie := session(t, mux)
	meeting := decode(t, request(t, mux, "POST", "/api/demo", nil, cookie, ""))
	id := text(meeting["id"])
	for _, item := range entries(meeting["actions"]) {
		object(item)["status"] = "confirmed"
	}
	patch, _ := json.Marshal(map[string]any{"actions": meeting["actions"]})
	request(t, mux, "PATCH", "/api/meetings/"+id, bytes.NewReader(patch), cookie, "application/json")
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			r := httptest.NewRequest("POST", "/api/meetings/"+id+"/confirm", nil)
			r.AddCookie(cookie)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != 200 {
				t.Errorf("confirm status %d", w.Code)
			}
		}()
	}
	group.Wait()
	if calls.Load() != 1 {
		t.Fatal("duplicate Fly callback", calls.Load())
	}
}

func TestDocumentExportRequiresConfirmedProtocol(t *testing.T) {
	a, mux := testServer(t)
	a.Export = func(map[string]any, string) ([]byte, error) {
		t.Fatal("unconfirmed export callback invoked")
		return nil, nil
	}
	cookie := session(t, mux)
	meeting := decode(t, request(t, mux, "POST", "/api/demo", nil, cookie, ""))
	for _, format := range []string{"pdf", "docx"} {
		w := request(t, mux, "GET", "/api/meetings/"+text(meeting["id"])+"/export?format="+format, nil, cookie, "")
		if w.Code != 409 {
			t.Fatal(format, w.Code)
		}
	}
}
