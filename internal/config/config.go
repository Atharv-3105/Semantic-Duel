package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	ServerPort       string
	GameDuration     int
	RateLimitSeconds int

	// Pre-computed embeddings paths (default mode — no Python at runtime).
	VocabPath      string
	EmbeddingsPath string

	// SemanticURL is the legacy HTTP endpoint for the Python FastAPI service.
	// Leave empty (default) to use pre-computed local embeddings instead.
	// Set SEMANTIC_SERVICE_URL env var to switch back to the HTTP client.
	SemanticURL string
}

func Load() *Config {
	cfg := &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		GameDuration:     getEnvInt("GAME_DURATION_SECONDS", 60),
		RateLimitSeconds: getEnvInt("RATE_LIMIT_SECONDS", 1),
		VocabPath:        getEnv("VOCAB_PATH", "data/vocabulary.json"),
		EmbeddingsPath:   getEnv("EMBEDDINGS_PATH", "data/embeddings.npy"),
		SemanticURL:      getEnv("SEMANTIC_SERVICE_URL", ""),
	}

	log.Println("[CONFIG] Loaded")
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("[CONFIG] Invalid int env %s = %s", key, v)
		}
		return i
	}
	return fallback
}