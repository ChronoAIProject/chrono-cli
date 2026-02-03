package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ChronoAIProject/chrono-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	initAPIURL     string
	initGitIgnore   bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Chrono configuration for your project",
	Long: `Initialize Chrono CLI for your project.

Creates .chrono/config.yaml and .chrono/skills/ directory.

Your project should follow this structure:
  - backend/      (for backend code)
  - frontend/     (for frontend code)

The platform will auto-detect your project type based on these folders.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initAPIURL, "api-url", "", "API server URL (overrides env var)")
	initCmd.Flags().BoolVar(&initGitIgnore, "add-gitignore", true, "Add .chrono to .gitignore")
}

func runInit(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check if .chrono already exists
	chronoDir := filepath.Join(wd, ".chrono")
	if stat, err := os.Stat(chronoDir); err == nil {
		if stat.IsDir() {
			fmt.Printf("⚠️  .chrono folder already exists\n")
			fmt.Println("Delete .chrono to reinitialize.")
			return nil
		}
	}

	fmt.Println("========================================")
	fmt.Println("Chrono CLI - Setup")
	fmt.Println("========================================")
	fmt.Println()

	// Determine API URL: flag > env var > default
	apiURL := initAPIURL
	if apiURL == "" {
		apiURL = os.Getenv("CHRONO_API_URL")
	}
	if apiURL == "" {
		// Use default from config package
		apiURL = config.Default().MCP.ServerURL
	}

	if apiURL == "" {
		// Require explicit API URL configuration
		return fmt.Errorf("API URL required. Set CHRONO_API_URL environment variable or use --api-url flag")
	}

	// Create .chrono directory
	if err := os.MkdirAll(chronoDir, 0755); err != nil {
		return fmt.Errorf("failed to create .chrono directory: %w", err)
	}

	// Create skills directory
	skillsDir := filepath.Join(chronoDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	// Create config file
	configPath := filepath.Join(chronoDir, "config.yaml")
	configContent := fmt.Sprintf(`# Chrono CLI Configuration
# This file contains project-specific settings for Chrono CLI

mcp:
    server_url: %s
skills:
    install_dir: .chrono/skills
    installed: []
`, apiURL)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println("✓ Created .chrono/config.yaml")
	fmt.Println("✓ Created .chrono/skills/")
	fmt.Println()

	// Add to .gitignore if requested
	if initGitIgnore {
		addToGitIgnore(wd)
	}

	// Auto-install all skills
	fmt.Println("Installing skills...")
	installedSkills, err := installAllSkills(apiURL)
	if err != nil {
		fmt.Printf("⚠️  Skill installation failed: %v\n", err)
		fmt.Println("You can install skills later with: chrono skill install <name>")
	} else if len(installedSkills) > 0 {
		// Update config with installed skills
		updateInstalledSkills(configPath, installedSkills, apiURL)
	}

	printConfigInstructions()
	return nil
}

func installAllSkills(serverURL string) ([]string, error) {
	// Skills to install from GitHub (folder names that contain SKILL.md)
	skills := []string{
		"chrono-setup",
		"chrono-storage",
		"chrono-check-setup",
		"chrono-deploy",
		"chrono-restart",
	}

	installed := make([]string, 0, len(skills))
	wd := wD()

	// Define install location type
	type installLocation struct {
		Path        string
		Description string
	}

	// Install to multiple locations
	installLocations := []installLocation{
		{
			Path:        filepath.Join(wd, ".chrono", "skills"),
			Description: ".chrono/skills/",
		},
	}

	// Always install to .cursor/skills/ (most common AI editor)
	cursorSkillsPath := filepath.Join(wd, ".cursor", "skills")
	if err := os.MkdirAll(cursorSkillsPath, 0755); err == nil {
		installLocations = append(installLocations, installLocation{
			Path:        cursorSkillsPath,
			Description: ".cursor/skills/",
		})
		fmt.Println("✓ Installing skills to .cursor/skills/")
	} else {
		fmt.Printf("⚠️  Failed to create .cursor/skills/: %v\n", err)
	}

	// Check for other AI editor folders and add them
	// We check if the parent folder exists and create skills/ subdirectory
	aiEditorFolders := []struct {
		parent     string
		skillsPath string
	}{
		{".claude", ".claude/skills"},
		{".codex", ".codex/skills"},
		{".gemini", ".gemini/skills"},
	}

	for _, folder := range aiEditorFolders {
		// Check if parent folder exists
		parentPath := filepath.Join(wd, folder.parent)
		if stat, err := os.Stat(parentPath); err == nil && stat.IsDir() {
			// Create skills/ subdirectory if it doesn't exist
			skillsPath := filepath.Join(wd, folder.skillsPath)
			if err := os.MkdirAll(skillsPath, 0755); err == nil {
				installLocations = append(installLocations, installLocation{
					Path:        skillsPath,
					Description: folder.skillsPath,
				})
				fmt.Printf("✓ Found %s - installing skills to %s\n", folder.parent, folder.skillsPath)
			} else if !os.IsNotExist(err) {
				fmt.Printf("⚠️  Found %s but couldn't create %s: %v\n", folder.parent, folder.skillsPath, err)
			}
		}
	}

	if len(installLocations) == 1 {
		fmt.Println("Installing skills to .chrono/skills/")
	} else {
		fmt.Printf("Installing skills to %d location(s)\n", len(installLocations))
	}

	// GitHub configuration
	githubRepo := "ChronoAIProject/chrono-cli"
	githubRef := "main"

	// Download and install each skill
	for _, skillName := range skills {
		var content []byte
		var downloadSource string

		// Build raw GitHub URL - skills are now in subdirectories with SKILL.md
		githubURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/skills/%s/SKILL.md",
			githubRepo, githubRef, skillName)

		fmt.Printf("  Downloading %s...", skillName)

		// Try to fetch from GitHub first
		resp, err := http.Get(githubURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			// Fallback: try to copy from local skills directory (for development)
			if resp != nil {
				resp.Body.Close()
			}

			// Try local file (for chrono.md during development)
			localSkillPath := filepath.Join(filepath.Dir(wD()), "skills", skillName, "SKILL.md")
			if localContent, err := os.ReadFile(localSkillPath); err == nil {
				content = localContent
				downloadSource = "local"
				fmt.Print(" (local)...")
			} else {
				// Try from backend API
				apiURL := serverURL + "/skills/" + skillName
				if apiResp, apiErr := http.Get(apiURL); apiErr == nil {
					defer apiResp.Body.Close()
					if apiResp.StatusCode == http.StatusOK {
						if apiContent, readErr := io.ReadAll(apiResp.Body); readErr == nil {
							content = apiContent
							downloadSource = "API"
							fmt.Print(" (API)...")
						}
					}
				}

				if len(content) == 0 {
					if err != nil {
						fmt.Printf(" ✗ failed: %v\n", err)
					} else {
						fmt.Printf(" ✗ HTTP %d\n", resp.StatusCode)
					}
					continue
				}
			}
		} else {
			defer resp.Body.Close()

			content, err = io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf(" ✗ failed: %v\n", err)
				continue
			}
			downloadSource = "GitHub"
		}

		// Install to all locations - save in subdirectory structure matching GitHub
		success := true
		for _, loc := range installLocations {
			// Create skill subdirectory
			skillDir := filepath.Join(loc.Path, skillName)
			if err := os.MkdirAll(skillDir, 0755); err != nil {
				fmt.Printf(" ✗ %s: %v\n", loc.Description, err)
				success = false
				continue
			}

			skillPath := filepath.Join(skillDir, "SKILL.md")
			if err := os.WriteFile(skillPath, content, 0644); err != nil {
				fmt.Printf(" ✗ %s: %v\n", loc.Description, err)
				success = false
			}
		}

		if success {
			installed = append(installed, skillName)
			if downloadSource != "" {
				fmt.Printf(" ✓ (from %s)\n", downloadSource)
			} else {
				fmt.Println(" ✓")
			}
		} else {
			fmt.Println(" ✗")
		}
	}

	return installed, nil
}

func updateInstalledSkills(configPath string, skills []string, serverURL string) error {
	// Create installed skills YAML
	installedYAML := "installed:\n"
	for _, skill := range skills {
		installedYAML += fmt.Sprintf("  - name: %s\n", skill)
		installedYAML += fmt.Sprintf("    url: %s/skills/%s/SKILL.md\n", serverURL, skill)
		installedYAML += fmt.Sprintf("    installed_at: \"%s\"\n", time.Now().Format(time.RFC3339))
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Replace the installed section
	newContent := string(data)
	oldInstalled := "installed: []"
	if strings.Contains(newContent, oldInstalled) {
		newContent = strings.Replace(newContent, oldInstalled, installedYAML, 1)
	} else {
		// If the pattern doesn't match, append to the file
		newContent = strings.TrimSuffix(newContent, "\n")
		newContent += "\n" + installedYAML
	}

	// Write back
	if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}

func wD() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func addToGitIgnore(wd string) {
	gitignorePath := filepath.Join(wd, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("ℹ️  Could not update .gitignore")
		return
	}

	gitignoreContent := string(content)
	var toAdd []string

	// Check what needs to be added
	if !strings.Contains(gitignoreContent, ".chrono/") {
		toAdd = append(toAdd, ".chrono/")
	}
	if !strings.Contains(gitignoreContent, ".cursor/skills/") {
		toAdd = append(toAdd, ".cursor/skills/")
	}

	if len(toAdd) == 0 {
		fmt.Println("ℹ️  .chrono/ and .cursor/skills/ already in .gitignore")
		return
	}

	var addContent string
	if gitignoreContent == "" {
		addContent = "# Chrono CLI\n"
	} else {
		addContent = "\n# Chrono CLI\n"
	}
	for _, item := range toAdd {
		addContent += item + "\n"
	}

	if err := os.WriteFile(gitignorePath, append([]byte(content), addContent...), 0644); err == nil {
		fmt.Printf("✓ Added %s to .gitignore\n", strings.Join(toAdd, ", "))
	}
}

func printConfigInstructions() {
	fmt.Println("========================================")
	fmt.Println("Setup Complete!")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Created:")
	fmt.Println("  .chrono/config.yaml")
	fmt.Println("  .chrono/skills/")
	fmt.Println()
}
