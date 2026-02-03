package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	configDir      = ".chrono"
	configFile     = "config.yaml"
	skillsDir      = "skills"
	// defaultBaseURL is the base API URL for API calls
	// MCP configs will append /mcp when writing to editor config files
	defaultBaseURL = "https://platform.aelf.dev/api/v1"
)

// Config represents the CLI configuration
type Config struct {
	Auth   AuthConfig   `yaml:"auth"`
	MCP    MCPConfig    `yaml:"mcp"`
	Skills SkillsConfig `yaml:"skills"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	AccessToken  string    `yaml:"access_token"`
	RefreshToken string    `yaml:"refresh_token"`
	TokenExpiry  time.Time `yaml:"token_expiry,omitempty"`
	UserID       string    `yaml:"user_id,omitempty"`
	Email        string    `yaml:"email,omitempty"`
}

// MarshalYAML customizes YAML marshaling for AuthConfig
func (a AuthConfig) MarshalYAML() (interface{}, error) {
	// Create a map to manually control what gets serialized
	type Alias AuthConfig
	aux := &struct {
		TokenExpiry *string `yaml:"token_expiry,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(&a),
	}

	// Only include token_expiry if it's not zero
	if !a.TokenExpiry.IsZero() {
		formatted := a.TokenExpiry.Format(time.RFC3339)
		aux.TokenExpiry = &formatted
	}

	return aux, nil
}

// UnmarshalYAML customizes YAML unmarshaling for AuthConfig
func (a *AuthConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type Alias AuthConfig
	aux := &struct {
		TokenExpiry string `yaml:"token_expiry"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := unmarshal(aux); err != nil {
		return err
	}

	// Parse token_expiry if it's not empty
	if aux.TokenExpiry != "" {
		t, err := time.Parse(time.RFC3339, aux.TokenExpiry)
		if err != nil {
			return fmt.Errorf("failed to parse token_expiry: %w", err)
		}
		a.TokenExpiry = t
	}

	return nil
}

// MCPConfig represents MCP server configuration
type MCPConfig struct {
	ServerURL string `yaml:"server_url"`
	APIToken  string `yaml:"api_token"`
}

// SkillsConfig represents skills configuration
type SkillsConfig struct {
	InstallDir string           `yaml:"install_dir"`
	Installed  []InstalledSkill `yaml:"installed"`
	GitHubRepo string           `yaml:"github_repo"` // e.g., "aelfdevops/chrono-skills"
	GitHubRef  string           `yaml:"github_ref"`  // branch, tag, or commit (default: "main")
}

// InstalledSkill represents an installed skill
type InstalledSkill struct {
	Name        string    `yaml:"name"`
	URL         string    `yaml:"url"`
	InstalledAt time.Time `yaml:"installed_at"`
}

// IsLoggedIn checks if the user is logged in
func (c *Config) IsLoggedIn() bool {
	return c.Auth.AccessToken != "" && !c.Auth.TokenExpiry.IsZero()
}

// IsTokenExpired checks if the access token is expired
func (c *Config) IsTokenExpired() bool {
	return c.Auth.TokenExpiry.Before(time.Now())
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, configDir, configFile), nil
}

// GetSkillsDir returns the path to the skills directory
func GetSkillsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, configDir, skillsDir), nil
}

// EnsureConfigDir ensures the config directory exists
func EnsureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, configDir)
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	skillsPath := filepath.Join(homeDir, configDir, skillsDir)
	if err := os.MkdirAll(skillsPath, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	return nil
}

// Load loads the configuration from the config file
// If the config file doesn't exist, returns a default config
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Default(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to the config file
func (c *Config) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Clear clears the authentication data from the config
func (c *Config) Clear() {
	c.Auth = AuthConfig{}
}

// Default returns a default configuration
func Default() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}

	return &Config{
		MCP: MCPConfig{
			ServerURL: defaultBaseURL,
		},
		Skills: SkillsConfig{
			InstallDir: filepath.Join(homeDir, configDir, skillsDir),
			GitHubRepo: "ChronoAIProject/chrono-cli", // Chrono CLI repository (contains skills/)
			GitHubRef:  "main",                        // Default branch
		},
	}
}

// GetSkillsGitHubURL returns the raw GitHub content URL for a skill
func (c *Config) GetSkillsGitHubURL(skillName string) string {
	if c.Skills.GitHubRepo == "" {
		return ""
	}

	// Use "main" as default ref if not specified
	ref := c.Skills.GitHubRef
	if ref == "" {
		ref = "main"
	}

	// Ensure skill name has .md extension
	if !strings.HasSuffix(skillName, ".md") {
		skillName += ".md"
	}

	// Build raw GitHub URL: https://raw.githubusercontent.com/{owner}/{repo}/{ref}/skills/{name}
	// Note: GitHubRepo format is "owner/repo"
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/skills/%s",
		c.Skills.GitHubRepo+"/"+ref, skillName)
}

// HasGitHubSkills returns true if GitHub skills repo is configured
func (c *Config) HasGitHubSkills() bool {
	return c.Skills.GitHubRepo != ""
}
