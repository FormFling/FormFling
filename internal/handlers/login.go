package handlers

import (
	"database/sql"
	"formfling/internal/config"
	"html/template"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type LoginHandler struct {
	config        *config.Config
	db            *sql.DB
	loginTemplate *template.Template
}

func NewLoginHandler(cfg *config.Config, db *sql.DB) *LoginHandler {
	loginTemplate, err := template.ParseFiles("./web/templates/login.html")
	if err != nil {
		log.Fatal("Error loading login template:", err)
	}

	return &LoginHandler{
		config:        cfg,
		db:            db,
		loginTemplate: loginTemplate,
	}
}

// Data structure for the login template
type LoginData struct {
	Method  string
	Message string
}

// Function to insert a user into the database
func insertUser(db *sql.DB, username, password string) error {
	// Hash the password before storing it in the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPassword)
	return err
}

// Function to verify login credentials
func verifyLogin(db *sql.DB, username, password string) bool {
	// Retrieve the hashed password from the database
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		return false
	}

	// Compare the provided password with the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// Function to check if a user exists in the database
func userExists(db *sql.DB) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users LIMIT 1)").Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func (h *LoginHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	// Redirect to login page if a user already exists
	if userExists(h.db) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Check if the request is a POST request
	if r.Method == http.MethodPost {
		exists := userExists(h.db)
		if exists {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Get the username and password from the form
		username := strings.TrimSpace(r.Form.Get("username"))
		password := r.Form.Get("password")

		// Insert the user into the database
		err = insertUser(h.db, username, password) // Default password for demonstration
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Execute the template with the error message
		err = h.loginTemplate.Execute(w, LoginData{Method: "/register", Message: "Invalid username or password"})
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Execute the template
	err := h.loginTemplate.Execute(w, LoginData{Method: "/register"})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *LoginHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Redirect to register page if a user doesn't exist yet
	if !userExists(h.db) {
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	// Check if the request is a POST request
	if r.Method == http.MethodPost {
		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Get the username and password from the form
		username := strings.TrimSpace(r.Form.Get("username"))
		password := r.Form.Get("password")

		// Verify login credentials
		if verifyLogin(h.db, username, password) {
			// Redirect to the greeting page with the username as a query parameter
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}

		// Execute the template with the error message
		err = h.loginTemplate.Execute(w, LoginData{Method: "/login", Message: "Invalid username or password"})
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Execute the template
	err := h.loginTemplate.Execute(w, LoginData{Method: "/login"})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
