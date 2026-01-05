package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName         string
	SMTPHost        string
	SMTPPort        int
	SMTPUser        string
	SMTPPassword    string
	EmailTo         string
	HealthcheckURL  string
	MonitorInterval time.Duration
	StateFilePath   string
}

func Load() (*Config, error) {
	cfg := &Config{}

	// App Name
	cfg.AppName = getEnv("APP_NAME", "ORGMServer")

	// SMTP Configuration
	cfg.SMTPHost = getEnv("SMTP_HOST", "smtp.gmail.com")
	
	portStr := getEnv("SMTP_PORT", "587")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("SMTP_PORT debe ser un número válido: %w", err)
	}
	cfg.SMTPPort = port

	cfg.SMTPUser = getEnv("SMTP_USER", "")
	if cfg.SMTPUser == "" {
		return nil, fmt.Errorf("SMTP_USER es requerido")
	}

	cfg.SMTPPassword = getEnv("SMTP_PASSWORD", "")
	if cfg.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD es requerido")
	}

	cfg.EmailTo = getEnv("EMAIL_TO", "")
	if cfg.EmailTo == "" {
		return nil, fmt.Errorf("EMAIL_TO es requerido")
	}

	// Optional configurations
	cfg.HealthcheckURL = getEnv("HEALTHCHECK_URL", "")
	
	intervalStr := getEnv("MONITOR_INTERVAL", "60")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("MONITOR_INTERVAL debe ser un número válido: %w", err)
	}
	cfg.MonitorInterval = time.Duration(interval) * time.Second

	cfg.StateFilePath = getEnv("STATE_FILE_PATH", "/tmp/orgmserver_state.json")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

