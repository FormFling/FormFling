package handlers

import (
	"html/template"
	"log"
	"net/http"

	"formfling/internal/config"
)

type TestFormHandler struct {
	config           *config.Config
	testFormTemplate *template.Template
}

func NewTestFormHandler(cfg *config.Config) *TestFormHandler {
	testFormTemplate, err := template.ParseFiles(cfg.TestFormTemplate)
	if err != nil {
		log.Fatal("Error loading status template:", err)
	}

	return &TestFormHandler{
		config:           cfg,
		testFormTemplate: testFormTemplate,
	}
}

type TestFormData struct {
	SiteKey string
}

func (h *TestFormHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := TestFormData{
		SiteKey: h.config.RecaptchaSiteKey,
	}

	if err := h.testFormTemplate.Execute(w, data); err != nil {
		http.Error(w, "Error rendering status page", http.StatusInternalServerError)
		return
	}
}
