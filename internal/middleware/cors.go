package middleware

import (
	"net/http"
	"os"
	"strings"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setAllowedOrigin(w, r)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setAllowedOrigin(w http.ResponseWriter, r *http.Request) {
	allowedOrigins := strings.TrimSpace(getEnv("CORS_ALLOWED_ORIGINS", "*"))
	requestOrigin := r.Header.Get("Origin")

	if allowedOrigins == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		return
	}

	for _, origin := range strings.Split(allowedOrigins, ",") {
		if strings.TrimSpace(origin) == requestOrigin {
			w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
			w.Header().Set("Vary", "Origin")
			return
		}
	}
}

func JSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
