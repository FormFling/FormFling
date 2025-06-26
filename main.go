package main

import (
	"log"
	"net/http"

	"formfling/config"
	"formfling/handlers"
	"formfling/middleware"
	"formfling/services"

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

	// Initialize email service
	emailService := services.NewEmailService(cfg)

	// Setup handlers
	submitHandler := handlers.NewSubmitHandler(cfg, emailService)
	healthHandler := handlers.NewHealthHandler()
	statusHandler := handlers.NewStatusHandler(cfg)

	// Setup router
	r := mux.NewRouter()
	r.Use(middleware.CORS(cfg))

	r.HandleFunc("/submit", submitHandler.Handle).Methods("POST", "OPTIONS")
	r.HandleFunc("/health", healthHandler.Handle).Methods("GET")
	r.HandleFunc("/status", statusHandler.Handle).Methods("GET")

	// Static file serving for images
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("./images/"))))

	log.Printf("FormFling server starting on port %s", cfg.Port)
	log.Printf("Allowed origins: %v", cfg.AllowedOrigins)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
