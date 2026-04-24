package monitor_test

import (
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/monitor"
	"github.com/user/vaultwatch/internal/vault"
)

func makeSecret(path string, expiresIn time.Duration) vault.SecretInfo {
	return vault.SecretInfo{
		Path:      path,
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func TestCheckExpiry_OK(t *testing.T) {
	secret := makeSecret("secret/db", 48*time.Hour)
	alert := monitor.CheckExpiry(secret, 24*time.Hour, 6*time.Hour)
	if alert.Status != monitor.StatusOK {
		t.Errorf("expected StatusOK, got %s", alert.Status)
	}
}

func TestCheckExpiry_Warning(t *testing.T) {
	secret := makeSecret("secret/db", 12*time.Hour)
	alert := monitor.CheckExpiry(secret, 24*time.Hour, 6*time.Hour)
	if alert.Status != monitor.StatusWarning {
		t.Errorf("expected StatusWarning, got %s", alert.Status)
	}
}

func TestCheckExpiry_Critical(t *testing.T) {
	secret := makeSecret("secret/db", 2*time.Hour)
	alert := monitor.CheckExpiry(secret, 24*time.Hour, 6*time.Hour)
	if alert.Status != monitor.StatusCritical {
		t.Errorf("expected StatusCritical, got %s", alert.Status)
	}
}

func TestCheckExpiry_Expired(t *testing.T) {
	secret := makeSecret("secret/db", -1*time.Minute)
	alert := monitor.CheckExpiry(secret, 24*time.Hour, 6*time.Hour)
	if alert.Status != monitor.StatusExpired {
		t.Errorf("expected StatusExpired, got %s", alert.Status)
	}
}

func TestCheckAll_FiltersOK(t *testing.T) {
	secrets := []vault.SecretInfo{
		makeSecret("secret/ok", 72*time.Hour),
		makeSecret("secret/warn", 12*time.Hour),
		makeSecret("secret/crit", 2*time.Hour),
		makeSecret("secret/exp", -5*time.Minute),
	}
	alerts := monitor.CheckAll(secrets, 24*time.Hour, 6*time.Hour)
	if len(alerts) != 3 {
		t.Errorf("expected 3 alerts, got %d", len(alerts))
	}
}

func TestFormatAlert_Expired(t *testing.T) {
	secret := makeSecret("secret/old", -10*time.Minute)
	alert := monitor.CheckExpiry(secret, 24*time.Hour, 6*time.Hour)
	msg := monitor.FormatAlert(alert)
	if len(msg) == 0 {
		t.Error("expected non-empty format string for expired secret")
	}
	if alert.Status.String() != "EXPIRED" {
		t.Errorf("expected EXPIRED string, got %s", alert.Status.String())
	}
}

func TestFormatAlert_Warning(t *testing.T) {
	secret := makeSecret("secret/soon", 10*time.Hour)
	alert := monitor.CheckExpiry(secret, 24*time.Hour, 6*time.Hour)
	msg := monitor.FormatAlert(alert)
	if len(msg) == 0 {
		t.Error("expected non-empty format string for warning secret")
	}
}
