package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"formfling/config"
	"formfling/models"
	"formfling/services"
	"formfling/utils"
)

type SubmitHandler struct {
	config       *config.Config
	emailService *services.EmailService
}

func NewSubmitHandler(cfg *config.Config, emailService *services.EmailService) *SubmitHandler {
	return &SubmitHandler{
		config:       cfg,
		emailService: emailService,
	}
}

func (h *SubmitHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		response := models.Response{Status: "error", Error: "must be a post"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		response := models.Response{Status: "error", Error: "failed to parse form"}
		json.NewEncoder(w).Encode(response)
		return
	}

	formData := models.FormData{
		Name:    utils.CleanString(r.FormValue("name")),
		Email:   utils.CleanString(r.FormValue("email")),
		Subject: utils.CleanString(r.FormValue("subject")),
		Message: utils.CleanString(r.FormValue("message")),
		Phone:   utils.CleanString(r.FormValue("phone")),
		Website: utils.CleanString(r.FormValue("website")),
	}

	// Validate form
	if err := utils.ValidateForm(formData); err != nil {
		response := models.Response{Status: "error", Error: "server rejected"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get origin for email
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = r.Header.Get("Referer")
	}

	// Send email
	if err := h.emailService.SendEmail(formData, origin); err != nil {
		log.Printf("Error sending email: %v", err)
		response := models.Response{Status: "error", Error: "failed to send email"}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.Response{Status: "message sent"}
	json.NewEncoder(w).Encode(response)
}
