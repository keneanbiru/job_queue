package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

// findProjectRoot walks up from the current directory to find the project root
func findProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	dir := wd
	for {
		// Check for go.mod to identify project root
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return wd
}

func LoadConfig() Config {
	// Try to load .env from project root
	projectRoot := findProjectRoot()
	envPath := filepath.Join(projectRoot, ".env")

	// Try loading from project root first
	if err := godotenv.Load(envPath); err != nil {
		// Fallback: try current directory
		_ = godotenv.Load()
		// Also try cmd/api folder
		_ = godotenv.Load(filepath.Join(projectRoot, "cmd", "api", ".env"))
	}

	cfg := Config{
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
	}

	if cfg.DBUser == "" {
		log.Fatal("❌ Missing DB config in .env file")
	}

	return cfg
}
