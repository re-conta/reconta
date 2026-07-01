package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/lucasbrum/reconta/api/internal/auth"
	"github.com/lucasbrum/reconta/api/internal/db"
	"github.com/lucasbrum/reconta/api/internal/user"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "3020")
	dbPath := getEnv("DB_PATH", "./data/reconta.db")
	secureCookies := getEnv("ENV", "development") == "production"

	conn, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("erro ao abrir banco de dados: %v", err)
	}
	defer conn.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	userRepo := user.NewRepository(conn)
	userHandler := user.NewHandler(userRepo)
	userHandler.RegisterRoutes(mux)

	authHandler := auth.NewHandler(auth.NewRepository(conn), userRepo, secureCookies)
	authHandler.RegisterRoutes(mux)

	addr := ":" + port
	log.Printf("servidor rodando em %s (db: %s)", addr, dbPath)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}

// withCORS habilita CORS para o servidor de desenvolvimento do Vite.
// Em produção o Nginx faz proxy same-origin em /api/, então isso é inofensivo.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv carrega variáveis de um arquivo .env simples (KEY=VALUE por linha),
// sem sobrescrever variáveis já definidas no ambiente. Usado apenas em
// desenvolvimento local — em produção o systemd injeta o EnvironmentFile.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		os.Setenv(key, strings.TrimSpace(value))
	}
}
