package services

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"time"

	"formfling/config"
	"formfling/models"
)

type EmailService struct {
	config        *config.Config
	emailTemplate *template.Template
}

func NewEmailService(cfg *config.Config) *EmailService {
	// Load email template
	emailTemplate, err := template.ParseFiles(cfg.EmailTemplate)
	if err != nil {
		log.Fatal("Error loading email template:", err)
	}

	return &EmailService{
		config:        cfg,
		emailTemplate: emailTemplate,
	}
}

func (s *EmailService) SendEmail(formData models.FormData, origin string) error {
	now := time.Now()
	templateData := models.EmailTemplateData{
		FormData:      formData,
		SubmittedTime: now.Format("03:04 PM"),
		SubmittedDate: now.Format("02 January 2006"),
		Origin:        origin,
	}

	var emailBody bytes.Buffer
	if err := s.emailTemplate.Execute(&emailBody, templateData); err != nil {
		return fmt.Errorf("error executing email template: %v", err)
	}

	// Set up authentication information
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// Create message
	subject := fmt.Sprintf("New submission from %s", s.config.FormTitle)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	msg := fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.FromEmail)
	msg += fmt.Sprintf("To: %s <%s>\r\n", s.config.ToName, s.config.ToEmail)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += mime
	msg += emailBody.String()

	// Connect to server and send email
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Handle different SMTP configurations
	if s.config.SMTPPort == 465 {
		// SSL/TLS connection for port 465
		return s.sendEmailSSL(addr, auth, msg)
	} else {
		// STARTTLS connection for port 587 (and others)
		return s.sendEmailSTARTTLS(addr, auth, msg)
	}
}

func (s *EmailService) sendEmailSSL(addr string, auth smtp.Auth, msg string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.config.SMTPHost,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Quit()

	// Authenticate
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}

	// Send email
	return s.sendEmailData(client, msg)
}

func (s *EmailService) sendEmailSTARTTLS(addr string, auth smtp.Auth, msg string) error {
	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Quit()

	// Start TLS if available
	if ok, _ := conn.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: s.config.SMTPHost,
		}
		if err := conn.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %v", err)
		}
	}

	// Authenticate
	if err := conn.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}

	// Send email
	return s.sendEmailData(conn, msg)
}

func (s *EmailService) sendEmailData(client *smtp.Client, msg string) error {
	// Set sender and recipient
	if err := client.Mail(s.config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	if err := client.Rcpt(s.config.ToEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send email body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %v", err)
	}

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email data: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close email writer: %v", err)
	}

	return nil
}
