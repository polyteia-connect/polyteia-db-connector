package job

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/marcboeker/go-duckdb/v2"
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
}

type Worker struct {
	id        string
	apiClient *polygo.Client
	db        *sql.DB
	wConfig   WorkerConfig
}

func NewWorker(ctx context.Context, apiClient *polygo.Client, wConfig WorkerConfig) (*Worker, error) {
	db, err := openDuckDB(ctx, wConfig.SourceDatabase)
	if err != nil {
		return nil, err
	}

	return &Worker{
		apiClient: apiClient,
		db:        db,
		wConfig:   wConfig,
		id:        fmt.Sprintf("%d", time.Now().Unix()),
	}, nil
}

func (w *Worker) ID() string {
	return w.id
}

func (w *Worker) Run(ctx context.Context) error {
	slog.InfoContext(ctx, "Running job worker")
	tempFile, err := os.CreateTemp(os.TempDir(), "*.parquet")
	if err != nil {
		return err
	}
	defer tempFile.Close() //nolint:errcheck

	slog.InfoContext(ctx, "Executing query and saving results", "file", tempFile.Name())
	// COPY SQL query results to temp file
	dbQuery := fmt.Sprintf("COPY (%s) TO '%s' (FORMAT PARQUET);", w.wConfig.SQLQuery, tempFile.Name())
	_, err = w.db.ExecContext(ctx, dbQuery)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "Generating dataset upload token")
	// Generate dataset upload token
	uploadToken, err := w.apiClient.GenerateDatasetUploadToken(ctx, polygo.DatasetUploadTokenRequest{
		ID:          w.wConfig.DatasetID,
		ContentType: "application/vnd.apache.parquet",
	})
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Uploading file to dataset", "dataset_id", w.wConfig.DatasetID, "file", tempFile.Name())
	// Upload file to dataset
	err = w.apiClient.UploadDataset(ctx, uploadToken.Token, tempFile.Name())
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Job finished successfully")

	return nil
}

func (w *Worker) Close() error {
	return w.db.Close()
}

func openDuckDB(ctx context.Context, dbConfig SourceDatabaseConfig) (*sql.DB, error) {
	connector, err := duckdb.NewConnector(":memory:", func(ctx driver.ExecerContext) error {
		return nil
	})
	if err != nil {
		return nil, err
	}

	var dbConfigString string

	switch dbConfig.Type {
	case "postgres":
		dbConfigString = fmt.Sprintf("ATTACH 'dbname=%s user=%s host=%s port=%s password=%s' AS db (TYPE postgres, READ_ONLY)", dbConfig.Name, dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.Password)
	case "mysql":
		dbConfigString = fmt.Sprintf("ATTACH 'dbname=%s user=%s host=%s port=%s password=%s' AS db (TYPE mysql, READ_ONLY)", dbConfig.Name, dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.Password)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	db := sql.OpenDB(connector)
	bootQueries := []string{
		fmt.Sprintf("INSTALL %s;", dbConfig.Type),
		fmt.Sprintf("LOAD %s;", dbConfig.Type),
		dbConfigString,
	}

	for _, query := range bootQueries {
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}
