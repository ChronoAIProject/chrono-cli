package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an API client for the Developer Platform
type Client struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
	apiToken   string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetAuthToken sets the JWT authentication token
func (c *Client) SetAuthToken(token string) {
	c.authToken = token
}

// SetAPIToken sets the API token
func (c *Client) SetAPIToken(token string) {
	c.apiToken = token
}

// Do performs an HTTP request with authentication
func (c *Client) Do(method, path string, body interface{}, response interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Prefer API token over JWT for API calls
	if c.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiToken)
	} else if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response body if provided
	if response != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// ============================================
// Device Flow Types
// ============================================

// DeviceFlowStartRequest represents a request to start device flow
type DeviceFlowStartRequest struct{}

// DeviceFlowStartResponse represents the response from starting device flow
type DeviceFlowStartResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// DeviceFlowPollRequest represents a request to poll device flow
type DeviceFlowPollRequest struct {
	DeviceCode string `json:"device_code"`
}

// DeviceFlowPollResponse represents the response from polling device flow
type DeviceFlowPollResponse struct {
	Status     string `json:"status,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	User        User   `json:"user,omitempty"`
}

// User represents a user
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	User        User   `json:"user"`
}

// ============================================
// Device Flow Methods
// ============================================

// StartDeviceFlow initiates the Keycloak device flow
func (c *Client) StartDeviceFlow() (*DeviceFlowStartResponse, error) {
	var resp DeviceFlowStartResponse
	err := c.Do("POST", "/auth/device/start", nil, &resp)
	return &resp, err
}

// PollDeviceFlow polls for device flow completion
func (c *Client) PollDeviceFlow(deviceCode string) (*DeviceFlowPollResponse, error) {
	req := DeviceFlowPollRequest{DeviceCode: deviceCode}
	var resp DeviceFlowPollResponse
	err := c.Do("POST", "/auth/device/poll", req, &resp)
	return &resp, err
}

// ============================================
// API Token Methods
// ============================================

// CreateAPITokenRequest represents a request to create an API token
type CreateAPITokenRequest struct {
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	TeamID    string `json:"team_id,omitempty"`
	ExpiresIn int    `json:"expires_in,omitempty"`
}

// CreateAPITokenResponse represents the response from creating an API token
type CreateAPITokenResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Token       string `json:"token"`
	TokenPrefix string `json:"token_prefix"`
	Scope       string `json:"scope"`
	TeamID      string `json:"team_id,omitempty"`
	Role        string `json:"role"`
	Teams       []string `json:"teams"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// APITokenResponse represents an API token (without the actual token)
type APITokenResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	TokenPrefix string    `json:"token_prefix"`
	Scope       string    `json:"scope"`
	TeamID      string    `json:"team_id,omitempty"`
	Role        string    `json:"role"`
	Teams       []string  `json:"teams"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// APITokenListResponse represents a list of API tokens
type APITokenListResponse struct {
	Tokens []*APITokenResponse `json:"tokens"`
	Total  int                 `json:"total"`
}

// CreateToken creates a new API token
func (c *Client) CreateToken(req *CreateAPITokenRequest) (*CreateAPITokenResponse, error) {
	var resp CreateAPITokenResponse
	err := c.Do("POST", "/auth/tokens", req, &resp)
	return &resp, err
}

// ListTokens lists all API tokens for the current user
func (c *Client) ListTokens() (*APITokenListResponse, error) {
	var resp APITokenListResponse
	err := c.Do("GET", "/auth/tokens", nil, &resp)
	return &resp, err
}

// RevokeToken revokes a specific API token
func (c *Client) RevokeToken(tokenID string) error {
	return c.Do("DELETE", "/auth/tokens/"+tokenID, nil, nil)
}

// ============================================
// MCP Methods
// ============================================

// MCPInfoResponse represents MCP server information
type MCPInfoResponse struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Instructions string `json:"instructions"`
}

// GetMCPInfo gets MCP server information
func (c *Client) GetMCPInfo() (*MCPInfoResponse, error) {
	var resp MCPInfoResponse
	err := c.Do("GET", "/mcp/info", nil, &resp)
	return &resp, err
}

// ============================================
// Skills Methods
// ============================================

// Skill represents a skill
type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// SkillsListResponse represents a list of skills
type SkillsListResponse struct {
	Skills []Skill `json:"skills"`
}

// ListSkills lists all available skills
func (c *Client) ListSkills() (*SkillsListResponse, error) {
	var resp SkillsListResponse
	err := c.Do("GET", "/skills", nil, &resp)
	return &resp, err
}

// DownloadSkill downloads a skill by name
func (c *Client) DownloadSkill(name string) (string, error) {
	url := c.baseURL + "/skills/" + name
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/markdown")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download skill: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download skill (status %d)", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read skill content: %w", err)
	}

	return string(content), nil
}
