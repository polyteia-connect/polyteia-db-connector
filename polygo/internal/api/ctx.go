package api

import (
	"context"
	"fmt"

	"resty.dev/v3"
)

type ctxKey int

const (
	requestCtxKey ctxKey = iota
)

type requestCtx struct {
	client *resty.Client
}

// getRequestCtx returns the request context.
func getRequestCtx(ctx context.Context) *requestCtx {
	return ctx.Value(requestCtxKey).(*requestCtx)
}

// WithRequestCtx sets the Personal Access Key and base URL in the context.
func WithRequestCtx(ctx context.Context, pak string, baseURL string) context.Context {
	client := resty.New().
		SetBaseURL(baseURL).
		SetAuthToken(pak).
		AddResponseMiddleware(func(client *resty.Client, resp *resty.Response) error {
			if resp.StatusCode() != 200 {
				return fmt.Errorf("unexpected response status code: %d, body: %s", resp.StatusCode(), resp.Bytes())
			}

			return nil
		})

	return context.WithValue(ctx, requestCtxKey, &requestCtx{
		client: client,
	})
}
