package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type EnvConfig interface {
	Load() // Load current config type with env vars or fallback values
}

// Config holds the application modules parameters when initializing the application.
type AppConfig struct {
	LoggerName string       // File name for the current logger (default: "root")
	Logger     LoggerConfig // Logger configuration
	Http       HttpConfig   // HTTP server configuration
	Amqp       ApmqConfig   // AMQP server configuration
	Db         DbConfig     // Database server configuration
	UseDb      bool         // Whether to open a database connection (default: false)
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Failed to load .env file", err)
	}
}

func ReloadEnv() error {
	return godotenv.Overload(".env")
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
