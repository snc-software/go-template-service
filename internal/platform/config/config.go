// Package config loads and validates process configuration from the environment.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Port     string
	LogLevel slog.Level
	Database Database
}

type Database struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

func (d Database) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		d.Host,
		d.Port,
		d.Name,
		d.User,
		d.Password,
		d.SSLMode,
	)
}

// LogValue keeps the password out of any log line that captures the config.
func (d Database) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", d.Host),
		slog.String("port", d.Port),
		slog.String("name", d.Name),
		slog.String("user", d.User),
		slog.String("sslmode", d.SSLMode),
	)
}

// Load reads configuration from the environment, reporting every missing
// required variable at once.
func Load() (Config, error) {
	var missing []string

	required := func(key string) string {
		value := os.Getenv(key)
		if value == "" {
			missing = append(missing, key)
		}

		return value
	}

	level, err := parseLevel(optional("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}

	config := Config{
		Port:     optional("PORT", "8080"),
		LogLevel: level,
		Database: Database{
			Host:     required("DB_HOST"),
			Port:     required("DB_PORT"),
			Name:     required("DB_NAME"),
			User:     required("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  optional("DB_SSLMODE", "require"),
		},
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return config, nil
}

func optional(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func parseLevel(raw string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		return 0, fmt.Errorf("invalid LOG_LEVEL %q: %w", raw, err)
	}

	return level, nil
}
