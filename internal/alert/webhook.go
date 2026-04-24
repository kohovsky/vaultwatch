package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookPayload is the JSON body sent to a webhook endpoint.
type WebhookPayload struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Secret    string    `json:"secret"`
	ExpiresAt time.Time `json:"expires_at"`
	Timestamp time.Time `json:"timestamp"`
}

// WebhookWriter sends alert notifications to an HTTP webhook endpoint.
type WebhookWriter struct {
	URL    string
	Client *http.Client
}

// NewWebhookWriter creates a WebhookWriter with a default HTTP client.
func NewWebhookWriter(url string) *WebhookWriter {
	return &WebhookWriter{
		URL: url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Write encodes the payload as JSON and POSTs it to the configured URL.
func (w *WebhookWriter) Write(payload WebhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.Client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d from %s", resp.StatusCode, w.URL)
	}
	return nil
}
