package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/aelfdevops/chrono/pkg/detector"
	"gopkg.in/yaml.v3"
)

var (
	detectJSON     bool
	detectSave     bool
	detectVerbose  bool
)

// detectCmd represents the detect command
var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Analyze project and detect tech stack",
	Long: `Detect project type, tech stack, and configuration.

This analyzes your current directory to identify:
- Project type (frontend, backend, fullstack)
- Frameworks and languages
- Dockerfile status
- Middleware dependencies (MongoDB, Redis, etc.)
- Environment variables

Use this to see how the platform will interpret your project.`,
	RunE: runDetect,
}

func init() {
	rootCmd.AddCommand(detectCmd)
	detectCmd.Flags().BoolVar(&detectJSON, "json", false, "Output as JSON")
	detectCmd.Flags().BoolVar(&detectSave, "save", false, "Save detection results to .chrono/metadata.yaml")
	detectCmd.Flags().BoolVar(&detectVerbose, "verbose", false, "Show detailed information")
}

func runDetect(cmd *cobra.Command, args []string) error {
	// Get current directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	fmt.Println("========================================")
	fmt.Println("Project Detection")
	fmt.Println("========================================")
	fmt.Printf("Scanning: %s\n\n", wd)

	// Run detector
	d := detector.NewDetector(wd)
	metadata, err := d.Detect()
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}

	// Display results
	if detectJSON {
		return outputJSON(metadata)
	}

	displayResults(metadata, detectVerbose)

	// Save to file if requested
	if detectSave {
		if err := saveMetadata(wd, metadata); err != nil {
			return fmt.Errorf("failed to save metadata: %w", err)
		}
		fmt.Println("\n✓ Saved to .chrono/metadata.yaml")
	}

	return nil
}

func displayResults(metadata *detector.Metadata, verbose bool) {
	// Project type
	typeIcon := map[detector.ProjectType]string{
		detector.ProjectTypeFrontend: "🎨",
		detector.ProjectTypeBackend:  "⚙️",
		detector.ProjectTypeFullstack: "🎨⚙️",
		detector.ProjectTypeUnknown:  "❓",
	}

	fmt.Printf("%s Project Type: %s\n", typeIcon[metadata.Project.Type], metadata.Project.Type)
	fmt.Printf("   Name: %s\n", metadata.Project.Name)
	fmt.Println()

	// Frontend
	if metadata.TechStack.Frontend != nil {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("Frontend")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		f := metadata.TechStack.Frontend
		fmt.Printf("  Framework: %s (%s)\n", f.Framework, f.Language)
		fmt.Printf("  Port:      %d\n", f.Port)
		fmt.Printf("  Build:     %s\n", f.BuildCmd)
		// Frontend SPAs don't need Dockerfile - platform uses nginx
		fmt.Printf("  Serving:   Static files via nginx (no Dockerfile needed)\n")
		if verbose && len(f.EnvVars) > 0 {
			fmt.Println("  Env Vars:")
			for key := range f.EnvVars {
				fmt.Printf("    - %s\n", key)
			}
		}
		fmt.Println()
	}

	// Backend
	if metadata.TechStack.Backend != nil {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("Backend")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		b := metadata.TechStack.Backend
		fmt.Printf("  Framework: %s (%s)\n", b.Framework, b.Language)
		if b.Version != "" {
			fmt.Printf("  Version:   %s\n", b.Version)
		}
		fmt.Printf("  Port:      %d\n", b.Port)
		fmt.Printf("  Docker:    ")
		if b.HasDockerfile {
			fmt.Printf("✓ %s\n", b.Dockerfile)
		} else {
			fmt.Println("✗ Missing (will be generated during deployment)")
		}
		if verbose {
			fmt.Printf("  Build:     %s\n", b.BuildCmd)
			fmt.Printf("  Start:     %s\n", b.StartCmd)
			if len(b.EnvVars) > 0 {
				fmt.Println("  Env Vars:")
				for key := range b.EnvVars {
					fmt.Printf("    - %s\n", key)
				}
			}
		}
		fmt.Println()
	}

	// Middleware
	middleware := []string{}
	if metadata.Middleware.MongoDB {
		middleware = append(middleware, "MongoDB")
	}
	if metadata.Middleware.Redis {
		middleware = append(middleware, "Redis")
	}
	if metadata.Middleware.Postgres {
		middleware = append(middleware, "PostgreSQL")
	}
	if metadata.Middleware.MySQL {
		middleware = append(middleware, "MySQL")
	}

	if len(middleware) > 0 {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("Middleware Detected")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for _, m := range middleware {
			fmt.Printf("  • %s\n", m)
		}
		fmt.Println()
	}

	// Status
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Only check backend Dockerfile - frontend uses nginx (no Dockerfile needed)
	readyToDeploy := true
	if metadata.TechStack.Backend != nil && !metadata.TechStack.Backend.HasDockerfile {
		readyToDeploy = false
	}

	if readyToDeploy {
		fmt.Println("  ✓ Ready to deploy")
		if metadata.TechStack.Backend != nil {
			fmt.Println("    Backend Dockerfile found")
		}
		if metadata.TechStack.Frontend != nil {
			fmt.Println("    Frontend will be built as static files")
		}
	} else {
		fmt.Println("  ⚠ Backend Dockerfile missing - will be generated during deployment")
	}

	if len(middleware) > 0 {
		fmt.Println("  ✓ Middleware detected - will be provisioned")
	}

	fmt.Println()

	// Next steps
	fmt.Println("Next Steps:")
	fmt.Println("  chrono detect --save    # Save metadata for deployment")
	fmt.Println("  chrono login            # Authenticate with platform")
	fmt.Println("  chrono mcp-setup        # Configure AI editor (Cursor, Claude Code)")
	fmt.Println("  # Then use /deploy in your AI editor")
}

func outputJSON(metadata *detector.Metadata) error {
	// Use json.Marshal for actual JSON output
	// For now, just output YAML for simplicity
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func saveMetadata(wd string, metadata *detector.Metadata) error {
	// Create .chrono directory
	chronoDir := filepath.Join(wd, ".chrono")
	if err := os.MkdirAll(chronoDir, 0755); err != nil {
		return err
	}

	// Add platform info
	metadata.Platform.CreatedBy = "chrono_cli"
	metadata.Project.DetectedAt = time.Now().Format(time.RFC3339)

	// Marshal to YAML
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	// Write to file
	metadataPath := filepath.Join(chronoDir, "metadata.yaml")
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return err
	}

	return nil
}
