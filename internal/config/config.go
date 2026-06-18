package config

import (
	"errors"
	"net"
	"os"
	"strings"
	"time"
)

// Config contains all process-level settings loaded from environment variables.
type Config struct {
	Addr            string
	AuthToken       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// Load reads configuration from the environment and enforces safe binding rules.
func Load() (Config, error) {
	cfg := Config{
		Addr:            getEnv("VPS_INSPECTOR_ADDR", "127.0.0.1:8719"),
		AuthToken:       strings.TrimSpace(os.Getenv("VPS_INSPECTOR_AUTH_TOKEN")),
		ReadTimeout:     getDuration("VPS_INSPECTOR_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getDuration("VPS_INSPECTOR_WRITE_TIMEOUT", 30*time.Second),
		ShutdownTimeout: getDuration("VPS_INSPECTOR_SHUTDOWN_TIMEOUT", 10*time.Second),
	}

	host, _, err := net.SplitHostPort(cfg.Addr)
	if err != nil {
		return Config{}, err
	}
	if isPublicBind(host) && cfg.AuthToken == "" {
		return Config{}, errors.New("VPS_INSPECTOR_AUTH_TOKEN is required when binding to a public address")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func isPublicBind(host string) bool {
	if host == "" || host == "0.0.0.0" || host == "::" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return true
	}
	return !ip.IsLoopback()
}
