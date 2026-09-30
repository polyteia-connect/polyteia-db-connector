package polygo

import (
	"context"

	"github.com/polyteia-connect/polyteia-db-connector/polygo/internal/api"
)

type UploadDatasetResponse struct {
	OK          bool   `json:"ok"`
	IngestID    string `json:"ingestId"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

// UploadDataset uploads a file to the dataset, replacing its data. The returned ingest ID can be
// used with GetDatasetIngestStatus to follow the ingestion.
func (c *Client) UploadDataset(ctx context.Context, token string, datasetID string, contentType string, filePath string) (*UploadDatasetResponse, error) {
	return api.UploadDataset[UploadDatasetResponse](c.apiCtx(ctx, token), datasetID, contentType, filePath)
}

type IngestStatus string

const (
	IngestStatusPending   IngestStatus = "pending"
	IngestStatusRunning   IngestStatus = "running"
	IngestStatusCompleted IngestStatus = "completed"
	IngestStatusFailed    IngestStatus = "failed"
)

type DatasetIngestStatusRequest struct {
	ID       string `json:"id"`
	IngestID string `json:"ingestId"`
}

type DatasetIngestStatusResponse struct {
	Status       IngestStatus `json:"status"`
	ErrorMessage string       `json:"errorMessage,omitempty"`
}

func (c *Client) GetDatasetIngestStatus(ctx context.Context, token string, request DatasetIngestStatusRequest) (*DatasetIngestStatusResponse, error) {
	return api.RPC[DatasetIngestStatusRequest, DatasetIngestStatusResponse](c.apiCtx(ctx, token), "dataset/getDatasetIngestStatus", request)
}

type Dataset struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	SolutionID     string `json:"solutionId"`
	Name           string `json:"name"`
}

type getDatasetRequest struct {
	ID string `json:"id"`
}

// GetDataset returns the dataset if the session's user is allowed to view it.
func (c *Client) GetDataset(ctx context.Context, token string, datasetID string) (*Dataset, error) {
	return api.RPC[getDatasetRequest, Dataset](c.apiCtx(ctx, token), "dataset/getDatasetById", getDatasetRequest{ID: datasetID})
}
