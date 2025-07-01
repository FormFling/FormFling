package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"formfling/internal/config"
)

// RecaptchaResponse represents the response from Google's reCAPTCHA verify API
type RecaptchaResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

// RecaptchaService handles reCAPTCHA v3 verification
type RecaptchaService struct {
	config *config.Config
	client *http.Client
}

// NewRecaptchaService creates a new reCAPTCHA verification service
func NewRecaptchaService(cfg *config.Config) *RecaptchaService {
	return &RecaptchaService{
		config: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// VerifyToken verifies a reCAPTCHA v3 token with Google's API
func (rs *RecaptchaService) VerifyToken(token, remoteIP string) error {
	if !rs.config.RecaptchaEnabled {
		return nil // reCAPTCHA is disabled, skip verification
	}

	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("reCAPTCHA token is required")
	}

	// Prepare the verification request
	data := url.Values{
		"secret":   {rs.config.RecaptchaSecretKey},
		"response": {token},
	}

	// Add remote IP if provided (optional but recommended)
	if remoteIP != "" {
		data.Set("remoteip", remoteIP)
	}

	// Make request to Google's verify API
	resp, err := rs.client.PostForm("https://www.google.com/recaptcha/api/siteverify", data)
	if err != nil {
		return fmt.Errorf("failed to verify reCAPTCHA: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var recaptchaResp RecaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&recaptchaResp); err != nil {
		return fmt.Errorf("failed to parse reCAPTCHA response: %v", err)
	}

	// Check if verification was successful
	if !recaptchaResp.Success {
		errorMsg := "reCAPTCHA verification failed"
		if len(recaptchaResp.ErrorCodes) > 0 {
			errorMsg += ": " + strings.Join(recaptchaResp.ErrorCodes, ", ")
		}
		return fmt.Errorf(errorMsg)
	}

	// Check the score (v3 specific)
	if recaptchaResp.Score < rs.config.RecaptchaMinScore {
		return fmt.Errorf("reCAPTCHA score too low: %s (minimum: %s)",
			formatScore(recaptchaResp.Score),
			formatScore(rs.config.RecaptchaMinScore))
	}

	// Check the action if configured
	if rs.config.RecaptchaAction != "" && recaptchaResp.Action != rs.config.RecaptchaAction {
		return fmt.Errorf("reCAPTCHA action mismatch: expected %s, got %s",
			rs.config.RecaptchaAction, recaptchaResp.Action)
	}

	return nil // Verification successful
}

// formatScore formats a float64 score for display
func formatScore(score float64) string {
	return strconv.FormatFloat(score, 'f', 2, 64)
}
