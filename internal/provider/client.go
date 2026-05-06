package provider

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is the CiviCRM API v4 HTTP client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// APIResponse represents the standard CiviCRM API v4 response
type APIResponse struct {
	Version      int              `json:"version"`
	Count        int              `json:"count"`
	Values       []map[string]any `json:"values"`
	ErrorCode    int              `json:"error_code,omitempty"`
	ErrorMessage string           `json:"error_message,omitempty"`
}

// NewClient creates a new CiviCRM API client
func NewClient(baseURL, apiKey string, insecure bool) (*Client, error) {
	// Normalize the base URL
	baseURL = strings.TrimSuffix(baseURL, "/")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}, nil
}

// buildEndpoint constructs the API endpoint URL
func (c *Client) buildEndpoint(entity, action string) string {
	return fmt.Sprintf("%s/civicrm/ajax/api4/%s/%s", c.baseURL, entity, action)
}

// doRequest performs an HTTP request to the CiviCRM API
func (c *Client) doRequest(method, endpoint string, params map[string]any) (*APIResponse, error) {
	// Encode parameters as JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	// URL encode the params
	formData := url.Values{}
	formData.Set("params", string(paramsJSON))

	var req *http.Request
	if method == http.MethodGet {
		reqURL := endpoint + "?" + formData.Encode()
		req, err = http.NewRequest(method, reqURL, nil)
	} else {
		req, err = http.NewRequest(method, endpoint, bytes.NewBufferString(formData.Encode()))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	// Check for API errors
	if apiResp.ErrorCode != 0 || apiResp.ErrorMessage != "" {
		return nil, fmt.Errorf("API error %d: %s", apiResp.ErrorCode, apiResp.ErrorMessage)
	}

	return &apiResp, nil
}

// Create creates a new entity
func (c *Client) Create(entity string, values map[string]any) (map[string]any, error) {
	endpoint := c.buildEndpoint(entity, "create")

	params := map[string]any{
		"values": values,
	}

	resp, err := c.doRequest(http.MethodPost, endpoint, params)
	if err != nil {
		return nil, err
	}

	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("no values returned from create operation")
	}

	return resp.Values[0], nil
}

// Get retrieves entities by ID or filter
func (c *Client) Get(entity string, where [][]any, select_ []string) ([]map[string]any, error) {
	endpoint := c.buildEndpoint(entity, "get")

	params := map[string]any{
		"where": where,
	}
	if len(select_) > 0 {
		params["select"] = select_
	}

	resp, err := c.doRequest(http.MethodPost, endpoint, params)
	if err != nil {
		return nil, err
	}

	return resp.Values, nil
}

// GetByID retrieves a single entity by ID
func (c *Client) GetByID(entity string, id int64, select_ []string) (map[string]any, error) {
	where := [][]any{
		{"id", "=", id},
	}

	results, err := c.Get(entity, where, select_)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s with ID %d not found", entity, id)
	}

	return results[0], nil
}

// Update updates an existing entity
func (c *Client) Update(entity string, id int64, values map[string]any) (map[string]any, error) {
	endpoint := c.buildEndpoint(entity, "update")

	params := map[string]any{
		"where": [][]any{
			{"id", "=", id},
		},
		"values": values,
	}

	resp, err := c.doRequest(http.MethodPost, endpoint, params)
	if err != nil {
		return nil, err
	}

	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("no values returned from update operation")
	}

	return resp.Values[0], nil
}

// Delete deletes an entity by ID
func (c *Client) Delete(entity string, id int64) error {
	endpoint := c.buildEndpoint(entity, "delete")

	params := map[string]any{
		"where": [][]any{
			{"id", "=", id},
		},
	}

	_, err := c.doRequest(http.MethodPost, endpoint, params)
	return err
}

// Helper functions for type conversion

// GetInt64 safely extracts an int64 from a map value
func GetInt64(m map[string]any, key string) (int64, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return int64(val), true
	case int64:
		return val, true
	case int:
		return int64(val), true
	case json.Number:
		i, err := val.Int64()
		return i, err == nil
	default:
		return 0, false
	}
}

// GetString safely extracts a string from a map value.
// CiviCRM's API v4 may return string-typed fields (e.g. OptionValue.value) as
// JSON numbers, so numeric types are converted to their decimal string form.
func GetString(m map[string]any, key string) (string, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return "", false
	}
	switch val := v.(type) {
	case string:
		return val, true
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10), true
		}
		return strconv.FormatFloat(val, 'f', -1, 64), true
	case int64:
		return strconv.FormatInt(val, 10), true
	case int:
		return strconv.FormatInt(int64(val), 10), true
	case json.Number:
		return val.String(), true
	default:
		return "", false
	}
}

// GetBool safely extracts a bool from a map value
func GetBool(m map[string]any, key string) (bool, bool) {
	v, ok := m[key]
	if !ok {
		return false, false
	}
	switch val := v.(type) {
	case bool:
		return val, true
	case float64:
		return val == 1, true
	case int:
		return val == 1, true
	case string:
		return val == "1" || val == "true", true
	default:
		return false, false
	}
}

// GetOptionGroupID retrieves the numeric ID of an option group by name
func (c *Client) GetOptionGroupID(name string) (int64, error) {
	where := [][]any{
		{"name", "=", name},
	}

	results, err := c.Get("OptionGroup", where, []string{"id"})
	if err != nil {
		return 0, fmt.Errorf("failed to look up option group '%s': %w", name, err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("option group '%s' not found", name)
	}

	id, ok := GetInt64(results[0], "id")
	if !ok {
		return 0, fmt.Errorf("option group '%s' has no valid id", name)
	}

	return id, nil
}
