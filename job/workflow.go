package job

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
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

// stepUpload uploads the CSV file to Polyteia
func (wf *Workflow) stepUpload(ctx context.Context, csvFile string) error {
	slog.DebugContext(ctx, "Generating dataset upload token")

	// Generate dataset upload token
	uploadToken, err := wf.apiClient.GenerateDatasetUploadToken(ctx, polygo.DatasetUploadTokenRequest{
		ID:          wf.config.DatasetID,
		ContentType: "text/csv",
	})
	if err != nil {
		return fmt.Errorf("failed to generate upload token: %w", err)
	}

	slog.InfoContext(ctx, "Uploading file to dataset", "dataset_id", wf.config.DatasetID, "file", csvFile)

	// Upload file to dataset
	if err := wf.apiClient.UploadDataset(ctx, uploadToken.Token, csvFile); err != nil {
		return fmt.Errorf("failed to upload dataset: %w", err)
	}

	slog.InfoContext(ctx, "File uploaded successfully")
	return nil
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
