package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestProtectRejectsForeignMutationAndPermitsExplicitOrigin(t *testing.T) {
	t.Setenv("BUTTERFLY_CORS_ORIGINS", "https://butterfly.example")
	h := protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	for _, tc := range []struct {
		method, origin string
		want           int
	}{
		{"POST", "https://attacker.example", 403},
		{"OPTIONS", "https://attacker.example", 403},
		{"POST", "https://butterfly.example", 200},
		{"OPTIONS", "https://butterfly.example", 204},
		{"POST", "http://example.com", 200},
	} {
		r := httptest.NewRequest(tc.method, "/api/session", nil)
		r.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("%s %s: %d want %d", tc.method, tc.origin, w.Code, tc.want)
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("private API cache not disabled")
		}
		if tc.origin == "https://butterfly.example" && w.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Fatal("approved credentials absent")
		}
	}
}

func TestProtectRejectsOversizedBody(t *testing.T) {
	h := protect(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("oversized body reached handler") }))
	r := httptest.NewRequest("POST", "/api/meetings", nil)
	r.ContentLength = (52 << 20) + 1
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 413 {
		t.Fatal(w.Code)
	}
}

func TestDotenvReadsLiteralValuesWithoutOverridingEnvironment(t *testing.T) {
	t.Setenv("BUTTERFLY_TEST_ENV_EXISTING", "already-set")
	key := "BUTTERFLY_TEST_ENV_LITERAL"
	old, existed := os.LookupEnv(key)
	_ = os.Unsetenv(key)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("BUTTERFLY_TEST_ENV_EXISTING=override\nBUTTERFLY_TEST_ENV_LITERAL='$(touch /tmp/never-run-by-dotenv)'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	loadEnv(path)
	if os.Getenv("BUTTERFLY_TEST_ENV_EXISTING") != "already-set" || os.Getenv(key) != "$(touch /tmp/never-run-by-dotenv)" {
		t.Fatal("dotenv values changed")
	}
}
