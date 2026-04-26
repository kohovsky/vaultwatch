package monitor

import (
	"testing"
	"time"
)

func TestDedupWindow_FirstCallNotDuplicate(t *testing.T) {
	d := NewDedupWindow(time.Minute)
	if d.IsDuplicate(DedupKey("secret/foo", "warning")) {
		t.Fatal("expected first call to not be a duplicate")
	}
}

func TestDedupWindow_SecondCallIsDuplicate(t *testing.T) {
	d := NewDedupWindow(time.Minute)
	key := DedupKey("secret/foo", "warning")
	d.IsDuplicate(key)
	if !d.IsDuplicate(key) {
		t.Fatal("expected second call within cooldown to be a duplicate")
	}
}

func TestDedupWindow_DifferentStatusNotDuplicate(t *testing.T) {
	d := NewDedupWindow(time.Minute)
	d.IsDuplicate(DedupKey("secret/foo", "warning"))
	if d.IsDuplicate(DedupKey("secret/foo", "critical")) {
		t.Fatal("expected different status to not be a duplicate")
	}
}

func TestDedupWindow_AfterCooldownNotDuplicate(t *testing.T) {
	d := NewDedupWindow(10 * time.Millisecond)
	key := DedupKey("secret/bar", "expired")
	d.IsDuplicate(key)
	time.Sleep(20 * time.Millisecond)
	if d.IsDuplicate(key) {
		t.Fatal("expected call after cooldown expiry to not be a duplicate")
	}
}

func TestDedupWindow_DefaultCooldown(t *testing.T) {
	d := NewDedupWindow(0)
	if d.cooldown != time.Hour {
		t.Fatalf("expected default cooldown of 1h, got %v", d.cooldown)
	}
}

func TestDedupWindow_Evict(t *testing.T) {
	d := NewDedupWindow(10 * time.Millisecond)
	key := DedupKey("secret/evict", "warning")
	d.IsDuplicate(key)
	time.Sleep(20 * time.Millisecond)
	d.Evict()
	d.mu.Lock()
	_, exists := d.seen[key]
	d.mu.Unlock()
	if exists {
		t.Fatal("expected expired entry to be evicted")
	}
}

func TestDedupKey_Format(t *testing.T) {
	key := DedupKey("secret/myapp/db", "critical")
	expected := "secret/myapp/db::critical"
	if key != expected {
		t.Fatalf("expected %q, got %q", expected, key)
	}
}
