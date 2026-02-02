package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		responseBody   interface{}
		responseStatus int
		authToken      string
		apiToken       string
		expectError    bool
		expectAuth     bool
		expectAPIToken bool
	}{
		{
			name:           "successful GET request",
			method:         "GET",
			path:           "/test",
			responseBody:   map[string]string{"message": "success"},
			responseStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "successful POST request",
			method:         "POST",
			path:           "/test",
			body:           map[string]string{"key": "value"},
			responseBody:   map[string]string{"result": "created"},
			responseStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "API error response",
			method:         "GET",
			path:           "/error",
			responseBody:   map[string]string{"error": "test error"},
			responseStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "auth token is sent",
			method:         "GET",
			path:           "/auth-test",
			responseBody:   map[string]string{"auth": "valid"},
			responseStatus: http.StatusOK,
			authToken:      "test-jwt-token",
			expectError:    false,
			expectAuth:     true,
		},
		{
			name:           "API token is preferred over JWT",
			method:         "GET",
			path:           "/token-test",
			responseBody:   map[string]string{"auth": "valid"},
			responseStatus: http.StatusOK,
			authToken:      "test-jwt-token",
			apiToken:       "dp_test-api-token",
			expectError:    false,
			expectAuth:     false,
			expectAPIToken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check method
				if r.Method != tt.method {
					t.Errorf("Expected method %s, got %s", tt.method, r.Method)
				}

				// Check path
				if r.URL.Path != tt.path {
					t.Errorf("Expected path %s, got %s", tt.path, r.URL.Path)
				}

				// Check auth header
				authHeader := r.Header.Get("Authorization")
				if tt.expectAPIToken {
					expectedAuth := "Bearer " + tt.apiToken
					if authHeader != expectedAuth {
						t.Errorf("Expected Authorization header %s, got %s", expectedAuth, authHeader)
					}
				} else if tt.expectAuth {
					expectedAuth := "Bearer " + tt.authToken
					if authHeader != expectedAuth {
						t.Errorf("Expected Authorization header %s, got %s", expectedAuth, authHeader)
					}
				}

				// Write response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			// Create client
			client := NewClient(server.URL)
			client.SetAuthToken(tt.authToken)
			client.SetAPIToken(tt.apiToken)

			// Make request
			var result map[string]string
			err := client.Do(tt.method, tt.path, tt.body, &result)

			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestClient_StartDeviceFlow(t *testing.T) {
	expectedResp := &DeviceFlowStartResponse{
		DeviceCode:              "test-device-code",
		UserCode:                "ABCD-1234",
		VerificationURI:         "https://example.com/verify",
		VerificationURIComplete: "https://example.com/verify?code=ABCD-1234",
		ExpiresIn:               600,
		Interval:                5,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/auth/device/start" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resp, err := client.StartDeviceFlow()

	if err != nil {
		t.Fatalf("StartDeviceFlow() failed: %v", err)
	}

	if resp.DeviceCode != expectedResp.DeviceCode {
		t.Errorf("DeviceCode = %v, want %v", resp.DeviceCode, expectedResp.DeviceCode)
	}

	if resp.UserCode != expectedResp.UserCode {
		t.Errorf("UserCode = %v, want %v", resp.UserCode, expectedResp.UserCode)
	}

	if resp.ExpiresIn != expectedResp.ExpiresIn {
		t.Errorf("ExpiresIn = %v, want %v", resp.ExpiresIn, expectedResp.ExpiresIn)
	}
}

func TestClient_PollDeviceFlow(t *testing.T) {
	tests := []struct {
		name           string
		deviceCode     string
		responseStatus int
		responseBody   interface{}
		expectError    bool
		expectStatus   string
	}{
		{
			name:       "pending authorization",
			deviceCode: "test-device-code",
			responseStatus: http.StatusAccepted,
			responseBody: map[string]string{
				"status": "authorization_pending",
			},
			expectError:  false,
			expectStatus: "authorization_pending",
		},
		{
			name:       "successful authorization",
			deviceCode: "test-device-code",
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"access_token": "test-access-token",
				"expires_in":   3600,
				"user": map[string]string{
					"id":    "user-123",
					"email": "test@example.com",
					"name":  "Test User",
					"role":  "developer",
				},
			},
			expectError: false,
		},
		{
			name:           "expired token",
			deviceCode:     "expired-code",
			responseStatus: http.StatusGone,
			responseBody: map[string]string{
				"error": "expired_token",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/auth/device/poll" {
					t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client := NewClient(server.URL)
			resp, err := client.PollDeviceFlow(tt.deviceCode)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.expectStatus != "" && resp.Status != tt.expectStatus {
				t.Errorf("Status = %v, want %v", resp.Status, tt.expectStatus)
			}
		})
	}
}

func TestClient_CreateToken(t *testing.T) {
	expectedResp := &CreateAPITokenResponse{
		ID:          "token-123",
		Name:        "Test Token",
		Token:       "dp_test_token_value",
		TokenPrefix: "dp_test...",
		Scope:       "personal",
		Role:        "developer",
		Teams:       []string{"team1"},
		ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
		CreatedAt:   time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/auth/tokens" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetAuthToken("test-auth-token")

	req := &CreateAPITokenRequest{
		Name:      "Test Token",
		Scope:     "personal",
		ExpiresIn: 365 * 24 * 60 * 60,
	}

	resp, err := client.CreateToken(req)

	if err != nil {
		t.Fatalf("CreateToken() failed: %v", err)
	}

	if resp.ID != expectedResp.ID {
		t.Errorf("ID = %v, want %v", resp.ID, expectedResp.ID)
	}

	if resp.Token != expectedResp.Token {
		t.Errorf("Token = %v, want %v", resp.Token, expectedResp.Token)
	}

	if resp.Name != expectedResp.Name {
		t.Errorf("Name = %v, want %v", resp.Name, expectedResp.Name)
	}
}

func TestClient_ListTokens(t *testing.T) {
	expectedResp := &APITokenListResponse{
		Tokens: []*APITokenResponse{
			{
				ID:          "token-1",
				Name:        "Token 1",
				TokenPrefix: "dp_abc...",
				Scope:       "personal",
				Role:        "developer",
				CreatedAt:   time.Now(),
			},
			{
				ID:          "token-2",
				Name:        "Token 2",
				TokenPrefix: "dp_xyz...",
				Scope:       "team",
				Role:        "admin",
				CreatedAt:   time.Now(),
			},
		},
		Total: 2,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/auth/tokens" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetAuthToken("test-auth-token")

	resp, err := client.ListTokens()

	if err != nil {
		t.Fatalf("ListTokens() failed: %v", err)
	}

	if resp.Total != expectedResp.Total {
		t.Errorf("Total = %v, want %v", resp.Total, expectedResp.Total)
	}

	if len(resp.Tokens) != len(expectedResp.Tokens) {
		t.Errorf("Tokens length = %v, want %v", len(resp.Tokens), len(expectedResp.Tokens))
	}
}

func TestClient_RevokeToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/auth/tokens/token-123" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Token revoked"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetAuthToken("test-auth-token")

	err := client.RevokeToken("token-123")

	if err != nil {
		t.Errorf("RevokeToken() failed: %v", err)
	}
}

func TestClient_GetMCPInfo(t *testing.T) {
	expectedResp := &MCPInfoResponse{
		Name:        "developer-platform",
		Version:     "1.0.0",
		Instructions: "Test instructions",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/mcp/info" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetAPIToken("test-api-token")

	resp, err := client.GetMCPInfo()

	if err != nil {
		t.Fatalf("GetMCPInfo() failed: %v", err)
	}

	if resp.Name != expectedResp.Name {
		t.Errorf("Name = %v, want %v", resp.Name, expectedResp.Name)
	}

	if resp.Version != expectedResp.Version {
		t.Errorf("Version = %v, want %v", resp.Version, expectedResp.Version)
	}
}

func TestClient_ListSkills(t *testing.T) {
	expectedResp := &SkillsListResponse{
		Skills: []Skill{
			{
				Name:        "deploy.md",
				Description: "Deploy current project",
				URL:         "/api/v1/skills/deploy.md",
			},
			{
				Name:        "restart.md",
				Description: "Rolling restart deployment",
				URL:         "/api/v1/skills/restart.md",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/skills" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	resp, err := client.ListSkills()

	if err != nil {
		t.Fatalf("ListSkills() failed: %v", err)
	}

	if len(resp.Skills) != len(expectedResp.Skills) {
		t.Errorf("Skills length = %v, want %v", len(resp.Skills), len(expectedResp.Skills))
	}
}

func TestClient_DownloadSkill(t *testing.T) {
	expectedContent := "# Test Skill\n\nThis is a test skill content."

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/skills/test.md" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "text/markdown")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	content, err := client.DownloadSkill("test.md")

	if err != nil {
		t.Fatalf("DownloadSkill() failed: %v", err)
	}

	if content != expectedContent {
		t.Errorf("Content = %v, want %v", content, expectedContent)
	}
}
