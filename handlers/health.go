package handlers

import (
	"encoding/json"
	"net/http"

	"formfling/models"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := models.Response{Status: "ok"}
	json.NewEncoder(w).Encode(response)
}
