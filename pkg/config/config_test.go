package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Default()

	if cfg.MCP.ServerURL == "" {
		t.Error("Expected default MCP server URL to be set")
	}

	if cfg.Skills.InstallDir == "" {
		t.Error("Expected default skills install dir to be set")
	}

	if cfg.IsLoggedIn() {
		t.Error("Expected default config to not be logged in")
	}
}

func TestConfigIsLoggedIn(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expiry   time.Time
		expected bool
	}{
		{
			name:     "not logged in - no token",
			token:    "",
			expiry:   time.Time{},
			expected: false,
		},
		{
			name:     "not logged in - zero expiry",
			token:    "test-token",
			expiry:   time.Time{},
			expected: false,
		},
		{
			name:     "logged in",
			token:    "test-token",
			expiry:   time.Now().Add(1 * time.Hour),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Auth: AuthConfig{
					AccessToken:  tt.token,
					TokenExpiry:  tt.expiry,
				},
			}

			result := cfg.IsLoggedIn()
			if result != tt.expected {
				t.Errorf("IsLoggedIn() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfigIsTokenExpired(t *testing.T) {
	tests := []struct {
		name     string
		expiry   time.Time
		expected bool
	}{
		{
			name:     "token expired",
			expiry:   time.Now().Add(-1 * time.Hour),
			expected: true,
		},
		{
			name:     "token valid",
			expiry:   time.Now().Add(1 * time.Hour),
			expected: false,
		},
		{
			name:     "zero expiry",
			expiry:   time.Time{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Auth: AuthConfig{
					TokenExpiry: tt.expiry,
				},
			}

			result := cfg.IsTokenExpired()
			if result != tt.expected {
				t.Errorf("IsTokenExpired() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for this test
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	// Create a config
	cfg := &Config{
		Auth: AuthConfig{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			TokenExpiry:  time.Now().Add(1 * time.Hour),
			UserID:       "user-123",
			Email:        "test@example.com",
		},
		MCP: MCPConfig{
			ServerURL: "https://api.example.com",
			APIToken:  "test-api-token",
		},
	}

	// Save the config
	err := cfg.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify the config file was created
	configPath := filepath.Join(tmpDir, configDir, configFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file not created at %s", configPath)
	}

	// Load the config
	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the loaded config matches
	if loadedCfg.Auth.AccessToken != cfg.Auth.AccessToken {
		t.Errorf("AccessToken = %v, want %v", loadedCfg.Auth.AccessToken, cfg.Auth.AccessToken)
	}

	if loadedCfg.Auth.Email != cfg.Auth.Email {
		t.Errorf("Email = %v, want %v", loadedCfg.Auth.Email, cfg.Auth.Email)
	}

	if loadedCfg.MCP.ServerURL != cfg.MCP.ServerURL {
		t.Errorf("ServerURL = %v, want %v", loadedCfg.MCP.ServerURL, cfg.MCP.ServerURL)
	}
}

func TestConfigClear(t *testing.T) {
	cfg := &Config{
		Auth: AuthConfig{
			AccessToken:  "test-token",
			TokenExpiry:  time.Now().Add(1 * time.Hour),
			UserID:       "user-123",
			Email:        "test@example.com",
		},
		MCP: MCPConfig{
			ServerURL: "https://api.example.com",
		},
	}

	// Verify logged in
	if !cfg.IsLoggedIn() {
		t.Error("Expected to be logged in before clear")
	}

	// Clear auth
	cfg.Clear()

	// Verify cleared
	if cfg.IsLoggedIn() {
		t.Error("Expected to not be logged in after clear")
	}

	if cfg.Auth.AccessToken != "" {
		t.Error("Expected AccessToken to be cleared")
	}

	if cfg.Auth.Email != "" {
		t.Error("Expected Email to be cleared")
	}

	// MCP config should remain
	if cfg.MCP.ServerURL == "" {
		t.Error("Expected MCP ServerURL to remain after clear")
	}
}

func TestGetConfigPath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for this test
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, configDir, configFile)
	if path != expectedPath {
		t.Errorf("GetConfigPath() = %v, want %v", path, expectedPath)
	}
}

func TestGetSkillsDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for this test
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	path, err := GetSkillsDir()
	if err != nil {
		t.Fatalf("GetSkillsDir() failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, configDir, skillsDir)
	if path != expectedPath {
		t.Errorf("GetSkillsDir() = %v, want %v", path, expectedPath)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for this test
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	// Ensure config dir
	err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() failed: %v", err)
	}

	// Verify directories were created
	configPath := filepath.Join(tmpDir, configDir)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config directory not created at %s", configPath)
	}

	skillsPath := filepath.Join(tmpDir, configDir, skillsDir)
	if _, err := os.Stat(skillsPath); os.IsNotExist(err) {
		t.Errorf("Skills directory not created at %s", skillsPath)
	}
}

func TestLoadReturnsDefaultWhenNotExists(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for this test
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	// Load when config doesn't exist
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Should return default config
	if cfg.MCP.ServerURL == "" {
		t.Error("Expected default config to be returned")
	}
}
