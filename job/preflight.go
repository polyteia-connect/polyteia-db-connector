package job

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/polyteia-connect/polyteia-db-connector/polygo"
)

// preflightProbeIngestID is an ingest ID that never exists. It is used to check upload permission
// without uploading anything, see checkUploadPermission.
const preflightProbeIngestID = "dsig_preflight_probe"

// Preflight checks that the source database and the Polyteia API are reachable with the configured
// credentials and that everything needed for an upload is in place. Every check is logged whether it
// passes or fails. All checks run, and an error listing the failed ones is returned at the end.
func Preflight(ctx context.Context, apiClient *polygo.Client, cfg WorkerConfig, baseURL string) error {
	slog.InfoContext(ctx, "Running startup checks")

	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	var failed []string
	check := func(name string, err error) bool {
		if err != nil {
			slog.ErrorContext(ctx, "Startup check failed", "check", name, "error", err)
			failed = append(failed, name)
			return false
		}
		return true
	}

	if check("source database", checkSourceDatabase(ctx, cfg)) {
		check("source query", checkSourceQuery(ctx, cfg))
	} else {
		slog.WarnContext(ctx, "Skipping source query check, it needs a database connection")
	}

	slog.InfoContext(ctx, "Connecting to Polyteia", "base_url", baseURL)
	session, err := apiClient.ExchangePersonalAccessKey(ctx)
	if check("personal access key", describe(err, "the personal access key could not be exchanged for a session")) {
		slog.InfoContext(ctx, "Startup check passed", "check", "personal access key",
			"base_url", baseURL, "organization_id", session.OrganizationID, "scope", session.Scope, "session_expires_at", session.ExpiresAt)

		dataset, err := apiClient.GetDataset(ctx, session.Token, cfg.DatasetID)
		if check("dataset access", describe(err, fmt.Sprintf("dataset %s could not be read, check that DATASET_ID exists in the organization", cfg.DatasetID))) {
			slog.InfoContext(ctx, "Startup check passed", "check", "dataset access",
				"dataset_id", dataset.ID, "dataset_name", dataset.Name, "solution_id", dataset.SolutionID)

			check("upload permission", checkUploadPermission(ctx, apiClient, session.Token, cfg.DatasetID))
			check("data processing terms", checkDpa(ctx, apiClient, session.Token, dataset.SolutionID))
		} else {
			slog.WarnContext(ctx, "Skipping upload permission and data processing terms checks, they need dataset access")
		}
	} else {
		slog.WarnContext(ctx, "Skipping dataset checks, they need a valid session")
	}

	if len(failed) > 0 {
		return fmt.Errorf("startup checks failed: %v", failed)
	}

	slog.InfoContext(ctx, "All startup checks passed")
	return nil
}

func checkSourceDatabase(ctx context.Context, cfg WorkerConfig) error {
	db := cfg.SourceDatabase
	slog.InfoContext(ctx, "Connecting to source database",
		"type", db.Type, "host", db.Host, "port", db.Port, "database", db.Name, "user", db.User)

	conn, err := OpenDatabase(ctx, db)
	if err != nil {
		return err
	}
	defer conn.Close() //nolint:errcheck

	slog.InfoContext(ctx, "Startup check passed", "check", "source database",
		"type", db.Type, "host", db.Host, "port", db.Port, "database", db.Name, "user", db.User)
	return nil
}

// checkSourceQuery prepares the query, which makes the server parse it and catches syntax errors and
// missing tables without running it. The SQL Server driver does not contact the server on prepare, so
// it is skipped there.
func checkSourceQuery(ctx context.Context, cfg WorkerConfig) error {
	switch cfg.SourceDatabase.Type {
	case "mssql", "sqlserver":
		slog.InfoContext(ctx, "Skipping source query check, not supported for this database type", "type", cfg.SourceDatabase.Type)
		return nil
	}

	conn, err := OpenDatabase(ctx, cfg.SourceDatabase)
	if err != nil {
		return err
	}
	defer conn.Close() //nolint:errcheck

	stmt, err := conn.PrepareContext(ctx, cfg.SQLQuery)
	if err != nil {
		return fmt.Errorf("SQL query is invalid, check SOURCE_DATABASE_SQL_QUERY: %w", err)
	}
	_ = stmt.Close()

	slog.InfoContext(ctx, "Startup check passed", "check", "source query")
	return nil
}

// checkUploadPermission relies on the ingest status endpoint requiring the same edit permission as
// uploads, and checking it before looking up the ingest: an unknown ingest ID answers NOT_FOUND when
// the user may edit the dataset and FORBIDDEN when not.
func checkUploadPermission(ctx context.Context, apiClient *polygo.Client, token string, datasetID string) error {
	_, err := apiClient.GetDatasetIngestStatus(ctx, token, polygo.DatasetIngestStatusRequest{
		ID:       datasetID,
		IngestID: preflightProbeIngestID,
	})

	var apiErr *polygo.Error
	switch {
	case err == nil, errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound:
		slog.InfoContext(ctx, "Startup check passed", "check", "upload permission", "dataset_id", datasetID)
		return nil
	case errors.As(err, &apiErr) && apiErr.Status == http.StatusForbidden:
		return fmt.Errorf("the user of the personal access key is not allowed to edit dataset %s", datasetID)
	default:
		return describe(err, "upload permission could not be checked")
	}
}

func checkDpa(ctx context.Context, apiClient *polygo.Client, token string, solutionID string) error {
	status, err := apiClient.GetDpaAcceptanceStatus(ctx, token, solutionID)
	if err != nil {
		return describe(err, "data processing terms status could not be checked")
	}

	if !status.Acknowledged {
		return fmt.Errorf("the user of the personal access key has not acknowledged the data processing terms of solution %s, uploads will be refused until they are acknowledged in the Polyteia app", solutionID)
	}

	slog.InfoContext(ctx, "Startup check passed", "check", "data processing terms", "solution_id", solutionID)
	return nil
}

// describe prefixes an error with what failed, adding a hint for the common API statuses.
func describe(err error, what string) error {
	if err == nil {
		return nil
	}

	var apiErr *polygo.Error
	if errors.As(err, &apiErr) {
		switch apiErr.Status {
		case http.StatusUnauthorized:
			return fmt.Errorf("%s, check PERSONAL_ACCESS_TOKEN and that the key is not revoked or expired: %w", what, err)
		case http.StatusForbidden:
			return fmt.Errorf("%s, the user of the personal access key lacks access (licensed membership, permissions, or personal access keys not allowed for the organization): %w", what, err)
		case http.StatusNotFound:
			return fmt.Errorf("%s, check DATASET_ID and POLYTEIA_ORGANIZATION_ID/POLYTEIA_ORGANIZATION_SLUG: %w", what, err)
		}
	}

	return fmt.Errorf("%s, check POLYTEIA_BASE_URL and network access: %w", what, err)
}
