package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"formfling/internal/config"
	"formfling/internal/handlers"
	"formfling/internal/middleware"
	"formfling/internal/services"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

// Database initialization function
func initDB() (*sql.DB, error) {
	// Ensure the data directory exists
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(dataDir, "formfling.db"))
	if err != nil {
		return nil, err
	}

	// Create the users table if it doesn't exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL
        )
    `)
	if err != nil {
		return nil, err
	}

	return db, nil
}

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

	db, err := initDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}
	defer db.Close()

	// Initialize email service
	emailService := services.NewEmailService(cfg)

	// Setup handlers
	submitHandler := handlers.NewSubmitHandler(cfg, emailService)
	healthHandler := handlers.NewHealthHandler()
	statusHandler := handlers.NewStatusHandler(cfg)
	adminHandler := handlers.NewAdminHandler(cfg, db)
	loginHandler := handlers.NewLoginHandler(cfg, db)

	// Setup router
	r := mux.NewRouter()
	r.Use(middleware.CORS(cfg))

	r.HandleFunc("/submit", submitHandler.Handle).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/f/{formId}", submitHandler.Handle).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/health", healthHandler.Handle).Methods(http.MethodGet)
	r.HandleFunc("/status", statusHandler.Handle).Methods(http.MethodGet)
	r.HandleFunc("/admin", adminHandler.Handle).Methods(http.MethodGet)
	r.HandleFunc("/login", loginHandler.HandleLogin).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/register", loginHandler.HandleRegister).Methods(http.MethodGet, http.MethodPost)

	// Static file serving
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	log.Printf("FormFling server starting on port %s", cfg.Port)
	log.Printf("Allowed origins: %v", cfg.AllowedOrigins)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
