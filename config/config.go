package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	PersonalAccessToken string `env:"PERSONAL_ACCESS_TOKEN" env-required:"true"`
	LogLevel            string `env:"LOG_LEVEL" env-default:"info"`
	LogFormat           string `env:"LOG_FORMAT" env-default:"text"`
	HealthCheckPort     string `env:"HEALTH_CHECK_PORT" env-default:"8080"`
	BaseURL             string `env:"POLYTEIA_BASE_URL" env-default:"https://app.polyteia.com"`
	DatasetID           string `env:"DATASET_ID" env-required:"true"`        // ID of the dataset to push the sql results to
	CronSchedule        string `env:"CRON_SCHEDULE" env-default:"0 0 * * *"` // every day at midnight
	SourceDatabase      struct {
		Host     string `env:"SOURCE_DATABASE_HOST" env-required:"true"`      // Host of the source database. E.g. localhost
		Port     string `env:"SOURCE_DATABASE_PORT" env-required:"true"`      // Port of the source database. E.g. 5432
		User     string `env:"SOURCE_DATABASE_USER" env-required:"true"`      // User of the source database. E.g. user
		Password string `env:"SOURCE_DATABASE_PASSWORD"`                      // Password of the source database. E.g. password
		Name     string `env:"SOURCE_DATABASE_NAME" env-required:"true"`      // Name of the source database. E.g. database
		Type     string `env:"SOURCE_DATABASE_TYPE" env-required:"true"`      // Type of the source database. Only mysql and postgres are supported for now
		SQLQuery string `env:"SOURCE_DATABASE_SQL_QUERY" env-required:"true"` // Query to run on source database
	}
}

func Auto() Config {
	var config Config

	// Try to read from .env file first
	if err := cleanenv.ReadConfig(".env", &config); err != nil {
		// If .env file is not found, read from environment variables
		if err := cleanenv.ReadEnv(&config); err != nil {
			panic(fmt.Errorf("failed to read config: %w", err))
		}
	}

	return config
}
