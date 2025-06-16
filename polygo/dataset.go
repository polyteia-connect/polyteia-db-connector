package polygo

import (
	"context"

	"github.com/polyteia-connect/polyteia-db-connector/polygo/internal/api"
)

type DatasetUploadTokenRequest struct {
	ID          string `json:"id,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type DatasetUploadTokenResponse struct {
	Token string `json:"token"`
}

func (c *Client) GenerateDatasetUploadToken(ctx context.Context, request DatasetUploadTokenRequest) (*DatasetUploadTokenResponse, error) {
	return api.Command[DatasetUploadTokenRequest, DatasetUploadTokenResponse](c.apiCtx(ctx), "generate_dataset_upload_token", request)
}

func (c *Client) UploadDataset(ctx context.Context, token string, filePath string) error {
	return api.Upload(c.apiCtx(ctx), token, filePath)
}
