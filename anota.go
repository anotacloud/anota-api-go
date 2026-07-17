// Package anota is a thin, dependency-free client for the anota REST API
// (https://anota.cloud) built entirely on the Go standard library.
//
// Every method returns the server's JSON decoded into a map[string]any, or an
// *APIError for any non-2xx response. Create an API key at
// https://anota.cloud/api-keys and pass it to New.
package anota

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// DefaultBaseURL is the production anota API root. Override Client.BaseURL to
// target a different environment.
const DefaultBaseURL = "https://anota.cloud/api/v1"

// APIError is returned for any non-2xx response. Message is taken from the
// response body's problem-details "detail" field, falling back to "title",
// then to the raw body. Network failures surface as the native net/http error,
// not an *APIError.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("anota: HTTP %d: %s", e.Status, e.Message)
}

// Client talks to the anota API. Construct one with New; the zero value is not
// usable. APIKey and BaseURL may be adjusted after construction, and a custom
// HTTPClient may be supplied (for timeouts, proxies, or tests).
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// New returns a Client for the given API key pointed at the production API.
func New(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    DefaultBaseURL,
		HTTPClient: http.DefaultClient,
	}
}

// request is the single HTTP core every method delegates to: it builds the URL
// and query, attaches the bearer token, JSON-encodes the body, and turns a
// non-2xx status into an *APIError. An empty 2xx body decodes to (nil, nil).
func (c *Client) request(ctx context.Context, method, path string, body any, query url.Values) (map[string]any, error) {
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = DefaultBaseURL
	}
	fullURL := base + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{Status: resp.StatusCode, Message: extractMessage(respBody)}
	}

	if len(bytes.TrimSpace(respBody)) == 0 {
		return nil, nil
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// extractMessage pulls a human-readable message out of an error body, preferring
// the ASP.NET problem-details "detail" field, then "title", then the raw body.
func extractMessage(body []byte) string {
	var problem map[string]any
	if json.Unmarshal(body, &problem) == nil {
		if detail, ok := problem["detail"].(string); ok && detail != "" {
			return detail
		}
		if title, ok := problem["title"].(string); ok && title != "" {
			return title
		}
	}
	return string(body)
}

// ----- forms -----

// ListForms returns every form in the workspace.
func (c *Client) ListForms(ctx context.Context) (map[string]any, error) {
	return c.request(ctx, http.MethodGet, "/forms", nil, nil)
}

// CreateForm creates a form with the given title and fields. Pass an empty
// description to omit it.
func (c *Client) CreateForm(ctx context.Context, title string, fields []map[string]any, description string) (map[string]any, error) {
	body := map[string]any{"title": title, "fields": fields}
	if description != "" {
		body["description"] = description
	}
	return c.request(ctx, http.MethodPost, "/forms", body, nil)
}

// GetForm returns a single form by id.
func (c *Client) GetForm(ctx context.Context, formID string) (map[string]any, error) {
	return c.request(ctx, http.MethodGet, "/forms/"+formID, nil, nil)
}

// AddFields appends fields to a form.
func (c *Client) AddFields(ctx context.Context, formID string, fields []map[string]any) (map[string]any, error) {
	return c.request(ctx, http.MethodPost, "/forms/"+formID+"/fields", map[string]any{"fields": fields}, nil)
}

// EditField updates one field on a form.
func (c *Client) EditField(ctx context.Context, formID, fieldID string, field map[string]any) (map[string]any, error) {
	path := "/forms/" + formID + "/fields/" + url.PathEscape(fieldID)
	return c.request(ctx, http.MethodPatch, path, map[string]any{"field": field}, nil)
}

// DeleteField removes one field from a form.
func (c *Client) DeleteField(ctx context.Context, formID, fieldID string) (map[string]any, error) {
	path := "/forms/" + formID + "/fields/" + url.PathEscape(fieldID)
	return c.request(ctx, http.MethodDelete, path, nil, nil)
}

// PublishForm publishes a form. Once published, existing fields are locked.
func (c *Client) PublishForm(ctx context.Context, formID string) (map[string]any, error) {
	return c.request(ctx, http.MethodPost, "/forms/"+formID+"/publish", nil, nil)
}

// RenameForm changes a form's title.
func (c *Client) RenameForm(ctx context.Context, formID, title string) (map[string]any, error) {
	return c.request(ctx, http.MethodPatch, "/forms/"+formID, map[string]any{"title": title}, nil)
}

// SetPdfTemplate assigns a PDF template (by storage key) to a form.
func (c *Client) SetPdfTemplate(ctx context.Context, formID, key string) (map[string]any, error) {
	return c.request(ctx, http.MethodPut, "/forms/"+formID+"/pdf-template", map[string]any{"key": key}, nil)
}

// DeleteForm deletes a form and its submissions.
func (c *Client) DeleteForm(ctx context.Context, formID string) (map[string]any, error) {
	return c.request(ctx, http.MethodDelete, "/forms/"+formID, nil, nil)
}

// CloneForm duplicates a form.
func (c *Client) CloneForm(ctx context.Context, formID string) (map[string]any, error) {
	return c.request(ctx, http.MethodPost, "/forms/"+formID+"/clone", nil, nil)
}

// ----- logic rules -----

// AddLogicRules adds conditional-logic rules to a form.
func (c *Client) AddLogicRules(ctx context.Context, formID string, rules []map[string]any) (map[string]any, error) {
	return c.request(ctx, http.MethodPost, "/forms/"+formID+"/logic-rules", map[string]any{"rules": rules}, nil)
}

// EditLogicRule replaces one logic rule on a form.
func (c *Client) EditLogicRule(ctx context.Context, formID, ruleID string, rule map[string]any) (map[string]any, error) {
	path := "/forms/" + formID + "/logic-rules/" + url.PathEscape(ruleID)
	return c.request(ctx, http.MethodPut, path, map[string]any{"rule": rule}, nil)
}

// DeleteLogicRule removes one logic rule from a form.
func (c *Client) DeleteLogicRule(ctx context.Context, formID, ruleID string) (map[string]any, error) {
	path := "/forms/" + formID + "/logic-rules/" + url.PathEscape(ruleID)
	return c.request(ctx, http.MethodDelete, path, nil, nil)
}

// ----- submissions -----

// ListSubmissions returns a page of a form's submissions. page defaults to 1
// and pageSize to 25 when non-positive; pass an empty status to list all.
func (c *Client) ListSubmissions(ctx context.Context, formID string, page, pageSize int, status string) (map[string]any, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 25
	}
	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("pageSize", strconv.Itoa(pageSize))
	if status != "" {
		query.Set("status", status)
	}
	return c.request(ctx, http.MethodGet, "/forms/"+formID+"/submissions", nil, query)
}

// GetSubmission returns a single submission by id.
func (c *Client) GetSubmission(ctx context.Context, submissionID string) (map[string]any, error) {
	return c.request(ctx, http.MethodGet, "/submissions/"+submissionID, nil, nil)
}

// CreateSubmission records a submission against a form. answers is keyed by
// field id; values are strings or []string.
func (c *Client) CreateSubmission(ctx context.Context, formID string, answers map[string]any) (map[string]any, error) {
	return c.request(ctx, http.MethodPost, "/forms/"+formID+"/submissions", map[string]any{"answers": answers}, nil)
}

// SetSubmissionStatus changes a submission's status (New, Read, Flagged, Spam).
func (c *Client) SetSubmissionStatus(ctx context.Context, submissionID, status string) (map[string]any, error) {
	return c.request(ctx, http.MethodPatch, "/submissions/"+submissionID+"/status", map[string]any{"status": status}, nil)
}

// DeleteSubmission deletes a submission.
func (c *Client) DeleteSubmission(ctx context.Context, submissionID string) (map[string]any, error) {
	return c.request(ctx, http.MethodDelete, "/submissions/"+submissionID, nil, nil)
}

// SubmissionStats returns aggregate statistics for a form's submissions.
func (c *Client) SubmissionStats(ctx context.Context, formID string) (map[string]any, error) {
	return c.request(ctx, http.MethodGet, "/forms/"+formID+"/stats", nil, nil)
}

// ----- templates -----

// ListTemplates returns the template gallery. language defaults to "es".
func (c *Client) ListTemplates(ctx context.Context, language string) (map[string]any, error) {
	if language == "" {
		language = "es"
	}
	query := url.Values{}
	query.Set("language", language)
	return c.request(ctx, http.MethodGet, "/templates", nil, query)
}

// CreateFormFromTemplate creates a new form from a template.
func (c *Client) CreateFormFromTemplate(ctx context.Context, templateID string) (map[string]any, error) {
	return c.request(ctx, http.MethodPost, "/forms/from-template/"+templateID, nil, nil)
}

// ----- webhooks -----

// ListWebhooks returns the webhooks registered on a form.
func (c *Client) ListWebhooks(ctx context.Context, formID string) (map[string]any, error) {
	return c.request(ctx, http.MethodGet, "/forms/"+formID+"/webhooks", nil, nil)
}

// AddWebhook registers a webhook URL on a form.
func (c *Client) AddWebhook(ctx context.Context, formID, webhookURL string) (map[string]any, error) {
	return c.request(ctx, http.MethodPost, "/forms/"+formID+"/webhooks", map[string]any{"url": webhookURL}, nil)
}

// DeleteWebhook removes a webhook from a form.
func (c *Client) DeleteWebhook(ctx context.Context, formID, webhookID string) (map[string]any, error) {
	return c.request(ctx, http.MethodDelete, "/forms/"+formID+"/webhooks/"+webhookID, nil, nil)
}
