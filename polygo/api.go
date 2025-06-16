package polygo

import (
	"context"
	"github.com/polyteia-connect/polyteia-db-connector/polygo/internal/api"
)

type Client struct {
	pak, baseURL string
}

func NewClient(pak string, baseURL string) *Client {
	return &Client{
		pak:     pak,
		baseURL: baseURL,
	}
}

func (c *Client) apiCtx(ctx context.Context) context.Context {
	return api.WithRequestCtx(ctx, c.pak, c.baseURL)
}
