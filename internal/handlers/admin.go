package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"formfling/internal/config"
	"formfling/internal/models"
)

type AdminHandler struct {
	config        *config.Config
	db            *sql.DB
	adminTemplate *template.Template
}

func NewAdminHandler(cfg *config.Config, db *sql.DB) *AdminHandler {
	adminTemplate, err := template.ParseFiles("./web/templates/admin.html")
	if err != nil {
		log.Fatal("Error loading admin template:", err)
	}

	return &AdminHandler{
		config:        cfg,
		db:            db,
		adminTemplate: adminTemplate,
	}
}

func (h *AdminHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := models.Response{Status: "admin"}
	json.NewEncoder(w).Encode(response)
}
