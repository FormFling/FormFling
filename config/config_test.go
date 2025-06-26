package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalPort := os.Getenv("PORT")
	originalSMTPHost := os.Getenv("SMTP_HOST")
	originalOrigins := os.Getenv("ALLOWED_ORIGINS")
	originalEmailTemplate := os.Getenv("EMAIL_TEMPLATE")

	// Clean up after test
	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("SMTP_HOST", originalSMTPHost)
		os.Setenv("ALLOWED_ORIGINS", originalOrigins)
		os.Setenv("EMAIL_TEMPLATE", originalEmailTemplate)
	}()

	// Test default values
	os.Unsetenv("PORT")
	os.Unsetenv("SMTP_HOST")
	os.Unsetenv("ALLOWED_ORIGINS")
	os.Unsetenv("EMAIL_TEMPLATE")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Expected port 8080, got %s", cfg.Port)
	}

	if cfg.SMTPHost != "smtp.gmail.com" {
		t.Errorf("Expected smtp.gmail.com, got %s", cfg.SMTPHost)
	}

	if cfg.SMTPPort != 587 {
		t.Errorf("Expected SMTP port 587, got %d", cfg.SMTPPort)
	}

	if cfg.EmailTemplate != "email_template.html" {
		t.Errorf("Expected email template 'email_template.html', got %s", cfg.EmailTemplate)
	}

	// Test custom values
	os.Setenv("PORT", "3000")
	os.Setenv("SMTP_HOST", "smtp.example.com")
	os.Setenv("ALLOWED_ORIGINS", "https://example.com,https://test.com")
	os.Setenv("EMAIL_TEMPLATE", "/custom/template.html")

	cfg = Load()

	if cfg.Port != "3000" {
		t.Errorf("Expected port 3000, got %s", cfg.Port)
	}

	if cfg.SMTPHost != "smtp.example.com" {
		t.Errorf("Expected smtp.example.com, got %s", cfg.SMTPHost)
	}

	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("Expected 2 allowed origins, got %d", len(cfg.AllowedOrigins))
	}

	if cfg.AllowedOrigins[0] != "https://example.com" {
		t.Errorf("Expected https://example.com, got %s", cfg.AllowedOrigins[0])
	}

	if cfg.EmailTemplate != "/custom/template.html" {
		t.Errorf("Expected email template '/custom/template.html', got %s", cfg.EmailTemplate)
	}
}

func TestGetEnv(t *testing.T) {
	result := getEnv("NONEXISTENT_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got %s", result)
	}

	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result = getEnv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got %s", result)
	}
}

func TestGetEnvAsInt(t *testing.T) {
	result := getEnvAsInt("NONEXISTENT_VAR", 123)
	if result != 123 {
		t.Errorf("Expected 123, got %d", result)
	}

	os.Setenv("TEST_INT_VAR", "456")
	defer os.Unsetenv("TEST_INT_VAR")

	result = getEnvAsInt("TEST_INT_VAR", 123)
	if result != 456 {
		t.Errorf("Expected 456, got %d", result)
	}

	// Test invalid int
	os.Setenv("TEST_INVALID_INT", "not_a_number")
	defer os.Unsetenv("TEST_INVALID_INT")

	result = getEnvAsInt("TEST_INVALID_INT", 789)
	if result != 789 {
		t.Errorf("Expected 789 (default), got %d", result)
	}
}
