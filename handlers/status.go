package handlers

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"formfling/config"
)

type StatusHandler struct {
	config         *config.Config
	statusTemplate *template.Template
}

func NewStatusHandler(cfg *config.Config) *StatusHandler {
	statusTemplate, err := template.ParseFiles(cfg.StatusTemplate)
	if err != nil {
		log.Fatal("Error loading status template:", err)
	}

	return &StatusHandler{
		config:         cfg,
		statusTemplate: statusTemplate,
	}
}

type StatusPageData struct {
	Status      string
	FormTitle   string
	Message     string
	RedirectURL string
}

func (h *StatusHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	status := r.URL.Query().Get("type")
	if status == "" {
		status = "success" // default
	}

	var message string
	if status == "success" {
		message = "Your message has been sent successfully!"
	} else {
		message = "There was an error sending your message. Please try again."
	}

	// Get redirect URL from referer or query parameter
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" {
		// Try to get referer, but clean it of status parameters
		if referer := r.Header.Get("Referer"); referer != "" {
			// Remove formfling_status parameter from referer
			if u, err := url.Parse(referer); err == nil {
				query := u.Query()
				query.Del("formfling_status")
				u.RawQuery = query.Encode()
				redirectURL = u.String()
			}
		}
	}

	data := StatusPageData{
		Status:      status,
		FormTitle:   h.config.FormTitle,
		Message:     message,
		RedirectURL: redirectURL,
	}

	if err := h.statusTemplate.Execute(w, data); err != nil {
		http.Error(w, "Error rendering status page", http.StatusInternalServerError)
		return
	}
}