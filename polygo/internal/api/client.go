package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	rpcURL           = "/rpc/"
	datasetUploadURL = "/api/upload/dataset/"
)

// rpcEnvelope wraps request and response bodies of the Polyteia API RPC endpoints.
type rpcEnvelope[T any] struct {
	JSON T `json:"json"`
}

// RPC calls a Polyteia API RPC endpoint (e.g. "dataset/getDatasetIngestStatus").
func RPC[T any, V any](ctx context.Context, procedure string, input T) (*V, error) {
	requestCtx := getRequestCtx(ctx)
	if requestCtx == nil {
		return nil, fmt.Errorf("api client: missing request context")
	}

	var result rpcEnvelope[V]
	var rpcErr rpcEnvelope[Error]

	resp, err := requestCtx.client.R().
		SetContext(ctx).
		SetBody(rpcEnvelope[T]{JSON: input}).
		SetResult(&result).
		SetError(&rpcErr).
		Post(rpcURL + procedure)
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, newError(resp.StatusCode(), &rpcErr.JSON, resp.Bytes())
	}

	return &result.JSON, nil
}

// UploadDataset streams a file as multipart/form-data to the dataset upload endpoint.
// The response body is decoded into V.
func UploadDataset[V any](ctx context.Context, datasetID string, contentType string, filePath string) (*V, error) {
	requestCtx := getRequestCtx(ctx)
	if requestCtx == nil {
		return nil, fmt.Errorf("api client: missing request context")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close() //nolint:errcheck

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Build the multipart envelope around the file so the body can be streamed
	// with a known content length instead of buffering the whole file in memory.
	var envelope bytes.Buffer
	writer := multipart.NewWriter(&envelope)

	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, escapeQuotes(filepath.Base(filePath))))
	partHeader.Set("Content-Type", contentType)
	if _, err := writer.CreatePart(partHeader); err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	headLen := envelope.Len()

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}
	head := envelope.Bytes()[:headLen]
	tail := envelope.Bytes()[headLen:]

	body := io.MultiReader(bytes.NewReader(head), file, bytes.NewReader(tail))
	uploadURL := requestCtx.baseURL + datasetUploadURL + url.PathEscape(datasetID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.ContentLength = int64(len(head)) + stat.Size() + int64(len(tail))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if requestCtx.token != "" {
		req.Header.Set("Authorization", "Bearer "+requestCtx.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var uploadErr struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}
		_ = json.Unmarshal(respBody, &uploadErr)

		return nil, newError(resp.StatusCode, &Error{Code: uploadErr.Code, Message: uploadErr.Error}, respBody)
	}

	var v V
	if err := json.Unmarshal(respBody, &v); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	return &v, nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
