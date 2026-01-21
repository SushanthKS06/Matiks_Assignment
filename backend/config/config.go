package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	InitialUsers   int
	MinRating      int
	MaxRating      int
	UpdateInterval int // milliseconds between simulated updates
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	initialUsers := 10000
	if val := os.Getenv("INITIAL_USERS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			initialUsers = parsed
		}
	}

	updateInterval := 100 // 100ms default
	if val := os.Getenv("UPDATE_INTERVAL"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			updateInterval = parsed
		}
	}

	return &Config{
		Port:           port,
		InitialUsers:   initialUsers,
		MinRating:      100,
		MaxRating:      5000,
		UpdateInterval: updateInterval,
	}
}
