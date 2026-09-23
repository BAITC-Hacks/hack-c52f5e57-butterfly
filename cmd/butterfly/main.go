package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"butterfly/internal/app"
	"butterfly/internal/integrations"
)

func main() {
	loadEnv(".env")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dir := os.Getenv("BUTTERFLY_DATA_DIR")
	if dir == "" {
		dir = "data/runtime"
	}
	application, err := app.New(dir)
	if err != nil {
		log.Fatal("initialize application: ", err)
	}
	application.Export = integrations.Export
	services := integrations.Start(ctx, dir)
	defer services.Close()
	application.OnCreated = services.Created
	application.OnReview = services.Reviewed
	mux := http.NewServeMux()
	application.Register(mux)
	services.Register(mux, application)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "version": "0.2.0", "runtime": "go", "provider": map[string]any{"llm_configured": os.Getenv("NEMOTRON_BASE_URL") != "" || os.Getenv("LOCAL_LLM_BASE_URL") != "", "asr_configured": os.Getenv("ASR_BASE_URL") != "" || os.Getenv("RIVA_ASR_ENDPOINT") != "", "live_verified": false}, "mcp": services.Health(), "telegram_configured": services.TelegramConfigured()})
	})
	mux.Handle("/", http.FileServer(http.Dir("frontend")))
	host := os.Getenv("BUTTERFLY_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	server := &http.Server{Addr: host + ":" + port, Handler: protect(mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 180 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	log.Printf("Butterfly Go listening on http://%s (data: %s)", server.Addr, filepath.Clean(dir))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// Dotenv values are read literally, never executed as shell commands.
func loadEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || strings.ContainsAny(key, " \t$") {
			continue
		}
		if _, set := os.LookupEnv(key); set {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		_ = os.Setenv(key, value)
	}
}

func protect(next http.Handler) http.Handler {
	origins := map[string]bool{}
	for _, origin := range strings.Split(os.Getenv("BUTTERFLY_CORS_ORIGINS"), ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" && origin != "*" {
			origins[origin] = true
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		origin := r.Header.Get("Origin")
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		allowed := origin == "" || origin == scheme+"://"+r.Host || origins[origin]
		if origin != "" && origins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			if !allowed {
				http.Error(w, "origin denied", 403)
				return
			}
			w.WriteHeader(204)
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" && !allowed {
			http.Error(w, "origin denied", 403)
			return
		}
		const maxBody = 52 << 20
		if r.ContentLength > maxBody {
			http.Error(w, "request too large", 413)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		next.ServeHTTP(w, r)
	})
}
