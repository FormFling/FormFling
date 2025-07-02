package utils

import (
	"testing"

	"formfling/internal/models"
)

func TestCleanString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal text", "normal text"},
		{"content-type: text/html", ": text/html"},
		{"bcc:evil@example.com", "evil@example.com"},
		{"  whitespace  ", "whitespace"},
		{"TO:someone@test.com", "someone@test.com"},
		{"href=\"malicious\"", "=\"malicious\""},
	}

	for _, test := range tests {
		result := CleanString(test.input)
		if result != test.expected {
			t.Errorf("CleanString(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestValidateEmail(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user+tag@domain.co.uk",
		"first.last@subdomain.example.org",
		"123@test.io",
	}

	invalidEmails := []string{
		"",
		"notanemail",
		"@example.com",
		"test@",
		"test.example.com",
		"test@.com",
		"test@com",
		"test space@example.com",
	}

	for _, email := range validEmails {
		if !ValidateEmail(email) {
			t.Errorf("ValidateEmail(%q) should be true", email)
		}
	}

	for _, email := range invalidEmails {
		if ValidateEmail(email) {
			t.Errorf("ValidateEmail(%q) should be false", email)
		}
	}
}

func TestValidateForm(t *testing.T) {
	// Valid form
	validForm := models.FormData{
		Name:    "John Doe",
		Email:   "john@example.com",
		Message: "This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters.",
	}

	if err := ValidateForm(validForm); err != nil {
		t.Errorf("ValidateForm should pass for valid form, got error: %v", err)
	}

	// Missing name
	invalidForm := models.FormData{
		Name:    "",
		Email:   "john@example.com",
		Message: "This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters.",
	}

	if err := ValidateForm(invalidForm); err == nil {
		t.Error("ValidateForm should fail for missing name")
	}

	// Missing email
	invalidForm = models.FormData{
		Name:    "John Doe",
		Email:   "",
		Message: "This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters.",
	}

	if err := ValidateForm(invalidForm); err == nil {
		t.Error("ValidateForm should fail for missing email")
	}

	// Invalid email
	invalidForm = models.FormData{
		Name:    "John Doe",
		Email:   "not-an-email",
		Message: "This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters. This is a valid message with enough characters.",
	}

	if err := ValidateForm(invalidForm); err == nil {
		t.Error("ValidateForm should fail for invalid email")
	}

	// Message too short
	invalidForm = models.FormData{
		Name:    "John Doe",
		Email:   "john@example.com",
		Message: "short",
	}

	if err := ValidateForm(invalidForm); err == nil {
		t.Error("ValidateForm should fail for message too short")
	}
}
