package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookWriter_Success(t *testing.T) {
	var received WebhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	w := NewWebhookWriter(server.URL)
	payload := WebhookPayload{
		Level:     "warning",
		Message:   "secret expiring soon",
		Secret:    "secret/db/password",
		ExpiresAt: time.Now().Add(48 * time.Hour),
		Timestamp: time.Now(),
	}

	if err := w.Write(payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Level != "warning" {
		t.Errorf("expected level 'warning', got %q", received.Level)
	}
	if received.Secret != "secret/db/password" {
		t.Errorf("expected secret path, got %q", received.Secret)
	}
}

func TestWebhookWriter_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	w := NewWebhookWriter(server.URL)
	err := w.Write(WebhookPayload{Level: "critical", Timestamp: time.Now()})
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestWebhookWriter_InvalidURL(t *testing.T) {
	w := NewWebhookWriter("http://127.0.0.1:0/nonexistent")
	err := w.Write(WebhookPayload{Level: "warning", Timestamp: time.Now()})
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}
