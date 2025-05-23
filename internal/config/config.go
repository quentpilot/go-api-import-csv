package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type EnvConfig interface {
	Load()
}

func LoadEnv() {
	_ = godotenv.Load()
}

// Get retrieves the value of the environment variable named by the key.
func Get(key string, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	return val
}

// GetBool retrieves the value of the environment variable named by the key.
func GetBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		log.Printf("Bad boolean value for key %s: %v", key, err)
		return fallback
	}
	return b
}

// GetInt retrieves the value of the environment variable named by the key.
func GetInt(key string, fallback int64) int64 {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Printf("Bad integer value for key %s: %v", key, err)
		return fallback
	}
	return i
}

// GetFloat retrieves the value of the environment variable named by the key.
func GetFloat(key string, fallback float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("Bad float value for key %s: %v", key, err)
		return fallback
	}
	return f
}
