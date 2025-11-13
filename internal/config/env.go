package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() { _ = godotenv.Load(".env") }

func GetEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func MustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env: %s", key)
	}
	return v
}
