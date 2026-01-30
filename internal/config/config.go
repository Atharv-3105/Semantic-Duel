package config 

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	ServerPort		string 
	SemanticURL		string 
	GameDuration	int
	RateLimitSeconds int 
}

func Load() *Config{
	cfg := &Config{
		ServerPort:		getEnv("SERVER_PORT", "8080"),
		SemanticURL: 	getEnv("SEMANTIC_SERVICE_URL", "http://localhost:8001"),
		GameDuration: 	getEnvInt("GAME_DURATION_SECONDS", 60),
		RateLimitSeconds: getEnvInt("RATE_LIMIT_SECONDS", 1),
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