package polygo

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestExchangePersonalAccessKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc/auth/personalAccessKey/exchange" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("exchange must not send an Authorization header")
		}

		var body map[string]map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["json"]["key"] != "pak_test" || body["json"]["organizationId"] != "org_1" {
			t.Fatalf("unexpected body %v", body)
		}
		if _, ok := body["json"]["organizationSlug"]; ok {
			t.Fatalf("organizationSlug must be omitted when empty")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"json":{"token":"tok","expiresAt":"2026-01-01T00:00:00.000Z","organizationId":"org_1","scope":"organization"},"meta":[[1,"expiresAt"]]}`)
	}))
	defer srv.Close()

	session, err := NewClient("pak_test", srv.URL, Organization{ID: "org_1"}).ExchangePersonalAccessKey(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if session.Token != "tok" || session.ExpiresAt.IsZero() {
		t.Fatalf("unexpected session %+v", session)
	}
}

func TestExchangePersonalAccessKeyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"json":{"defined":false,"code":"UNAUTHORIZED","status":401,"message":"Invalid personal access key"}}`)
	}))
	defer srv.Close()

	_, err := NewClient("pak_bad", srv.URL, Organization{Slug: "org"}).ExchangePersonalAccessKey(context.Background())

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected api error, got %v", err)
	}
	if apiErr.Status != 401 || apiErr.Code != "UNAUTHORIZED" || apiErr.Message != "Invalid personal access key" {
		t.Fatalf("unexpected error %+v", apiErr)
	}
}

func TestUploadDataset(t *testing.T) {
	content := "id,name\n1,foo\n2,bar\n"
	filePath := filepath.Join(t.TempDir(), "query_result_1.csv")
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/upload/dataset/ds_1" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Fatalf("unexpected auth header %q", r.Header.Get("Authorization"))
		}
		if len(r.TransferEncoding) != 0 || r.ContentLength <= 0 {
			t.Fatalf("expected a fixed content length, got %d %v", r.ContentLength, r.TransferEncoding)
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		got, _ := io.ReadAll(file)
		if string(got) != content || header.Filename != "query_result_1.csv" || header.Header.Get("Content-Type") != "text/csv" {
			t.Fatalf("unexpected file %q %q %q", got, header.Filename, header.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"ingestId":"dsig_1","size":21,"contentType":"text/csv"}`)
	}))
	defer srv.Close()

	resp, err := NewClient("pak_test", srv.URL, Organization{ID: "org_1"}).UploadDataset(context.Background(), "tok", "ds_1", "text/csv", filePath)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IngestID != "dsig_1" {
		t.Fatalf("unexpected response %+v", resp)
	}
}

func TestUploadDatasetError(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(filePath, []byte("a\n1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPreconditionFailed)
		_, _ = io.WriteString(w, `{"error":"Data processing terms not acknowledged","code":"dpa.not-acknowledged"}`)
	}))
	defer srv.Close()

	_, err := NewClient("pak_test", srv.URL, Organization{ID: "org_1"}).UploadDataset(context.Background(), "tok", "ds_1", "text/csv", filePath)

	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Status != 412 || apiErr.Code != "dpa.not-acknowledged" {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestGetDatasetIngestStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc/dataset/getDatasetIngestStatus" || r.Header.Get("Authorization") != "Bearer tok" {
			t.Fatalf("unexpected request %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"json":{"status":"failed","errorMessage":"bad csv"}}`)
	}))
	defer srv.Close()

	status, err := NewClient("pak_test", srv.URL, Organization{ID: "org_1"}).GetDatasetIngestStatus(context.Background(), "tok", DatasetIngestStatusRequest{ID: "ds_1", IngestID: "dsig_1"})
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != IngestStatusFailed || status.ErrorMessage != "bad csv" {
		t.Fatalf("unexpected status %+v", status)
	}
}
