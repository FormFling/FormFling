package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"formfling/internal/models"
)

func TestHealthHandler(t *testing.T) {
	handler := NewHealthHandler()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(handler.Handle).ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check content type
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("Health handler returned wrong content type: got %v want %v", ct, expected)
	}

	// Check response body
	var response models.Response
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got %s", response.Status)
	}

	if response.Error != "" {
		t.Errorf("Expected no error, got %s", response.Error)
	}
}
