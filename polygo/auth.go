package polygo

import (
	"context"
	"time"

	"github.com/polyteia-connect/polyteia-db-connector/polygo/internal/api"
)

type exchangePersonalAccessKeyRequest struct {
	Key              string `json:"key"`
	OrganizationID   string `json:"organizationId,omitempty"`
	OrganizationSlug string `json:"organizationSlug,omitempty"`
}

// Session is a short-lived session obtained by exchanging a personal access key.
type Session struct {
	Token          string    `json:"token"`
	ExpiresAt      time.Time `json:"expiresAt"`
	OrganizationID string    `json:"organizationId"`
	Scope          string    `json:"scope"`
}

// ExchangePersonalAccessKey exchanges the client's personal access key for a session token
// scoped to the client's organization.
func (c *Client) ExchangePersonalAccessKey(ctx context.Context) (*Session, error) {
	return api.RPC[exchangePersonalAccessKeyRequest, Session](c.apiCtx(ctx, ""), "auth/personalAccessKey/exchange", exchangePersonalAccessKeyRequest{
		Key:              c.pak,
		OrganizationID:   c.organization.ID,
		OrganizationSlug: c.organization.Slug,
	})
}
