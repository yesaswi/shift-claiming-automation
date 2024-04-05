package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              int
	ProjectID         string
	DatabaseID        string
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8080
	}

	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		projectID = "autoclaimer-42"
	}
	databaseID := os.Getenv("DATABASE_ID")
	if databaseID == "" {
		databaseID = "autoclaimer-42-db"
	}
	return &Config{
		Port:              port,
		ProjectID:         projectID,
		DatabaseID:        databaseID,
	}, nil
}
