package config

import (
	"testing"
)

func TestConfig_DefaultValues(t *testing.T) {
	// Unset all relevant env vars so defaults are used.
	t.Setenv("SERVER_PORT", "")
	t.Setenv("SEMANTIC_SERVICE_URL", "")
	t.Setenv("GAME_DURATION_SECONDS", "")
	t.Setenv("RATE_LIMIT_SECONDS", "")

	cfg := Load()

	if cfg.ServerPort != "8080" {
		t.Errorf("expected default ServerPort=8080, got %s", cfg.ServerPort)
	}
	if cfg.SemanticURL != "http://localhost:8001" {
		t.Errorf("expected default SemanticURL=http://localhost:8001, got %s", cfg.SemanticURL)
	}
	if cfg.GameDuration != 60 {
		t.Errorf("expected default GameDuration=60, got %d", cfg.GameDuration)
	}
	if cfg.RateLimitSeconds != 1 {
		t.Errorf("expected default RateLimitSeconds=1, got %d", cfg.RateLimitSeconds)
	}
}

func TestConfig_ServerPortOverride(t *testing.T) {
	t.Setenv("SERVER_PORT", "9090")
	cfg := Load()
	if cfg.ServerPort != "9090" {
		t.Errorf("expected ServerPort=9090, got %s", cfg.ServerPort)
	}
}

func TestConfig_SemanticURLOverride(t *testing.T) {
	t.Setenv("SEMANTIC_SERVICE_URL", "http://ml-service:8001")
	cfg := Load()
	if cfg.SemanticURL != "http://ml-service:8001" {
		t.Errorf("expected SemanticURL=http://ml-service:8001, got %s", cfg.SemanticURL)
	}
}

func TestConfig_GameDurationOverride(t *testing.T) {
	t.Setenv("GAME_DURATION_SECONDS", "90")
	cfg := Load()
	if cfg.GameDuration != 90 {
		t.Errorf("expected GameDuration=90, got %d", cfg.GameDuration)
	}
}

func TestConfig_RateLimitSecondsOverride(t *testing.T) {
	t.Setenv("RATE_LIMIT_SECONDS", "2")
	cfg := Load()
	if cfg.RateLimitSeconds != 2 {
		t.Errorf("expected RateLimitSeconds=2, got %d", cfg.RateLimitSeconds)
	}
}

func TestConfig_AllOverrides(t *testing.T) {
	t.Setenv("SERVER_PORT", "3000")
	t.Setenv("SEMANTIC_SERVICE_URL", "http://remote:9000")
	t.Setenv("GAME_DURATION_SECONDS", "120")
	t.Setenv("RATE_LIMIT_SECONDS", "3")

	cfg := Load()

	if cfg.ServerPort != "3000" {
		t.Errorf("ServerPort: expected 3000, got %s", cfg.ServerPort)
	}
	if cfg.SemanticURL != "http://remote:9000" {
		t.Errorf("SemanticURL: expected http://remote:9000, got %s", cfg.SemanticURL)
	}
	if cfg.GameDuration != 120 {
		t.Errorf("GameDuration: expected 120, got %d", cfg.GameDuration)
	}
	if cfg.RateLimitSeconds != 3 {
		t.Errorf("RateLimitSeconds: expected 3, got %d", cfg.RateLimitSeconds)
	}
}
