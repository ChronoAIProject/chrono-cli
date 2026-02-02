package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/aelfdevops/chrono/pkg/api"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate via Keycloak device flow",
	Long: `Authenticate with the Developer Platform using Keycloak device flow.

This will open your browser or provide a code to enter for authentication.`,
	RunE: runLogin,
}

var openBrowser bool

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().BoolVar(&openBrowser, "browser", true, "automatically open browser for authentication")
}

func runLogin(cmd *cobra.Command, args []string) error {
	cfg := GetConfig()

	// Check if already logged in
	if cfg.IsLoggedIn() && !cfg.IsTokenExpired() {
		fmt.Printf("Already logged in as %s\n", cfg.Auth.Email)
		fmt.Println("Use 'chrono logout' to logout first if you want to re-authenticate.")
		return nil
	}

	// Create API client
	client := api.NewClient(cfg.MCP.ServerURL)

	// Start device flow
	fmt.Println("Initiating authentication...")
	fmt.Println()

	startResp, err := client.StartDeviceFlow()
	if err != nil {
		return fmt.Errorf("failed to start device flow: %w", err)
	}

	// Display user code and verification URL
	fmt.Println(strings.Repeat("─", 52))
	fmt.Println("  Authentication Required")
	fmt.Println(strings.Repeat("─", 52))
	fmt.Println()
	fmt.Printf("Enter this code: \033[1;37m\033[1;44m %s \033[0m\n", formatUserCode(startResp.UserCode))
	fmt.Println()
	fmt.Println("Then visit:")
	fmt.Printf("\033[4m%s\033[0m\n", startResp.VerificationURI)
	fmt.Println()
	fmt.Println(strings.Repeat("─", 52))
	fmt.Println()

	// Optionally open browser
	if openBrowser {
		fmt.Println("Opening browser...")
		// Try to open the browser with the complete verification URI
		browserCmd := exec.Command("open", startResp.VerificationURIComplete)
		if err := browserCmd.Start(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Could not open browser automatically: %v\n", err)
		}
	} else {
		fmt.Println("Please open the URL above in your browser manually.")
	}
	fmt.Println()

	// Poll for completion
	fmt.Println("Waiting for authentication...")
	pollInterval := time.Duration(startResp.Interval) * time.Second
	if pollInterval < 5*time.Second {
		pollInterval = 5 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	timeout := time.After(time.Duration(startResp.ExpiresIn) * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("authentication timed out. Please try again.")
		case <-ticker.C:
			pollResp, err := client.PollDeviceFlow(startResp.DeviceCode)
			if err != nil {
				// Check if it's still pending
				if pollResp != nil && pollResp.Status == "authorization_pending" {
					fmt.Print(".")
					continue
				}
				return fmt.Errorf("failed to poll device flow: %w", err)
			}

			// Check status
			if pollResp.Status == "authorization_pending" || pollResp.Status == "slow_down" {
				fmt.Print(".")
				if pollResp.Status == "slow_down" {
					ticker.Reset(pollInterval + 5*time.Second)
				}
				continue
			}

			// Success!
			fmt.Println()

			// Save credentials to config
			cfg.Auth.AccessToken = pollResp.AccessToken
			cfg.Auth.TokenExpiry = time.Now().Add(time.Duration(pollResp.ExpiresIn) * time.Second)
			cfg.Auth.UserID = pollResp.User.ID
			cfg.Auth.Email = pollResp.User.Email

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			// Show success message
			fmt.Println()
			fmt.Println("✓ Successfully authenticated!")
			fmt.Printf("  Logged in as: \033[1m%s\033[0m\n", pollResp.User.Email)
			fmt.Printf("  Role: %s\n", pollResp.User.Role)
			fmt.Println()

			// Show next steps
			fmt.Println("Next Steps:")
			fmt.Println("  chrono mcp-setup     # Configure AI editor")
			fmt.Println("  chrono detect --save # Analyze project (optional)")
			fmt.Println()

			return nil
		}
	}
}

// formatUserCode formats the user code for display (XXXX-XXXX)
func formatUserCode(code string) string {
	if len(code) == 8 {
		return code[:4] + "-" + code[4:]
	}
	return code
}

// ============================================
// Logout Command
// ============================================

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear local credentials",
	Long:  `Logout from the Developer Platform and clear stored credentials.`,
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	cfg := GetConfig()

	// Check if logged in
	if !cfg.IsLoggedIn() {
		fmt.Println("Not logged in.")
		return nil
	}

	fmt.Printf("Logging out %s...\n", cfg.Auth.Email)

	// Clear credentials
	email := cfg.Auth.Email
	cfg.Clear()

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to clear credentials: %w", err)
	}

	fmt.Printf("✓ Logged out successfully\n")
	fmt.Printf("  Goodbye, %s!\n", email)
	fmt.Println()

	return nil
}

// ============================================
// Status Command
// ============================================

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current login status",
	Long:  `Show the current authentication status and user information.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg := GetConfig()

	fmt.Println("Chrono CLI Status")
	fmt.Println(strings.Repeat("─", 42))

	if !cfg.IsLoggedIn() {
		fmt.Println("Status: \033[33mNot logged in\033[0m")
		fmt.Println()
		fmt.Println("Use 'chrono login' to authenticate.")
		return nil
	}

	if cfg.IsTokenExpired() {
		fmt.Println("Status: \033[33mSession expired\033[0m")
		fmt.Printf("  Email: %s\n", cfg.Auth.Email)
		fmt.Println()
		fmt.Println("Use 'chrono login' to re-authenticate.")
		return nil
	}

	fmt.Println("Status: \033[32mLogged in\033[0m")
	fmt.Printf("  Email: %s\n", cfg.Auth.Email)
	fmt.Printf("  User ID: %s\n", cfg.Auth.UserID)
	if !cfg.Auth.TokenExpiry.IsZero() {
		timeUntilExpiry := time.Until(cfg.Auth.TokenExpiry)
		if timeUntilExpiry > 0 {
			fmt.Printf("  Token expires in: %s\n", timeUntilExpiry.Round(time.Minute))
		} else {
			fmt.Printf("  Token expired: %s\n", cfg.Auth.TokenExpiry.Format(time.RFC3339))
		}
	}
	fmt.Println()

	return nil
}
