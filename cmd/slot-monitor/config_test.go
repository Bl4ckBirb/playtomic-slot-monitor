package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

const goodConfig = `{
  "timezone": "Europe/Berlin",
  "look_ahead_days": 7,
  "clubs": [{ "tenant_id": "t1", "name": "Club" }],
  "watch_windows": [{ "days": ["mon"], "start": "17:30", "end": "21:00" }],
  "filters": { "durations": [90] }
}`

func TestLoadConfigGood(t *testing.T) {
	cfg, err := loadConfig(writeTemp(t, goodConfig))
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(cfg.Clubs) != 1 || cfg.Clubs[0].TenantID != "t1" {
		t.Errorf("clubs = %+v", cfg.Clubs)
	}
	if _, err := cfg.parseWindows(); err != nil {
		t.Errorf("parseWindows: %v", err)
	}
	if _, err := cfg.location(); err != nil {
		t.Errorf("location: %v", err)
	}
}

func TestLoadConfigMissingComma(t *testing.T) {
	// Missing comma after the timezone line.
	broken := `{
  "timezone": "Europe/Berlin"
  "look_ahead_days": 7,
  "clubs": [{ "tenant_id": "t1" }],
  "watch_windows": [{ "days": ["mon"], "start": "17:30", "end": "21:00" }]
}`
	_, err := loadConfig(writeTemp(t, broken))
	if err == nil {
		t.Fatal("expected error for missing comma")
	}
	msg := err.Error()
	if !strings.Contains(msg, "line 3") {
		t.Errorf("error should point at line 3, got: %s", msg)
	}
	if !strings.Contains(msg, "syntax error") {
		t.Errorf("error should mention syntax error, got: %s", msg)
	}
}

func TestLoadConfigUnknownField(t *testing.T) {
	// "duration" instead of "durations".
	broken := `{
  "clubs": [{ "tenant_id": "t1" }],
  "watch_windows": [{ "days": ["mon"], "start": "17:30", "end": "21:00" }],
  "filters": { "duration": [90] }
}`
	_, err := loadConfig(writeTemp(t, broken))
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
	if !strings.Contains(err.Error(), "duration") {
		t.Errorf("error should name the unknown field, got: %s", err)
	}
}

func TestLoadConfigTypeError(t *testing.T) {
	// look_ahead_days as a string, not a number.
	broken := `{
  "look_ahead_days": "seven",
  "clubs": [{ "tenant_id": "t1" }],
  "watch_windows": [{ "days": ["mon"], "start": "17:30", "end": "21:00" }]
}`
	_, err := loadConfig(writeTemp(t, broken))
	if err == nil {
		t.Fatal("expected type error")
	}
	if !strings.Contains(err.Error(), "look_ahead_days") {
		t.Errorf("error should name the field, got: %s", err)
	}
}

func TestLoadConfigBadDay(t *testing.T) {
	broken := `{
  "clubs": [{ "tenant_id": "t1" }],
  "watch_windows": [{ "days": ["montag"], "start": "17:30", "end": "21:00" }]
}`
	cfg, err := loadConfig(writeTemp(t, broken))
	if err != nil {
		t.Fatalf("loadConfig should accept structurally valid JSON: %v", err)
	}
	if _, err := cfg.parseWindows(); err == nil || !strings.Contains(err.Error(), "montag") {
		t.Errorf("parseWindows should reject 'montag', got: %v", err)
	}
}
