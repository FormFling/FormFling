package services

import "formfling/internal/models"

// EmailSender defines the interface for sending emails
type EmailSender interface {
	SendEmail(formData models.FormData, origin string) error
}

// Ensure EmailService implements EmailSender
var _ EmailSender = (*EmailService)(nil)
