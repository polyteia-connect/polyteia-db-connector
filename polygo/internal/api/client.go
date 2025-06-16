package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type requestType string

const (
	apiURL      = "/api"
	uploadURL   = "/upload"
	downloadURL = "/download"

	commandRequestType requestType = "command"
	queryRequestType   requestType = "query"
)

// Command makes a command request to the API.
func Command[T any, V any](ctx context.Context, command string, requestParams T) (*V, error) {
	return makeRequest[T, V](ctx, commandRequestType, command, requestParams)
}

// Query makes a query request to the API.
func Query[T any, V any](ctx context.Context, query string, requestParams T) (*V, error) {
	return makeRequest[T, V](ctx, queryRequestType, query, requestParams)
}

// Upload makes a upload request to the API.
func Upload(ctx context.Context, uploadToken string, filePath string) error {
	requestCtx := getRequestCtx(ctx)
	if requestCtx == nil {
		return fmt.Errorf("api client: missing request context")
	}

	baseURL := requestCtx.client.BaseURL()
	if baseURL == "" {
		return fmt.Errorf("base URL not set in client")
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close() //nolint:errcheck

	// Get file stats
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %v", err)
	}

	// Create a buffer to store the multipart form
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Create the file part
	part, err := writer.CreateFormFile("file", stat.Name())
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	// Copy the file into the part
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// Close the writer to finalize the multipart form
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	// Create the request with the complete body
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+uploadURL, &b)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("X-Upload-Token", uploadToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.ContentLength = int64(b.Len())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}

		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Download makes a download request to the API.
func Download(ctx context.Context, downloadToken string) (io.Reader, error) {
	requestCtx := getRequestCtx(ctx)
	if requestCtx == nil {
		return nil, fmt.Errorf("api client: missing request context")
	}

	response, err := requestCtx.client.R().SetQueryParam("token", downloadToken).SetContext(ctx).Get(downloadURL)
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

func makeRequest[T any, V any](ctx context.Context, reqType requestType, cmdOrQuery string, requestParams T) (*V, error) {
	requestCtx := getRequestCtx(ctx)
	if requestCtx == nil {
		return nil, fmt.Errorf("api client: missing request context")
	}

	body := map[string]any{
		string(reqType): cmdOrQuery,
		"params":        requestParams,
	}

	var result map[string]any
	var v V

	_, err := requestCtx.client.R().SetContext(ctx).SetBody(body).SetResult(&result).Post(apiURL)
	if err != nil {
		return nil, err
	}

	if err := getError(result); err != nil {
		return nil, err
	}

	if err := getData(result, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func getData(result map[string]any, dataAs any) error {
	if result == nil {
		return nil
	}

	if val, ok := result["data"]; ok {
		// unmarshal error
		raw, err := json.Marshal(val)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(raw, dataAs); err != nil {
			return err
		}

		return nil
	}

	return nil
}
