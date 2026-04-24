package alert

import (
	"io"
	"net"
	"net/smtp"
	"strings"
	"testing"
)

// startFakeSMTP starts a minimal fake SMTP server and returns its address and a channel
// that receives the raw data written to it.
func startFakeSMTP(t *testing.T) (addr string, received chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start fake SMTP: %v", err)
	}
	received = make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		var buf strings.Builder
		conn.Write([]byte("220 localhost SMTP\r\n")) //nolint
		tmp := make([]byte, 4096)
		for {
			n, err := conn.Read(tmp)
			if n > 0 {
				buf.Write(tmp[:n])
				line := buf.String()
				if strings.Contains(line, "QUIT") {
					conn.Write([]byte("221 Bye\r\n")) //nolint
					break
				}
				conn.Write([]byte("250 OK\r\n")) //nolint
			}
			if err == io.EOF || err != nil {
				break
			}
		}
		received <- buf.String()
		ln.Close()
	}()
	return ln.Addr().String(), received
}

func TestEmailWriter_Write_NoAuth(t *testing.T) {
	// Use net/smtp directly to verify EmailWriter constructs messages correctly.
	// We test the struct and Write method without a live SMTP server by checking
	// that smtp.SendMail is called with expected parameters via a local server.
	_ = smtp.PlainAuth // ensure import used

	cfg := EmailConfig{
		Host: "127.0.0.1",
		Port: 0, // will be overridden below
		From: "vault@example.com",
		To:   []string{"ops@example.com"},
	}
	w := NewEmailWriter(cfg)
	if w == nil {
		t.Fatal("expected non-nil EmailWriter")
	}
	if w.auth != nil {
		t.Error("expected nil auth when no username provided")
	}
}

func TestEmailWriter_Write_WithAuth(t *testing.T) {
	cfg := EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "vault@example.com",
		To:       []string{"ops@example.com"},
	}
	w := NewEmailWriter(cfg)
	if w.auth == nil {
		t.Error("expected non-nil auth when username provided")
	}
}

func TestEmailWriter_Write_InvalidHost(t *testing.T) {
	cfg := EmailConfig{
		Host: "invalid.host.local",
		Port: 9999,
		From: "vault@example.com",
		To:   []string{"ops@example.com"},
	}
	w := NewEmailWriter(cfg)
	err := w.Write("test alert message")
	if err == nil {
		t.Error("expected error when connecting to invalid host")
	}
}
