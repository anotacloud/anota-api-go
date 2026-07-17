package anota

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient spins up an httptest server and returns a Client pointed at it.
func newTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := New("test-key")
	client.BaseURL = server.URL
	return client, server
}

// (a) ListForms issues GET {base}/forms with the bearer header.
func TestListFormsSendsGetWithAuth(t *testing.T) {
	var gotMethod, gotPath, gotAuth string
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotAuth = r.Method, r.URL.Path, r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"forms":[]}`))
	})
	defer server.Close()

	if _, err := client.ListForms(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/forms" {
		t.Errorf("path = %q, want /forms", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("auth = %q, want %q", gotAuth, "Bearer test-key")
	}
}

// (b) CreateSubmission sends the exact JSON body {"answers":{"f_1":"hola"}}.
func TestCreateSubmissionSendsJSONBody(t *testing.T) {
	var gotBody, gotContentType string
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		gotBody = string(data)
		gotContentType = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{"id":"s_1"}`))
	})
	defer server.Close()

	if _, err := client.CreateSubmission(context.Background(), "form_1", map[string]any{"f_1": "hola"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"answers":{"f_1":"hola"}}`
	if gotBody != want {
		t.Errorf("body = %q, want %q", gotBody, want)
	}
	if gotContentType != "application/json" {
		t.Errorf("content-type = %q, want application/json", gotContentType)
	}
}

// (c) A 400 problem-details body becomes an *APIError with status + detail.
func TestErrorResponseRaisesAPIError(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"Error: bad"}`))
	})
	defer server.Close()

	_, err := client.ListForms(context.Background())
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Status != 400 {
		t.Errorf("status = %d, want 400", apiErr.Status)
	}
	if apiErr.Message != "Error: bad" {
		t.Errorf("message = %q, want %q", apiErr.Message, "Error: bad")
	}
}

// (d) ListSubmissions builds ?page=2&pageSize=10&status=New.
func TestListSubmissionsBuildsQuery(t *testing.T) {
	var gotQuery string
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"submissions":[]}`))
	})
	defer server.Close()

	if _, err := client.ListSubmissions(context.Background(), "form_1", 2, 10, "New"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "page=2&pageSize=10&status=New"
	if gotQuery != want {
		t.Errorf("query = %q, want %q", gotQuery, want)
	}
}

// status is omitted from the query when empty, and defaults fill in.
func TestListSubmissionsOmitsEmptyStatus(t *testing.T) {
	var gotQuery string
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"submissions":[]}`))
	})
	defer server.Close()

	if _, err := client.ListSubmissions(context.Background(), "form_1", 0, 0, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "page=1&pageSize=25"
	if gotQuery != want {
		t.Errorf("query = %q, want %q", gotQuery, want)
	}
}

// An empty 2xx body decodes to (nil, nil).
func TestEmptyBodyReturnsNil(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	result, err := client.DeleteForm(context.Background(), "form_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("result = %v, want nil", result)
	}
}
