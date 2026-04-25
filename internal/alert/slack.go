package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SlackWriter sends alert messages to a Slack incoming webhook URL.
type SlackWriter struct {
	webhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// NewSlackWriter creates a SlackWriter targeting the given Slack webhook URL.
func NewSlackWriter(webhookURL string) *SlackWriter {
	return &SlackWriter{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Write sends the message payload to Slack.
func (s *SlackWriter) Write(message string) error {
	payload := slackPayload{Text: message}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: failed to marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack: unexpected status code %d: %s", resp.StatusCode, bytes.TrimSpace(respBody))
	}
	return nil
}
