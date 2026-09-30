package polygo

import (
	"context"

	"github.com/polyteia-connect/polyteia-db-connector/polygo/internal/api"
)

// Error is an error returned by the Polyteia API.
type Error = api.Error

type Client struct {
	pak, baseURL string
	organization Organization
}

// Organization identifies the organization a personal access key is exchanged for.
// Exactly one of ID or Slug should be set.
type Organization struct {
	ID   string
	Slug string
}

func NewClient(pak string, baseURL string, organization Organization) *Client {
	return &Client{
		pak:          pak,
		baseURL:      baseURL,
		organization: organization,
	}
}

func (c *Client) apiCtx(ctx context.Context, token string) context.Context {
	return api.WithRequestCtx(ctx, c.baseURL, token)
}
