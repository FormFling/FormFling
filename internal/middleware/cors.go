package middleware

import (
	"net/http"

	"formfling/internal/config"
)

func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed (skip check if ALLOWED_ORIGINS is "*" or empty)
			if len(cfg.AllowedOrigins) > 0 {
				allowed := false
				for _, allowedOrigin := range cfg.AllowedOrigins {
					if origin == allowedOrigin {
						allowed = true
						break
					}
				}
				if !allowed && origin != "" {
					http.Error(w, "Origin not allowed", http.StatusForbidden)
					return
				}
			}

			// Set CORS headers - allow all origins if no restrictions
			if len(cfg.AllowedOrigins) == 0 {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
