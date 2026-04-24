package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSlackWriter_Success(t *testing.T) {
	var received slackPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	w := NewSlackWriter(server.URL)
	err := w.Write("test alert message")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if received.Text != "test alert message" {
		t.Errorf("expected message %q, got %q", "test alert message", received.Text)
	}
}

func TestSlackWriter_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	w := NewSlackWriter(server.URL)
	err := w.Write("alert")
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestSlackWriter_InvalidURL(t *testing.T) {
	w := NewSlackWriter("http://127.0.0.1:0/invalid")
	err := w.Write("alert")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestSlackWriter_ContentType(t *testing.T) {
	var contentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	w := NewSlackWriter(server.URL)
	_ = w.Write("hello")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}
}
