// Package config loads and validates the application's settings.
//
// Settings come from environment variables, are read once at startup, and are
// passed onward as a Config value. Nothing deeper in the application reads the
// environment on its own.
//
// Startup calls LoadFile to pull a local .env into the environment, then Load to
// read and validate it.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds every setting the application needs at startup.
type Config struct {
	Env         string // "development" or "production"
	Port        int    // port the HTTP server listens on
	DatabaseURL string // PostgreSQL connection string
}

// LoadFile reads `key=value` lines from an environment file, usually ".env",
// and puts them into the process environment. Call it before Load.
//
// Variables that are already present in the real environment are left alone, so
// an exported value always beats the file. That ordering is what makes the file
// a convenience for local development rather than a way to override a
// deliberately configured server.
//
// A missing file is not an error: in production the configuration comes from
// the real environment and no .env file exists.
func LoadFile(path string) error {
	err := godotenv.Load(path)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("read %s: %w", path, err)
}

// Load reads the configuration from the environment, applying a default for
// each optional setting.
//
// It returns an error if a value is present but unusable, so that a
// misconfigured application fails immediately at startup rather than on some
// later request.
func Load() (Config, error) {
	env := getenv("APP_ENV", "development")
	if env != "development" && env != "production" {
		return Config{}, fmt.Errorf("APP_ENV must be %q or %q, got %q", "development", "production", env)
	}

	portText := getenv("PORT", "8080")
	port, err := strconv.Atoi(portText)
	if err != nil {
		return Config{}, fmt.Errorf("PORT must be a number, got %q", portText)
	}
	if port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be between 1 and 65535, got %d", port)
	}

	// Required, with no sensible default: guessing a connection string would only
	// turn a missing setting into a confusing connection error.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required (see .env.example)")
	}

	return Config{
		Env:         env,
		Port:        port,
		DatabaseURL: databaseURL,
	}, nil
}

// IsDevelopment reports whether the application is running in development mode.
func (c Config) IsDevelopment() bool {
	return c.Env == "development"
}

// getenv returns the value of the environment variable named by key, or
// fallback when that variable is unset or empty.
func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
