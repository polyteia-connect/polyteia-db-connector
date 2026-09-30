package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	PersonalAccessToken string        `env:"PERSONAL_ACCESS_TOKEN" env-required:"true"`
	LogLevel            string        `env:"LOG_LEVEL" env-default:"info"`
	LogFormat           string        `env:"LOG_FORMAT" env-default:"text"`
	HealthCheckPort     string        `env:"HEALTH_CHECK_PORT" env-default:"8080"`
	BaseURL             string        `env:"POLYTEIA_BASE_URL" env-default:"https://web.polyteia.de"`
	OrganizationID      string        `env:"POLYTEIA_ORGANIZATION_ID"`              // ID of the organization the dataset belongs to. Either this or POLYTEIA_ORGANIZATION_SLUG is required
	OrganizationSlug    string        `env:"POLYTEIA_ORGANIZATION_SLUG"`            // Slug of the organization the dataset belongs to. Either this or POLYTEIA_ORGANIZATION_ID is required
	DatasetID           string        `env:"DATASET_ID" env-required:"true"`        // ID of the dataset to push the sql results to
	CronSchedule        string        `env:"CRON_SCHEDULE" env-default:"0 0 * * *"` // every day at midnight
	IngestPollInterval  time.Duration `env:"INGEST_POLL_INTERVAL" env-default:"5s"` // How often to check the ingest status after an upload
	IngestTimeout       time.Duration `env:"INGEST_TIMEOUT" env-default:"30m"`      // How long to wait for the ingest to finish after an upload
	SourceDatabase      struct {
		Host     string `env:"SOURCE_DATABASE_HOST" env-required:"true"`      // Host of the source database. E.g. localhost
		Port     string `env:"SOURCE_DATABASE_PORT" env-required:"true"`      // Port of the source database. E.g. 5432
		User     string `env:"SOURCE_DATABASE_USER" env-required:"true"`      // User of the source database. E.g. user
		Password string `env:"SOURCE_DATABASE_PASSWORD"`                      // Password of the source database. E.g. password
		Name     string `env:"SOURCE_DATABASE_NAME" env-required:"true"`      // Name of the source database. E.g. database
		Type     string `env:"SOURCE_DATABASE_TYPE" env-required:"true"`      // Type of the source database. Supported: postgres, mysql, sqlserver
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

	if err := config.validate(); err != nil {
		panic(fmt.Errorf("invalid config: %w", err))
	}

	return config
}

func (c Config) validate() error {
	if (c.OrganizationID == "") == (c.OrganizationSlug == "") {
		return fmt.Errorf("exactly one of POLYTEIA_ORGANIZATION_ID or POLYTEIA_ORGANIZATION_SLUG must be set")
	}

	if c.IngestPollInterval <= 0 {
		return fmt.Errorf("INGEST_POLL_INTERVAL must be positive")
	}

	if c.IngestTimeout <= 0 {
		return fmt.Errorf("INGEST_TIMEOUT must be positive")
	}

	return nil
}
