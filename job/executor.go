package job

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"time"
)

// ExecuteQueryToCSV executes a SQL query against the database and writes
// the results to a CSV file. Returns the path to the temporary CSV file.
// The caller is responsible for cleaning up the temporary file.
func ExecuteQueryToCSV(ctx context.Context, db *sql.DB, query string) (string, error) {
	// Trim surrounding quotes if present (common when reading from .env files)
	query = strings.TrimSpace(query)
	if len(query) >= 2 {
		if (query[0] == '"' && query[len(query)-1] == '"') ||
			(query[0] == '\'' && query[len(query)-1] == '\'') {
			query = query[1 : len(query)-1]
		}
	}

	if query == "" {
		return "", fmt.Errorf("query is empty after trimming")
	}

	// Create temporary CSV file
	tempFile, err := os.CreateTemp(os.TempDir(), "query_result_*.csv")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Ensure file is closed if we return early
	fileClosed := false
	defer func() {
		if !fileClosed {
			tempFile.Close() //nolint:errcheck
		}
	}()

	// Execute the query
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to get columns: %w", err)
	}

	if len(columns) == 0 {
		return "", fmt.Errorf("query returned no columns")
	}

	// Create CSV writer
	writer := csv.NewWriter(tempFile)

	// Write header
	if err := writer.Write(columns); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Prepare scan destination
	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	const flushInterval = 1000 // Flush every 1000 rows to prevent excessive memory usage
	// Write rows
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return "", fmt.Errorf("failed to scan row %d: %w", rowCount+1, err)
		}

		// Convert values to strings for CSV
		record := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				record[i] = ""
			} else {
				slog.DebugContext(ctx, "Value", "value", val, "type", reflect.TypeOf(val))
				// Handle different types appropriately
				switch v := val.(type) {
				case []byte:
					record[i] = string(v)
				case time.Time:
					// Format time in RFC3339 format (ISO8601) without duplicate timezone
					record[i] = v.Format(time.RFC3339Nano)
				case io.Reader:
					// For blob types, read and convert to string
					data, readErr := io.ReadAll(v)
					if readErr != nil {
						slog.WarnContext(ctx, "Failed to read blob data", "error", readErr, "column", columns[i])
						record[i] = ""
					} else {
						record[i] = string(data)
					}
				default:
					record[i] = fmt.Sprintf("%v", val)
				}
			}
		}

		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write CSV row %d: %w", rowCount+1, err)
		}
		rowCount++

		// Periodically flush to disk to prevent excessive memory usage
		if rowCount%flushInterval == 0 {
			writer.Flush()
			if err := writer.Error(); err != nil {
				return "", fmt.Errorf("failed to flush CSV writer at row %d: %w", rowCount, err)
			}
			// Also sync periodically to ensure data is written to disk
			if err := tempFile.Sync(); err != nil {
				return "", fmt.Errorf("failed to sync temporary file at row %d: %w", rowCount, err)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating rows: %w", err)
	}

	// Flush the CSV writer to ensure all data is written to the file
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	// Sync the file to ensure data is written to disk
	if err := tempFile.Sync(); err != nil {
		return "", fmt.Errorf("failed to sync temporary file: %w", err)
	}

	// Close the file before returning
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temporary file: %w", err)
	}
	fileClosed = true

	slog.InfoContext(ctx, "Query executed successfully", "rows", rowCount, "file", tempFile.Name())

	return tempFile.Name(), nil
}
