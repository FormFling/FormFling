package utils

import (
	"fmt"
	"regexp"
	"strings"

	"formfling/models"
)

func CleanString(input string) string {
	// Remove potentially dangerous content
	bad := []string{
		"content-type",
		"bcc:",
		"to:",
		"cc:",
		"href",
	}

	result := input
	for _, badWord := range bad {
		result = strings.ReplaceAll(strings.ToLower(result), badWord, "")
	}

	return strings.TrimSpace(result)
}

func ValidateEmail(email string) bool {
	// RFC 5322 compliant email regex (simplified)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func ValidateForm(form models.FormData) error {
	if strings.TrimSpace(form.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if strings.TrimSpace(form.Email) == "" {
		return fmt.Errorf("email is required")
	}

	if !ValidateEmail(form.Email) {
		return fmt.Errorf("email not valid")
	}

	if len(strings.TrimSpace(form.Message)) < 10 {
		return fmt.Errorf("message not valid")
	}

	return nil
}
