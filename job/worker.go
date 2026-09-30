package job

import (
	"context"
	"fmt"
	"time"

	"github.com/polyteia-connect/polyteia-db-connector/polygo"
)

type SourceDatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Type     string
}

type WorkerConfig struct {
	DatasetID      string
	SQLQuery       string
	SourceDatabase SourceDatabaseConfig
	// IngestPollInterval is how often the ingest status is checked after an upload.
	IngestPollInterval time.Duration
	// IngestTimeout is how long to wait for the ingest to finish after an upload.
	IngestTimeout time.Duration
}

type Worker struct {
	apiClient *polygo.Client
	workflow  *Workflow
	wConfig   WorkerConfig
	id        string
}

func NewWorker(apiClient *polygo.Client, wConfig WorkerConfig) (*Worker, error) {
	workflow := NewWorkflow(apiClient, wConfig)

	return &Worker{
		apiClient: apiClient,
		workflow:  workflow,
		wConfig:   wConfig,
		id:        fmt.Sprintf("%d", time.Now().Unix()),
	}, nil
}

func (w *Worker) ID() string {
	return w.id
}

func (w *Worker) Run(ctx context.Context) error {
	return w.workflow.Execute(ctx)
}

func (w *Worker) Close() error {
	// Workflow handles its own cleanup via defer in Execute()
	// This method is kept for backward compatibility
	return nil
}
