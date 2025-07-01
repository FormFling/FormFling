package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"formfling/internal/config"
	"formfling/internal/models"
	"formfling/internal/services"
)

// Mock email service for testing
type mockEmailService struct {
	shouldFail bool
}

func (m *mockEmailService) SendEmail(formData models.FormData, origin string) error {
	if m.shouldFail {
		return errors.New("mock email service error")
	}
	return nil
}

// Ensure mockEmailService implements EmailSender
var _ services.EmailSender = (*mockEmailService)(nil)

func TestSubmitHandler_RedirectMode(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "test@example.com",
		ToEmail:      "recipient@example.com",
		FormTitle:    "Test Form",
	}

	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	// Test form submission without AJAX headers (should redirect)
	formData := url.Values{
		"name":    {"John Doe"},
		"email":   {"john@example.com"},
		"message": {"This is a test message with enough characters"},
	}

	req, err := http.NewRequest("POST", "/submit", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://example.com/contact")

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should redirect (303 status)
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %v", status)
	}

	// Should have location header with success status
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "/status?type=success&redirect=https%3A%2F%2Fexample.com%2Fcontact") {
		t.Error(
			"Expected redirect URL to contain ",
			"/status?type=success&redirect=https%3A%2F%2Fexample.com%2Fcontact",
			", got ", location,
		)
	}
}

func TestSubmitHandler_AjaxMode(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "test@example.com",
		ToEmail:      "recipient@example.com",
		FormTitle:    "Test Form",
	}

	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	formData := url.Values{
		"name":    {"John Doe"},
		"email":   {"john@example.com"},
		"message": {"This is a test message with enough characters"},
	}

	req, err := http.NewRequest("POST", "/submit", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest") // AJAX indicator

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should return JSON (200 status)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status 200, got %v", status)
	}

	// Should have JSON content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", contentType)
	}

	// Should contain success message
	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "message sent") {
		t.Errorf("Expected response to contain 'message sent', got %s", responseBody)
	}
}

func TestSubmitHandler_JSONRequest_Success(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "test@example.com",
		ToEmail:      "recipient@example.com",
		FormTitle:    "Test Form",
	}

	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	// Create JSON request body
	formData := models.FormData{
		Name:    "John Doe",
		Email:   "john@example.com",
		Message: "This is a test message with enough characters",
	}

	jsonData, err := json.Marshal(formData)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/submit", bytes.NewReader(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should return JSON (200 status) because of JSON content type
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status 200, got %v", status)
	}

	// Should have JSON content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", contentType)
	}

	// Parse response
	var response models.Response
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not unmarshal response: %v", err)
	}

	if response.Status != "message sent" {
		t.Errorf("Expected status 'message sent', got %s", response.Status)
	}
}

func TestSubmitHandler_JSONRequest_ValidationError(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "test@example.com",
		ToEmail:      "recipient@example.com",
		FormTitle:    "Test Form",
	}

	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	// Create JSON request body with invalid data
	formData := models.FormData{
		Name:    "",
		Email:   "invalid-email",
		Message: "short",
	}

	jsonData, err := json.Marshal(formData)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/submit", bytes.NewReader(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should return error status
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %v", status)
	}

	// Parse response
	var response models.Response
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not unmarshal response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got %s", response.Status)
	}
}

func TestSubmitHandler_JSONRequest_ParseError(t *testing.T) {
	cfg := &config.Config{
		FormTitle: "Test Form",
	}

	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	// Create invalid JSON
	invalidJSON := `{"name": "John", "email": }`

	req, err := http.NewRequest("POST", "/submit", strings.NewReader(invalidJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should return error status
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %v", status)
	}

	// Parse response
	var response models.Response
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not unmarshal response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got %s", response.Status)
	}

	if response.Error != "failed to parse JSON" {
		t.Errorf("Expected error 'failed to parse JSON', got %s", response.Error)
	}
}

func TestSubmitHandler_CustomRedirect(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "test@example.com",
		ToEmail:      "recipient@example.com",
		FormTitle:    "Test Form",
	}

	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	// Test with custom redirect URL
	formData := url.Values{
		"name":      {"John Doe"},
		"email":     {"john@example.com"},
		"message":   {"This is a test message with enough characters"},
		"_redirect": {"https://example.com/thank-you"},
	}

	req, err := http.NewRequest("POST", "/submit", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should redirect to custom URL
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "example.com/thank-you") {
		t.Errorf("Expected redirect to custom URL, got %s", location)
	}
	if !strings.Contains(location, "formfling_status=success") {
		t.Errorf("Expected redirect URL to contain status parameter, got %s", location)
	}
}

func TestSubmitHandler_ValidationError(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "test@example.com",
		ToEmail:      "recipient@example.com",
		FormTitle:    "Test Form",
	}

	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	// Test with invalid data (missing required fields)
	formData := url.Values{
		"name":    {""},
		"email":   {"invalid-email"},
		"message": {"short"},
	}

	req, err := http.NewRequest("POST", "/submit", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://example.com/contact")

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should redirect with error status
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %v", status)
	}

	location := rr.Header().Get("Location")
	if !strings.Contains(location, "/status?type=error&redirect=https%3A%2F%2Fexample.com%2Fcontact") {
		t.Error(
			"Expected redirect URL to contain ",
			"/status?type=error&redirect=https%3A%2F%2Fexample.com%2Fcontact",
			", got ", location,
		)
	}
}

func TestSubmitHandler_AjaxValidationError(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "test@example.com",
		ToEmail:      "recipient@example.com",
		FormTitle:    "Test Form",
	}

	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	// Test AJAX request with invalid data
	formData := url.Values{
		"name":    {""},
		"email":   {"invalid-email"},
		"message": {"short"},
	}

	req, err := http.NewRequest("POST", "/submit", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should return JSON error
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %v", status)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", contentType)
	}

	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "error") {
		t.Errorf("Expected response to contain error, got %s", responseBody)
	}
}

func TestSubmitHandler_MethodNotAllowed(t *testing.T) {
	cfg := &config.Config{
		FormTitle: "Test Form",
	}
	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	// Test GET request (should fail)
	req, err := http.NewRequest("GET", "/submit", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should redirect with error status
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %v", status)
	}

	location := rr.Header().Get("Location")
	// Should redirect to status page since there's no referer
	if !strings.Contains(location, "/status?type=error") {
		t.Errorf("Expected redirect URL to contain /status?type=error, got %s", location)
	}
}

func TestSubmitHandler_EmailServiceError(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "test@example.com",
		ToEmail:      "recipient@example.com",
		FormTitle:    "Test Form",
	}

	// Mock email service that fails
	emailService := &mockEmailService{shouldFail: true}
	handler := NewSubmitHandler(cfg, emailService, nil)

	formData := url.Values{
		"name":    {"John Doe"},
		"email":   {"john@example.com"},
		"message": {"This is a test message with enough characters"},
	}

	req, err := http.NewRequest("POST", "/submit", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest") // AJAX mode

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should return error status
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", status)
	}

	// Should return JSON error
	var response models.Response
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not unmarshal response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got %s", response.Status)
	}

	if response.Error != "failed to send email" {
		t.Errorf("Expected error 'failed to send email', got %s", response.Error)
	}
}

func TestIsAjaxRequest(t *testing.T) {
	cfg := &config.Config{}
	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	tests := []struct {
		name     string
		headers  map[string]string
		expected bool
	}{
		{
			name: "XMLHttpRequest header",
			headers: map[string]string{
				"X-Requested-With": "XMLHttpRequest",
			},
			expected: true,
		},
		{
			name: "JSON content type",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expected: true,
		},
		{
			name: "JSON accept header",
			headers: map[string]string{
				"Accept": "application/json",
			},
			expected: true,
		},
		{
			name: "Mixed accept header with JSON",
			headers: map[string]string{
				"Accept": "text/html,application/json,*/*",
			},
			expected: true,
		},
		{
			name: "Regular form request",
			headers: map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			},
			expected: false,
		},
		{
			name:     "No special headers",
			headers:  map[string]string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/submit", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := handler.isAjaxRequest(req)
			if result != tt.expected {
				t.Errorf("isAjaxRequest() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetRedirectURL(t *testing.T) {
	cfg := &config.Config{}
	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	tests := []struct {
		name         string
		formData     url.Values
		referer      string
		status       string
		containsText string
	}{
		{
			name: "Custom redirect URL",
			formData: url.Values{
				"_redirect": {"https://example.com/thank-you"},
			},
			status:       "success",
			containsText: "formfling_status=success",
		},
		{
			name:         "Referer with success",
			formData:     url.Values{},
			referer:      "https://example.com/contact",
			status:       "success",
			containsText: "/status?type=success&redirect=https%3A%2F%2Fexample.com%2Fcontact",
		},
		{
			name:         "No referer",
			formData:     url.Values{},
			referer:      "",
			status:       "error",
			containsText: "/status?type=error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/submit", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.referer != "" {
				req.Header.Set("Referer", tt.referer)
			}
			req.ParseForm()

			result := handler.getRedirectURL(req, tt.status)

			if !strings.Contains(result, tt.containsText) {
				t.Errorf("Expected URL to contain %s, got %s", tt.containsText, result)
			}
		})
	}
}

func TestAddStatusParam(t *testing.T) {
	cfg := &config.Config{}
	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	tests := []struct {
		name     string
		baseURL  string
		status   string
		expected string
	}{
		{
			name:     "Simple URL",
			baseURL:  "https://example.com/contact",
			status:   "success",
			expected: "formfling_status=success",
		},
		{
			name:     "URL with existing params",
			baseURL:  "https://example.com/contact?foo=bar",
			status:   "error",
			expected: "formfling_status=error",
		},
		{
			name:     "Invalid URL gets parameter anyway",
			baseURL:  "not-a-valid-url",
			status:   "success",
			expected: "formfling_status=success", // url.Parse is lenient
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.addStatusParam(tt.baseURL, tt.status)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected URL to contain %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSubmitHandler_ParseFormError(t *testing.T) {
	cfg := &config.Config{
		FormTitle: "Test Form",
	}
	emailService := &mockEmailService{}
	handler := NewSubmitHandler(cfg, emailService, nil)

	// Create a request with malformed form data
	req, err := http.NewRequest("POST", "/submit", strings.NewReader("%"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Should redirect with error status
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %v", status)
	}

	location := rr.Header().Get("Location")
	// Should redirect to status page since there's no referer
	if !strings.Contains(location, "/status?type=error") {
		t.Errorf("Expected redirect URL to contain /status?type=error, got %s", location)
	}
}
