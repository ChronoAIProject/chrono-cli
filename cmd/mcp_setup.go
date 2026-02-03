package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ChronoAIProject/chrono-cli/pkg/api"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	mcpEditor string
	mcpToken  string
)

// mcpSetupCmd represents the mcp-setup command
var mcpSetupCmd = &cobra.Command{
	Use:   "mcp-setup [editor]",
	Short: "Configure AI editor to use Chrono as an MCP server",
	Long: `Configure your AI editor to use Chrono CLI as an MCP server.

Supports:
- Cursor IDE
- Claude Code
- Codex
- Gemini CLI

Arguments:
  editor  Optional. Skip prompt by specifying: cursor, claude-code, codex, gemini

Flags:
  --editor string    Specify editor directly (cursor, claude-code, codex, gemini)
  --token string      Use existing API token (skips login)`,
	RunE: runMCPSetup,
}

func init() {
	rootCmd.AddCommand(mcpSetupCmd)
	mcpSetupCmd.Flags().StringVar(&mcpEditor, "editor", "", "AI editor (cursor, claude-code, codex, gemini)")
	mcpSetupCmd.Flags().StringVar(&mcpToken, "token", "", "API token (skips login)")
}

func runMCPSetup(cmd *cobra.Command, args []string) error {
	cfg := GetConfig()

	fmt.Println("========================================")
	fmt.Println("Chrono CLI - AI Editor MCP Configuration")
	fmt.Println("========================================")
	fmt.Println()

	var token string
	var serverURL string

	// Check if token is provided via flag
	if mcpToken != "" {
		fmt.Println("Using provided API token...")
		token = mcpToken
		serverURL = cfg.MCP.ServerURL
		fmt.Printf("✓ Using token: %s...\n", token[:min(10, len(token))])
		fmt.Println()
	} else {
		// Check if user is logged in
		fmt.Println("Checking authentication...")
		if !cfg.IsLoggedIn() || cfg.IsTokenExpired() {
			fmt.Println("⚠️  Not logged in. Please authenticate:")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  1. Run 'chrono login' first")
			fmt.Println("  2. Use --token flag with existing API token")
			fmt.Println()
			return fmt.Errorf("not logged in. Run 'chrono login' first or use --token")
		}

		fmt.Printf("✓ Logged in as %s\n", cfg.Auth.Email)
		fmt.Println()

		// Create API token
		fmt.Println("Creating API token for MCP...")
		client := api.NewClient(cfg.MCP.ServerURL)
		client.SetAuthToken(cfg.Auth.AccessToken)

		tokenResp, err := client.CreateToken(&api.CreateAPITokenRequest{
			Name:      "AI Editor MCP",
			Scope:     "personal",
			ExpiresIn: 365 * 24 * 60 * 60, // 1 year
		})
		if err != nil {
			return fmt.Errorf("failed to create API token: %w", err)
		}

		token = tokenResp.Token
		serverURL = cfg.MCP.ServerURL
		fmt.Printf("✓ API Token created: %s...\n", token[:10])
		fmt.Println()
	}

	// Determine editor (from flag, arg, or prompt)
	editor := mcpEditor
	if editor == "" && len(args) > 0 {
		editor = args[0]
	}

	var selectionIdx int
	if editor != "" {
		// Map editor name to index
		selectionIdx = mapEditorToIndex(editor)
		if selectionIdx < 0 {
			return fmt.Errorf("unknown editor: %s. Valid options: cursor, claude-code, codex, gemini", editor)
		}
		fmt.Printf("✓ Editor: %s\n", getEditorName(selectionIdx))
		fmt.Println()
	} else {
		// Interactive prompt
		prompt := promptui.Select{
			Label: "Which AI editor are you using?",
			Items: []string{
				"Cursor IDE",
				"Claude Code",
				"Codex",
				"Gemini CLI",
			},
		}
		var err error
		selectionIdx, _, err = prompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		fmt.Println()
	}

	// Show configuration based on selection
	switch selectionIdx {
	case 0:
		showCursorConfig(serverURL, token)
	case 1:
		showClaudeCodeConfig(serverURL, token)
	case 2:
		showCodexConfig(serverURL, token)
	case 3:
		showGeminiConfig(serverURL, token)
	}

	// Show available tools
	showAvailableTools()

	return nil
}

func showCursorConfig(serverURL, token string) {
	fmt.Println("========================================")
	fmt.Println("Cursor IDE Configuration")
	fmt.Println("========================================")
	fmt.Println()

	cursorConfigDir := filepath.Join(getWDir(), ".cursor")
	cursorConfigPath := filepath.Join(cursorConfigDir, "mcp.json")

	if err := os.MkdirAll(cursorConfigDir, 0755); err != nil {
		fmt.Printf("⚠️  Failed to create .cursor directory: %v\n", err)
		fmt.Println("\nManual configuration:")
		printCursorMCPConfig(serverURL, token)
		return
	}

	// Use merge logic for consistency
	if err := mergeCursorConfig(cursorConfigPath, serverURL, token); err != nil {
		fmt.Printf("⚠️  Failed to write .cursor/mcp.json: %v\n", err)
		fmt.Println("\nManual configuration:")
		printCursorMCPConfig(serverURL, token)
		return
	}

	fmt.Println("✓ Created/updated .cursor/mcp.json")
	fmt.Println()

	// Verify MCP connection
	fmt.Println("Verifying MCP connection...")
	if err := testMCPConnection(serverURL, token); err != nil {
		fmt.Printf("⚠️  MCP connection test failed: %v\n", err)
		fmt.Println("  Please check your network and try again")
		fmt.Println()
	} else {
		fmt.Println("✓ MCP connection verified successfully")
		fmt.Println()
	}

	fmt.Println("Next steps:")
	fmt.Println("  1. Open Cursor")
	fmt.Println("  2. Wait for .cursor/mcp.json to be detected")
	fmt.Println("  3. MCP tools will be available")
	fmt.Println()
}

func getWDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func showClaudeCodeConfig(serverURL, token string) {
	fmt.Println("========================================")
	fmt.Println("Claude Code Configuration")
	fmt.Println("========================================")
	fmt.Println()

	configPath := filepath.Join(getWDir(), ".mcp.json")

	// Merge or create config
	if err := mergeMCPConfig(configPath, serverURL, token); err != nil {
		fmt.Printf("⚠️  Failed to create .mcp.json: %v\n", err)
		fmt.Println("\nManual configuration:")
		fmt.Println("Create/Edit: .mcp.json")
		fmt.Println()
		printMCPConfigJSON(serverURL, token)
		return
	}

	fmt.Println("✓ Created/updated .mcp.json")
	fmt.Println()

	// Verify MCP connection
	fmt.Println("Verifying MCP connection...")
	if err := testMCPConnection(serverURL, token); err != nil {
		fmt.Printf("⚠️  MCP connection test failed: %v\n", err)
		fmt.Println("  Please check your network and try again")
		fmt.Println()
	} else {
		fmt.Println("✓ MCP connection verified successfully")
		fmt.Println()
	}

	fmt.Println("Next steps:")
	fmt.Println("  1. Restart Claude Code")
	fmt.Println("  2. MCP tools will be available")
	fmt.Println()
}

func showCodexConfig(serverURL, token string) {
	fmt.Println("========================================")
	fmt.Println("Codex Configuration")
	fmt.Println("========================================")
	fmt.Println()

	codexDir := filepath.Join(getWDir(), ".codex")
	configPath := filepath.Join(codexDir, "mcp.json")

	// Create .codex directory if needed
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		fmt.Printf("⚠️  Failed to create .codex directory: %v\n", err)
		fmt.Println("\nManual configuration:")
		printMCPConfigJSON(serverURL, token)
		return
	}

	// Merge or create config
	if err := mergeMCPConfig(configPath, serverURL, token); err != nil {
		fmt.Printf("⚠️  Failed to create .codex/mcp.json: %v\n", err)
		fmt.Println("\nManual configuration:")
		printMCPConfigJSON(serverURL, token)
		return
	}

	fmt.Println("✓ Created/updated .codex/mcp.json")
	fmt.Println()

	// Verify MCP connection
	fmt.Println("Verifying MCP connection...")
	if err := testMCPConnection(serverURL, token); err != nil {
		fmt.Printf("⚠️  MCP connection test failed: %v\n", err)
		fmt.Println("  Please check your network and try again")
		fmt.Println()
	} else {
		fmt.Println("✓ MCP connection verified successfully")
		fmt.Println()
	}

	fmt.Println("Next steps:")
	fmt.Println("  1. Restart Codex")
	fmt.Println("  2. MCP tools will be available")
	fmt.Println()
}

func showGeminiConfig(serverURL, token string) {
	fmt.Println("========================================")
	fmt.Println("Gemini CLI Configuration")
	fmt.Println("========================================")
	fmt.Println()

	geminiDir := filepath.Join(getWDir(), ".gemini")
	configPath := filepath.Join(geminiDir, "settings.json")

	// Create .gemini directory if needed
	if err := os.MkdirAll(geminiDir, 0755); err != nil {
		fmt.Printf("⚠️  Failed to create .gemini directory: %v\n", err)
		fmt.Println("\nManual configuration:")
		printMCPConfigJSON(serverURL, token)
		return
	}

	// Merge or create config
	if err := mergeMCPConfig(configPath, serverURL, token); err != nil {
		fmt.Printf("⚠️  Failed to create .gemini/settings.json: %v\n", err)
		fmt.Println("\nManual configuration:")
		printMCPConfigJSON(serverURL, token)
		return
	}

	fmt.Println("✓ Created/updated .gemini/settings.json")
	fmt.Println()

	// Verify MCP connection
	fmt.Println("Verifying MCP connection...")
	if err := testMCPConnection(serverURL, token); err != nil {
		fmt.Printf("⚠️  MCP connection test failed: %v\n", err)
		fmt.Println("  Please check your network and try again")
		fmt.Println()
	} else {
		fmt.Println("✓ MCP connection verified successfully")
		fmt.Println()
	}

	fmt.Println("Next steps:")
	fmt.Println("  1. Restart Gemini CLI")
	fmt.Println("  2. MCP tools will be available")
	fmt.Println()
}

func showAvailableTools() {
	fmt.Println("========================================")
	fmt.Println("Available MCP Tools")
	fmt.Println("========================================")
	fmt.Println()
	tools := []string{
		"check_github_connection - Check GitHub OAuth status",
		"list_projects - List all projects",
		"create_project - Create a new project",
		"list_pipelines - List all pipelines",
		"create_pipeline - Create CI/CD pipeline",
		"trigger_pipeline_run - Trigger deployment",
		"get_run_status - Check deployment status",
		"restart_deployment - Rolling restart",
		"get_deployment_status - Get deployment status",
	}
	for _, tool := range tools {
		fmt.Printf("  • %s\n", tool)
	}
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("Next Steps")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("✓ MCP configuration has been automatically added to your editor")
	fmt.Println("✓ Restart your AI editor to start using the MCP tools")
	fmt.Println()
}

func printCursorMCPConfig(serverURL, token string) {
	fmt.Println("```json")
	fmt.Printf(`{
  "mcpServers": {
    "developer-platform": {
      "url": "%s/mcp",
      "headers": {
        "Authorization": "Bearer %s"
      }
    }
  }
}
`, serverURL, token)
	fmt.Println("```")
}

// mapEditorToIndex maps editor name/alias to selection index
// Returns -1 if not found
func mapEditorToIndex(editor string) int {
	// Normalize input
	editor = strings.ToLower(strings.ReplaceAll(editor, "-", ""))
	editor = strings.ReplaceAll(editor, "_", "")

	switch editor {
	case "cursor", "cursoride":
		return 0
	case "claude", "claudecode", "claude-code":
		return 1
	case "codex":
		return 2
	case "gemini", "gemicli", "gemini-cli":
		return 3
	default:
		return -1
	}
}

// getEditorName returns the display name for an editor index
func getEditorName(idx int) string {
	names := []string{
		"Cursor IDE",
		"Claude Code",
		"Codex",
		"Gemini CLI",
	}
	if idx >= 0 && idx < len(names) {
		return names[idx]
	}
	return "Unknown"
}

// mergeCursorConfig merges or creates the Cursor MCP config file
// Cursor format: { "mcpServers": { "server-name": { ... } } }
func mergeCursorConfig(configPath, serverURL, token string) error {
	type MCPConfig struct {
		MCPServers map[string]interface{} `json:"mcpServers"`
	}

	var config MCPConfig

	// Read existing config if it exists
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			// File exists but invalid JSON, start fresh
			config = MCPConfig{MCPServers: make(map[string]interface{})}
		}
	} else {
		// File doesn't exist, create new
		config = MCPConfig{MCPServers: make(map[string]interface{})}
	}

	// Ensure mcpServers map exists
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]interface{})
	}

	// Add or update the developer-platform server
	config.MCPServers["developer-platform"] = map[string]interface{}{
		"url": serverURL + "/mcp",
		"headers": map[string]string{
			"Authorization": "Bearer " + token,
		},
	}

	// Marshal back to JSON with proper indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// mergeMCPConfig merges or creates the MCP config file for Claude Code, Codex, Gemini
// Standard format: { "mcpServers": { "server-name": { ... } } }
func mergeMCPConfig(configPath, serverURL, token string) error {
	type MCPConfig struct {
		MCPServers map[string]interface{} `json:"mcpServers"`
	}

	var config MCPConfig

	// Read existing config if it exists
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			// File exists but invalid JSON, start fresh
			config = MCPConfig{MCPServers: make(map[string]interface{})}
		}
	} else {
		// File doesn't exist, create new
		config = MCPConfig{MCPServers: make(map[string]interface{})}
	}

	// Ensure mcpServers map exists
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]interface{})
	}

	// Add or update the developer-platform server
	config.MCPServers["developer-platform"] = map[string]interface{}{
		"url": serverURL + "/mcp",
		"headers": map[string]string{
			"Authorization": "Bearer " + token,
		},
	}

	// Marshal back to JSON with proper indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// printMCPConfigJSON prints the standard MCP config JSON format
func printMCPConfigJSON(serverURL, token string) {
	fmt.Println("```json")
	fmt.Printf(`{
  "mcpServers": {
    "developer-platform": {
      "url": "%s/mcp",
      "headers": {
        "Authorization": "Bearer %s"
      }
    }
  }
}
`, serverURL, token)
	fmt.Println("```")
}

// testMCPConnection tests the MCP connection to verify it works
func testMCPConnection(serverURL, token string) error {
	client := api.NewClient(serverURL)
	client.SetAPIToken(token)

	resp, err := client.GetMCPInfo()
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	fmt.Printf("  Connected to: %s (v%s)\n", resp.Name, resp.Version)
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
