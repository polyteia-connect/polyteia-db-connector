package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/polyteia-connect/polyteia-db-connector/config"
	"github.com/polyteia-connect/polyteia-db-connector/job"
	"github.com/polyteia-connect/polyteia-db-connector/polygo"
	"github.com/robfig/cron/v3"
	slogctx "github.com/veqryn/slog-context"
)

const (
	maxJobRetries = 3
	retryInterval = 1 * time.Minute
)

func main() {
	ctx := context.Background()
	c := config.Auto()
	setupLogger(c)

	crn := cron.New(cron.WithChain(
		cron.Recover(crnLogger{}),
	))
	defer crn.Stop()

	client := polygo.NewClient(c.PersonalAccessToken, c.BaseURL)

	_, err := crn.AddFunc(c.CronSchedule, func() {
		worker, err := newWorker(ctx, client, c)
		if err != nil {
			slog.ErrorContext(ctx, "Error creating new worker", "error", err)
			return
		}
		defer worker.Close() //nolint:errcheck

		ctx = slogctx.Append(ctx, "job_id", worker.ID())

		if err := doWithRetry(ctx, func() error {
			return worker.Run(ctx)
		}); err != nil {
			slog.ErrorContext(ctx, "Failed to run job after all retries", "error", err)
		}
	})

	if err != nil {
		slog.ErrorContext(ctx, "Error creating cron job", "error", err)
		os.Exit(1)
	}

	go startHealthCheck(ctx, c.HealthCheckPort, crn)

	slog.InfoContext(ctx, "Starting cron scheduler", slog.String("schedule", c.CronSchedule))

	crn.Run()
}

func startHealthCheck(ctx context.Context, port string, crn *cron.Cron) {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if crn.Entries() == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	slog.InfoContext(ctx, "Starting health check server", slog.String("port", port))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.ErrorContext(ctx, "Error starting health check server", slog.Any("error", err))
	}
}

func newWorker(ctx context.Context, client *polygo.Client, c config.Config) (*job.Worker, error) {
	slog.InfoContext(ctx, "Starting worker...")
	return job.NewWorker(ctx, client, job.WorkerConfig{
		DatasetID: c.DatasetID,
		SQLQuery:  c.SourceDatabase.SQLQuery,
		SourceDatabase: job.SourceDatabaseConfig{
			Host:     c.SourceDatabase.Host,
			Port:     c.SourceDatabase.Port,
			User:     c.SourceDatabase.User,
			Password: c.SourceDatabase.Password,
			Name:     c.SourceDatabase.Name,
			Type:     c.SourceDatabase.Type,
		},
	})
}

func doWithRetry(ctx context.Context, f func() error) error {
	var lastErr error
	for i := 0; i < maxJobRetries; i++ {
		if err := f(); err != nil {
			slog.ErrorContext(ctx, "Error in function execution", "error", err)
			slog.WarnContext(ctx, "Retrying function execution", "retry", i+1, "interval", retryInterval.String())
			lastErr = err
			<-time.After(retryInterval)
		} else {
			return nil
		}
	}

	return lastErr
}

type crnLogger struct{}

func (l crnLogger) Error(err error, msg string, keysAndValues ...any) {
	if keysAndValues == nil {
		keysAndValues = make([]any, 0)
	}

	keysAndValues = append(keysAndValues, "error", err)

	slog.Error(msg, keysAndValues...)
}

func (crnLogger) Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func setupLogger(c config.Config) {
	var handler slog.Handler

	handlerOptions := &slog.HandlerOptions{}
	configLogLevel := strings.ToLower(c.LogLevel)

	switch configLogLevel {
	case "debug":
		handlerOptions.Level = slog.LevelDebug
	case "info":
		handlerOptions.Level = slog.LevelInfo
	case "warn":
		handlerOptions.Level = slog.LevelWarn
	case "error":
		handlerOptions.Level = slog.LevelError
	default:
		handlerOptions.Level = slog.LevelInfo
	}

	if c.LogFormat == "json" {
		handler = slogctx.NewHandler(slog.NewJSONHandler(os.Stdout, handlerOptions), nil)
	} else {
		handler = slogctx.NewHandler(slog.NewTextHandler(os.Stdout, handlerOptions), nil)
	}

	slog.SetDefault(slog.New(handler))
}
