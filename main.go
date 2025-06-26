package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type FormData struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
}

type EmailTemplateData struct {
	FormData      FormData
	SubmittedTime string
	SubmittedDate string
	Origin        string
}

type Config struct {
	Port            string
	SMTPHost        string
	SMTPPort        int
	SMTPUsername    string
	SMTPPassword    string
	FromEmail       string
	FromName        string
	ToEmail         string
	ToName          string
	AllowedOrigins  []string
	FormTitle       string
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

var config Config
var emailTemplate *template.Template

func init() {
	config = Config{
		Port:         getEnv("PORT", "8080"),
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", ""),
		FromName:     getEnv("FROM_NAME", "FormFling Bot"),
		ToEmail:      getEnv("TO_EMAIL", ""),
		ToName:       getEnv("TO_NAME", ""),
		FormTitle:    getEnv("FORM_TITLE", "Contact Me"),
	}

	// Parse allowed origins
	allowedOriginsStr := getEnv("ALLOWED_ORIGINS", "*")
	if allowedOriginsStr != "*" && allowedOriginsStr != "" {
		config.AllowedOrigins = strings.Split(allowedOriginsStr, ",")
		for i := range config.AllowedOrigins {
			config.AllowedOrigins[i] = strings.TrimSpace(config.AllowedOrigins[i])
		}
	}

	// Load email template
	var err error
	emailTemplate, err = template.ParseFiles("email_template.html")
	if err != nil {
		log.Fatal("Error loading email template:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed (skip check if ALLOWED_ORIGINS is "*" or empty)
		if len(config.AllowedOrigins) > 0 {
			allowed := false
			for _, allowedOrigin := range config.AllowedOrigins {
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
		if len(config.AllowedOrigins) == 0 {
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

func cleanString(input string) string {
	// Remove potentially dangerous content
	bad := []string{
		"content-type",
		"bcc:",
		"to:",
		"cc:",
		"href",
	}
	
	result := input
	for _, badWord := range bad {
		result = strings.ReplaceAll(strings.ToLower(result), badWord, "")
	}
	
	return strings.TrimSpace(result)
}

func validateEmail(email string) bool {
	// RFC 5322 compliant email regex (simplified)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func validateForm(form FormData) error {
	if strings.TrimSpace(form.Name) == "" {
		return fmt.Errorf("name is required")
	}
	
	if strings.TrimSpace(form.Email) == "" {
		return fmt.Errorf("email is required")
	}
	
	if !validateEmail(form.Email) {
		return fmt.Errorf("email not valid")
	}
	
	if len(strings.TrimSpace(form.Message)) < 10 {
		return fmt.Errorf("message not valid")
	}
	
	return nil
}

func sendEmail(formData FormData, origin string) error {
	now := time.Now()
	templateData := EmailTemplateData{
		FormData:      formData,
		SubmittedTime: now.Format("03:04 PM"),
		SubmittedDate: now.Format("02 January 2006"),
		Origin:        origin,
	}

	var emailBody bytes.Buffer
	if err := emailTemplate.Execute(&emailBody, templateData); err != nil {
		return fmt.Errorf("error executing email template: %v", err)
	}

	// Set up authentication information
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)

	// Create message
	subject := fmt.Sprintf("❗ %s Form ❗", config.FormTitle)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	
	msg := fmt.Sprintf("From: %s <%s>\r\n", config.FromName, config.FromEmail)
	msg += fmt.Sprintf("To: %s <%s>\r\n", config.ToName, config.ToEmail)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += mime
	msg += emailBody.String()

	// Connect to server and send email
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
	
	// Handle different SMTP configurations
	if config.SMTPPort == 465 {
		// SSL/TLS connection for port 465
		conn, err := tls.Dial("tcp", addr, &tls.Config{
			ServerName: config.SMTPHost,
		})
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %v", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, config.SMTPHost)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %v", err)
		}
		defer client.Quit()

		// Authenticate
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %v", err)
		}

		// Send email
		return sendEmailData(client, msg)
	} else {
		// STARTTLS connection for port 587 (and others)
		conn, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %v", err)
		}
		defer conn.Quit()

		// Start TLS if available
		if ok, _ := conn.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName: config.SMTPHost,
			}
			if err := conn.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("failed to start TLS: %v", err)
			}
		}

		// Authenticate
		if err := conn.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %v", err)
		}

		// Send email
		return sendEmailData(conn, msg)
	}
}

func sendEmailData(client *smtp.Client, msg string) error {
	// Set sender and recipient
	if err := client.Mail(config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	if err := client.Rcpt(config.ToEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send email body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %v", err)
	}

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email data: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close email writer: %v", err)
	}

	return nil
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		response := Response{Status: "error", Error: "must be a post"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		response := Response{Status: "error", Error: "failed to parse form"}
		json.NewEncoder(w).Encode(response)
		return
	}

	formData := FormData{
		Name:    cleanString(r.FormValue("name")),
		Email:   cleanString(r.FormValue("email")),
		Subject: cleanString(r.FormValue("subject")),
		Message: cleanString(r.FormValue("message")),
		Phone:   cleanString(r.FormValue("phone")),
		Website: cleanString(r.FormValue("website")),
	}

	// Validate form
	if err := validateForm(formData); err != nil {
		response := Response{Status: "error", Error: "server rejected"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get origin for email
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = r.Header.Get("Referer")
	}

	// Send email
	if err := sendEmail(formData, origin); err != nil {
		log.Printf("Error sending email: %v", err)
		response := Response{Status: "error", Error: "failed to send email"}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := Response{Status: "message sent"}
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := Response{Status: "ok"}
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Validate required environment variables
	if config.SMTPUsername == "" || config.SMTPPassword == "" {
		log.Fatal("SMTP_USERNAME and SMTP_PASSWORD are required")
	}
	if config.FromEmail == "" || config.ToEmail == "" {
		log.Fatal("FROM_EMAIL and TO_EMAIL are required")
	}

	r := mux.NewRouter()
	r.Use(corsMiddleware)
	
	r.HandleFunc("/submit", handleSubmit).Methods("POST", "OPTIONS")
	r.HandleFunc("/health", handleHealth).Methods("GET")

	log.Printf("FormFling server starting on port %s", config.Port)
	log.Printf("Allowed origins: %v", config.AllowedOrigins)
	
	if err := http.ListenAndServe(":"+config.Port, r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}