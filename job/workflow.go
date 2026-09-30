package job

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/polyteia-connect/polyteia-db-connector/polygo"
)

// Workflow represents a job execution workflow with multiple steps
type Workflow struct {
	apiClient *polygo.Client
	config    WorkerConfig
	db        *sql.DB
	csvFile   string // Track CSV file for cleanup
}

// NewWorkflow creates a new workflow instance
func NewWorkflow(apiClient *polygo.Client, config WorkerConfig) *Workflow {
	return &Workflow{
		apiClient: apiClient,
		config:    config,
	}
}

// Execute runs the complete workflow: Connect → Execute → Upload
func (wf *Workflow) Execute(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting workflow execution")

	// Step 1: Connect to database
	if err := wf.stepConnect(ctx); err != nil {
		return fmt.Errorf("workflow step 'connect' failed: %w", err)
	}
	defer wf.cleanup()

	// Step 2: Execute query and export to CSV
	csvFile, err := wf.stepExecute(ctx)
	if err != nil {
		return fmt.Errorf("workflow step 'execute' failed: %w", err)
	}
	wf.csvFile = csvFile
	defer wf.cleanupFiles(ctx)

	// Step 3: Upload to Polyteia
	if err := wf.stepUpload(ctx, csvFile); err != nil {
		return fmt.Errorf("workflow step 'upload' failed: %w", err)
	}

	slog.InfoContext(ctx, "Workflow completed successfully")
	return nil
}

// stepConnect establishes a connection to the source database
func (wf *Workflow) stepConnect(ctx context.Context) error {
	slog.InfoContext(ctx, "Connecting to database", "type", wf.config.SourceDatabase.Type, "host", wf.config.SourceDatabase.Host)

	// Retry connection with exponential backoff
	var err error
	maxRetries := 3
	backoff := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			slog.WarnContext(ctx, "Retrying database connection", "attempt", attempt+1, "backoff", backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2 // Exponential backoff
		}

		wf.db, err = OpenDatabase(ctx, wf.config.SourceDatabase)
		if err == nil {
			slog.InfoContext(ctx, "Database connection established successfully")
			return nil
		}

		slog.WarnContext(ctx, "Database connection attempt failed", "attempt", attempt+1, "error", err)
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

// stepExecute runs the SQL query and exports results to CSV
func (wf *Workflow) stepExecute(ctx context.Context) (string, error) {
	slog.InfoContext(ctx, "Executing query and exporting to CSV")

	csvFile, err := ExecuteQueryToCSV(ctx, wf.db, wf.config.SQLQuery)
	if err != nil {
		return "", err
	}

	slog.InfoContext(ctx, "Query executed and exported to CSV", "file", csvFile)
	return csvFile, nil
}

// stepUpload uploads the CSV file to Polyteia and waits for the ingest to finish
func (wf *Workflow) stepUpload(ctx context.Context, csvFile string) error {
	slog.DebugContext(ctx, "Exchanging personal access key for a session")

	session, err := wf.apiClient.ExchangePersonalAccessKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to exchange personal access key: %w", err)
	}

	slog.InfoContext(ctx, "Uploading file to dataset", "dataset_id", wf.config.DatasetID, "file", csvFile)

	upload, err := wf.apiClient.UploadDataset(ctx, session.Token, wf.config.DatasetID, "text/csv", csvFile)
	if err != nil {
		return fmt.Errorf("failed to upload dataset: %w", err)
	}

	slog.InfoContext(ctx, "File uploaded successfully, waiting for ingest", "ingest_id", upload.IngestID, "size", upload.Size)

	return wf.waitForIngest(ctx, session, upload.IngestID)
}

// waitForIngest polls the ingest status until it completes, fails or the ingest timeout is reached
func (wf *Workflow) waitForIngest(ctx context.Context, session *polygo.Session, ingestID string) error {
	ctx, cancel := context.WithTimeout(ctx, wf.config.IngestTimeout)
	defer cancel()

	ticker := time.NewTicker(wf.config.IngestPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("ingest %s did not finish within %s: %w", ingestID, wf.config.IngestTimeout, ctx.Err())
		case <-ticker.C:
		}

		// Sessions are short-lived, renew it before it expires during a long ingest
		if time.Until(session.ExpiresAt) < time.Minute {
			renewed, err := wf.apiClient.ExchangePersonalAccessKey(ctx)
			if err != nil {
				return fmt.Errorf("failed to renew session: %w", err)
			}
			session = renewed
		}

		status, err := wf.apiClient.GetDatasetIngestStatus(ctx, session.Token, polygo.DatasetIngestStatusRequest{
			ID:       wf.config.DatasetID,
			IngestID: ingestID,
		})
		if err != nil {
			// Client errors will not resolve by retrying, anything else may be transient
			var apiErr *polygo.Error
			if errors.As(err, &apiErr) && apiErr.Status >= 400 && apiErr.Status < 500 && apiErr.Status != http.StatusTooManyRequests {
				return fmt.Errorf("failed to get ingest status: %w", err)
			}

			slog.WarnContext(ctx, "Failed to get ingest status, retrying", "ingest_id", ingestID, "error", err)
			continue
		}

		slog.DebugContext(ctx, "Ingest status", "ingest_id", ingestID, "status", status.Status)

		switch status.Status {
		case polygo.IngestStatusCompleted:
			slog.InfoContext(ctx, "Ingest completed successfully", "ingest_id", ingestID)
			return nil
		case polygo.IngestStatusFailed:
			return fmt.Errorf("ingest %s failed: %s", ingestID, status.ErrorMessage)
		}
	}
}

// cleanup closes database connections and cleans up resources
func (wf *Workflow) cleanup() {
	if wf.db != nil {
		if err := wf.db.Close(); err != nil {
			slog.WarnContext(context.Background(), "Error closing database connection", "error", err)
		}
	}
}

// cleanupFiles removes temporary files created during workflow execution
func (wf *Workflow) cleanupFiles(ctx context.Context) {
	if wf.csvFile != "" {
		if err := os.Remove(wf.csvFile); err != nil {
			slog.WarnContext(ctx, "Error removing temporary CSV file", "file", wf.csvFile, "error", err)
		} else {
			slog.DebugContext(ctx, "Cleaned up temporary CSV file", "file", wf.csvFile)
		}
	}
}
