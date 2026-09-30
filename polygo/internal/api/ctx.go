package api

import (
	"context"

	"resty.dev/v3"
)

type ctxKey int

const (
	requestCtxKey ctxKey = iota
)

type requestCtx struct {
	client  *resty.Client
	baseURL string
	token   string
}

// getRequestCtx returns the request context.
func getRequestCtx(ctx context.Context) *requestCtx {
	v, _ := ctx.Value(requestCtxKey).(*requestCtx)
	return v
}

// WithRequestCtx sets the base URL and, if not empty, the session token used as bearer auth in the context.
func WithRequestCtx(ctx context.Context, baseURL string, token string) context.Context {
	client := resty.New().SetBaseURL(baseURL)
	if token != "" {
		client.SetAuthToken(token)
	}

	return context.WithValue(ctx, requestCtxKey, &requestCtx{
		client:  client,
		baseURL: baseURL,
		token:   token,
	})
}
