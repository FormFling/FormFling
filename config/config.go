package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port           string
	SMTPHost       string
	SMTPPort       int
	SMTPUsername   string
	SMTPPassword   string
	FromEmail      string
	FromName       string
	ToEmail        string
	ToName         string
	AllowedOrigins []string
	FormTitle      string
	EmailTemplate  string
}

func Load() *Config {
	config := &Config{
		Port:           getEnv("PORT", "8080"),
		SMTPHost:       getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:       getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername:   getEnv("SMTP_USERNAME", ""),
		SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		FromEmail:      getEnv("FROM_EMAIL", ""),
		FromName:       getEnv("FROM_NAME", "FormFling Bot"),
		ToEmail:        getEnv("TO_EMAIL", ""),
		ToName:         getEnv("TO_NAME", ""),
		FormTitle:      getEnv("FORM_TITLE", "Contact Me"),
		EmailTemplate:  getEnv("EMAIL_TEMPLATE", "email_template.html"),
	}

	// Parse allowed origins
	allowedOriginsStr := getEnv("ALLOWED_ORIGINS", "*")
	if allowedOriginsStr != "*" && allowedOriginsStr != "" {
		config.AllowedOrigins = strings.Split(allowedOriginsStr, ",")
		for i := range config.AllowedOrigins {
			config.AllowedOrigins[i] = strings.TrimSpace(config.AllowedOrigins[i])
		}
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
