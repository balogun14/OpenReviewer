package config

import (
	"bufio"
	"os"
	"strings"
)

type DotEnvStatus struct {
	Path   string
	Loaded bool
}

func loadDotEnv(path string) DotEnvStatus {
	status := DotEnvStatus{Path: path}

	file, err := os.Open(path)
	if err != nil {
		return status
	}
	defer file.Close()
	status.Loaded = true

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		_ = os.Setenv(key, value)
	}

	return status
}
