package config

import "os"

type Config struct {
	Addr                string
	GitHubWebhookSecret string
}

func FromEnv() Config {
	addr := os.Getenv("OPENREVIEW_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{
		Addr:                addr,
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
	}
}
