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
	originalRecaptchaSecret := os.Getenv("RECAPTCHA_SECRET_KEY")
	originalRecaptchaScore := os.Getenv("RECAPTCHA_MIN_SCORE")
	originalRecaptchaAction := os.Getenv("RECAPTCHA_ACTION")

	// Clean up after test
	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("SMTP_HOST", originalSMTPHost)
		os.Setenv("ALLOWED_ORIGINS", originalOrigins)
		os.Setenv("EMAIL_TEMPLATE", originalEmailTemplate)
		os.Setenv("RECAPTCHA_SECRET_KEY", originalRecaptchaSecret)
		os.Setenv("RECAPTCHA_MIN_SCORE", originalRecaptchaScore)
		os.Setenv("RECAPTCHA_ACTION", originalRecaptchaAction)
	}()

	// Test default values
	os.Unsetenv("PORT")
	os.Unsetenv("SMTP_HOST")
	os.Unsetenv("ALLOWED_ORIGINS")
	os.Unsetenv("EMAIL_TEMPLATE")
	os.Unsetenv("RECAPTCHA_SECRET_KEY")
	os.Unsetenv("RECAPTCHA_MIN_SCORE")
	os.Unsetenv("RECAPTCHA_ACTION")

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

	if cfg.EmailTemplate != "./web/templates/email_template.html" {
		t.Errorf("Expected email template './web/templates/email_template.html', got %s", cfg.EmailTemplate)
	}

	if cfg.StatusTemplate != "./web/templates/status_template.html" {
		t.Errorf("Expected status template './web/templates/status_template.html', got %s", cfg.StatusTemplate)
	}

	// Test reCAPTCHA defaults
	if cfg.RecaptchaEnabled {
		t.Error("Expected reCAPTCHA to be disabled by default")
	}

	if cfg.RecaptchaMinScore != 0.5 {
		t.Errorf("Expected default reCAPTCHA min score 0.5, got %f", cfg.RecaptchaMinScore)
	}

	if cfg.RecaptchaAction != "submit" {
		t.Errorf("Expected default reCAPTCHA action 'submit', got %s", cfg.RecaptchaAction)
	}

	// Test custom values
	os.Setenv("PORT", "3000")
	os.Setenv("SMTP_HOST", "smtp.example.com")
	os.Setenv("ALLOWED_ORIGINS", "https://example.com,https://test.com")
	os.Setenv("EMAIL_TEMPLATE", "/custom/email_template.html")
	os.Setenv("STATUS_TEMPLATE", "/custom/status_template.html")
	os.Setenv("RECAPTCHA_SECRET_KEY", "test-secret-key")
	os.Setenv("RECAPTCHA_MIN_SCORE", "0.8")
	os.Setenv("RECAPTCHA_ACTION", "contact")

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

	if cfg.EmailTemplate != "/custom/email_template.html" {
		t.Errorf("Expected email template '/custom/email_template.html', got %s", cfg.EmailTemplate)
	}

	if cfg.StatusTemplate != "/custom/status_template.html" {
		t.Errorf("Expected status template '/custom/status_template.html', got %s", cfg.StatusTemplate)
	}

	// Test reCAPTCHA custom values
	if !cfg.RecaptchaEnabled {
		t.Error("Expected reCAPTCHA to be enabled when secret key is provided")
	}

	if cfg.RecaptchaSecretKey != "test-secret-key" {
		t.Errorf("Expected reCAPTCHA secret key 'test-secret-key', got %s", cfg.RecaptchaSecretKey)
	}

	if cfg.RecaptchaMinScore != 0.8 {
		t.Errorf("Expected reCAPTCHA min score 0.8, got %f", cfg.RecaptchaMinScore)
	}

	if cfg.RecaptchaAction != "contact" {
		t.Errorf("Expected reCAPTCHA action 'contact', got %s", cfg.RecaptchaAction)
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

func TestGetEnvAsFloat(t *testing.T) {
	result := getEnvAsFloat("NONEXISTENT_VAR", 1.23)
	if result != 1.23 {
		t.Errorf("Expected 1.23, got %f", result)
	}

	os.Setenv("TEST_FLOAT_VAR", "4.56")
	defer os.Unsetenv("TEST_FLOAT_VAR")

	result = getEnvAsFloat("TEST_FLOAT_VAR", 1.23)
	if result != 4.56 {
		t.Errorf("Expected 4.56, got %f", result)
	}

	// Test invalid float
	os.Setenv("TEST_INVALID_FLOAT", "not_a_number")
	defer os.Unsetenv("TEST_INVALID_FLOAT")

	result = getEnvAsFloat("TEST_INVALID_FLOAT", 7.89)
	if result != 7.89 {
		t.Errorf("Expected 7.89 (default), got %f", result)
	}
}
