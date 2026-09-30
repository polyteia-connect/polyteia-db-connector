package polygo

import (
	"context"

	"github.com/polyteia-connect/polyteia-db-connector/polygo/internal/api"
)

type dpaAcceptanceStatusRequest struct {
	SolutionID string `json:"solutionId"`
}

type DpaAcceptanceStatus struct {
	Acknowledged bool `json:"acknowledged"`
}

// GetDpaAcceptanceStatus returns whether the session's user has acknowledged the current data processing terms
// of the solution. Uploads to the solution's datasets are refused until they are acknowledged.
func (c *Client) GetDpaAcceptanceStatus(ctx context.Context, token string, solutionID string) (*DpaAcceptanceStatus, error) {
	return api.RPC[dpaAcceptanceStatusRequest, DpaAcceptanceStatus](c.apiCtx(ctx, token), "dpa/getMyAcceptanceStatus", dpaAcceptanceStatusRequest{SolutionID: solutionID})
}
