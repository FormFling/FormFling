package main

import (
	"log"
	"net/http"

	"formfling/internal/config"
	"formfling/internal/handlers"
	"formfling/internal/middleware"
	"formfling/internal/services"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate required environment variables
	if cfg.SMTPUsername == "" || cfg.SMTPPassword == "" {
		log.Fatal("SMTP_USERNAME and SMTP_PASSWORD are required")
	}
	if cfg.FromEmail == "" || cfg.ToEmail == "" {
		log.Fatal("FROM_EMAIL and TO_EMAIL are required")
	}

	// Initialize services
	emailService := services.NewEmailService(cfg)
	recaptchaService := services.NewRecaptchaService(cfg)

	// Log reCAPTCHA status
	if cfg.RecaptchaEnabled {
		log.Printf("reCAPTCHA v3 enabled (min score: %.2f, action: %s)",
			cfg.RecaptchaMinScore, cfg.RecaptchaAction)
	} else {
		log.Printf("reCAPTCHA v3 disabled (no secret key provided)")
	}

	// Setup handlers
	submitHandler := handlers.NewSubmitHandler(cfg, emailService, recaptchaService)
	healthHandler := handlers.NewHealthHandler()
	statusHandler := handlers.NewStatusHandler(cfg)

	// Setup router
	r := mux.NewRouter()
	r.Use(middleware.CORS(cfg))

	r.HandleFunc("/submit", submitHandler.Handle).Methods("POST", "OPTIONS")
	r.HandleFunc("/health", healthHandler.Handle).Methods("GET")
	r.HandleFunc("/status", statusHandler.Handle).Methods("GET")

	if cfg.EnableTestForm {
		testFormHandler := handlers.NewTestFormHandler(cfg)
		r.HandleFunc("/test_form", testFormHandler.Handle).Methods("GET")
	}

	// Static file serving
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))

	log.Printf("FormFling server starting on port %s", cfg.Port)
	log.Printf("Allowed origins: %v", cfg.AllowedOrigins)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
