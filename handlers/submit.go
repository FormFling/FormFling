package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"formfling/config"
	"formfling/models"
	"formfling/services"
	"formfling/utils"
)

type SubmitHandler struct {
	config       *config.Config
	emailService services.EmailSender
}

func NewSubmitHandler(cfg *config.Config, emailService services.EmailSender) *SubmitHandler {
	return &SubmitHandler{
		config:       cfg,
		emailService: emailService,
	}
}

func (h *SubmitHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.handleError(w, r, "must be a post", http.StatusMethodNotAllowed)
		return
	}

	var formData models.FormData
	var err error

	// Parse data based on content type
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// Parse JSON request body
		formData, err = h.parseJSONRequest(r)
		if err != nil {
			h.handleError(w, r, "failed to parse JSON", http.StatusBadRequest)
			return
		}
	} else {
		// Parse form data (default)
		if err := r.ParseForm(); err != nil {
			h.handleError(w, r, "failed to parse form", http.StatusBadRequest)
			return
		}
		formData = models.FormData{
			Name:    utils.CleanString(r.FormValue("name")),
			Email:   utils.CleanString(r.FormValue("email")),
			Subject: utils.CleanString(r.FormValue("subject")),
			Message: utils.CleanString(r.FormValue("message")),
			Phone:   utils.CleanString(r.FormValue("phone")),
			Website: utils.CleanString(r.FormValue("website")),
		}
	}

	// Validate form
	if err := utils.ValidateForm(formData); err != nil {
		h.handleError(w, r, "server rejected", http.StatusBadRequest)
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
		h.handleError(w, r, "failed to send email", http.StatusInternalServerError)
		return
	}

	h.handleSuccess(w, r)
}

func (h *SubmitHandler) parseJSONRequest(r *http.Request) (models.FormData, error) {
	var formData models.FormData

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return formData, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	// Parse JSON
	if err := json.Unmarshal(body, &formData); err != nil {
		return formData, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Clean the data
	formData.Name = utils.CleanString(formData.Name)
	formData.Email = utils.CleanString(formData.Email)
	formData.Subject = utils.CleanString(formData.Subject)
	formData.Message = utils.CleanString(formData.Message)
	formData.Phone = utils.CleanString(formData.Phone)
	formData.Website = utils.CleanString(formData.Website)

	return formData, nil
}

func (h *SubmitHandler) handleSuccess(w http.ResponseWriter, r *http.Request) {
	// Check if this is an AJAX request (API mode)
	if h.isAjaxRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		response := models.Response{Status: "message sent"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Handle form submission with redirect
	redirectURL := h.getRedirectURL(r, "success")
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *SubmitHandler) handleError(w http.ResponseWriter, r *http.Request, errorMsg string, statusCode int) {
	// Check if this is an AJAX request (API mode)
	if h.isAjaxRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		response := models.Response{Status: "error", Error: errorMsg}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Handle form submission with redirect to error page
	redirectURL := h.getRedirectURL(r, "error")
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *SubmitHandler) isAjaxRequest(r *http.Request) bool {
	// Check for common AJAX indicators
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		r.Header.Get("Content-Type") == "application/json" ||
		strings.Contains(r.Header.Get("Accept"), "application/json")
}

func (h *SubmitHandler) getRedirectURL(r *http.Request, status string) string {
	// Only use explicit redirect URL from form data
	if redirectURL := r.FormValue("_redirect"); redirectURL != "" {
		// Add status parameter to the explicit redirect
		return h.addStatusParam(redirectURL, status)
	}

	// Always redirect to status page, pass referer as redirect parameter for the "Go Back" functionality
	referer := r.Header.Get("Referer")
	if referer != "" {
		// Pass the referer as redirect parameter so status page can redirect back
		return fmt.Sprintf("/status?type=%s&redirect=%s", status, url.QueryEscape(referer))
	}

	// If no referer, just redirect to status page
	return fmt.Sprintf("/status?type=%s", status)
}

func (h *SubmitHandler) addStatusParam(baseURL, status string) string {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Sprintf("/status?type=%s", status)
	}

	// Add or update the status parameter
	query := parsedURL.Query()
	query.Set("formfling_status", status)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String()
}
