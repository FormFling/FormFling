package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	SMTPHost           string
	SMTPPort           int
	SMTPUsername       string
	SMTPPassword       string
	FromEmail          string
	FromName           string
	ToEmail            string
	ToName             string
	AllowedOrigins     []string
	FormTitle          string
	EmailTemplate      string
	StatusTemplate     string
	TestFormTemplate   string
	EnableTestForm     bool
	RecaptchaEnabled   bool
	RecaptchaSiteKey   string
	RecaptchaSecretKey string
	RecaptchaMinScore  float64
	RecaptchaAction    string
}

func Load() *Config {
	config := &Config{
		Port:               getEnv("PORT", "8080"),
		SMTPHost:           getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:           getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername:       getEnv("SMTP_USERNAME", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		FromEmail:          getEnv("FROM_EMAIL", ""),
		FromName:           getEnv("FROM_NAME", "FormFling"),
		ToEmail:            getEnv("TO_EMAIL", ""),
		ToName:             getEnv("TO_NAME", ""),
		FormTitle:          getEnv("FORM_TITLE", "Contact Me"),
		EmailTemplate:      getEnv("EMAIL_TEMPLATE", "./web/templates/email_template.html"),
		StatusTemplate:     getEnv("STATUS_TEMPLATE", "./web/templates/status_template.html"),
		TestFormTemplate:   getEnv("TEST_FORM_TEMPLATE", "./web/templates/test_form_template.html"),
		EnableTestForm:     getEnvAsBool("ENABLE_TEST_FORM", false),
		RecaptchaSiteKey:   getEnv("RECAPTCHA_SITE_KEY", ""),
		RecaptchaSecretKey: getEnv("RECAPTCHA_SECRET_KEY", ""),
		RecaptchaMinScore:  getEnvAsFloat("RECAPTCHA_MIN_SCORE", 0.5),
		RecaptchaAction:    getEnv("RECAPTCHA_ACTION", "submit"),
	}

	// Enable reCAPTCHA if secret key is provided
	config.RecaptchaEnabled = config.RecaptchaSecretKey != ""

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

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
