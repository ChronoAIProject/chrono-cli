// Package template provides project scaffolding functionality for Chrono CLI
package template

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v3"
)

// ProjectType represents the type of project
type ProjectType string

const (
	ProjectTypeFrontend ProjectType = "frontend"
	ProjectTypeBackend  ProjectType = "backend"
	ProjectTypeFullstack ProjectType = "fullstack"
)

// Template represents a project template
type Template struct {
	DirectoryName string                `yaml:"-"` // Directory name (not in YAML)
	Name          string                `yaml:"name"`
	Type          ProjectType           `yaml:"type"`
	Description   string                `yaml:"description"`
	Variables     []Variable            `yaml:"variables"`
	Files         []FileTemplate        `yaml:"files"`
	Commands      []PostInitCommand     `yaml:"commands,omitempty"`
}

// Variable represents a template variable
type Variable struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Default     string   `yaml:"default,omitempty"`
	Required    bool     `yaml:"required"`
	Options     []string `yaml:"options,omitempty"`
	Type        string   `yaml:"type"` // string, select, bool
}

// FileTemplate represents a file to be generated
type FileTemplate struct {
	Path      string `yaml:"path"`
	Content   string `yaml:"content"`
	Template  string `yaml:"template"`
	Condition string `yaml:"condition,omitempty"` // Go template condition
}

// PostInitCommand represents a command to run after initialization
type PostInitCommand struct {
	Run      string `yaml:"run"`
	Message  string `yaml:"message,omitempty"`
	Optional bool   `yaml:"optional,omitempty"`
}

// Engine handles template processing
type Engine struct {
	templateDir string
}

// NewEngine creates a new template engine
func NewEngine(templateDir string) *Engine {
	return &Engine{
		templateDir: templateDir,
	}
}

// ListTemplates returns all available templates
func (e *Engine) ListTemplates() ([]*Template, error) {
	var templates []*Template

	// Read template directories
	entries, err := os.ReadDir(e.templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		templatePath := filepath.Join(e.templateDir, entry.Name(), "template.yaml")
		content, err := os.ReadFile(templatePath)
		if err != nil {
			continue // Skip if no template.yaml
		}

		var tmpl Template
		if err := yaml.Unmarshal(content, &tmpl); err != nil {
			continue // Skip invalid templates
		}

		// Set the directory name (used for loading)
		tmpl.DirectoryName = entry.Name()

		templates = append(templates, &tmpl)
	}

	return templates, nil
}

// LoadTemplate loads a specific template by name
func (e *Engine) LoadTemplate(name string) (*Template, error) {
	templatePath := filepath.Join(e.templateDir, name, "template.yaml")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s: %w", name, err)
	}

	var tmpl Template
	if err := yaml.Unmarshal(content, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	return &tmpl, nil
}

// LoadTemplateFromFile loads a template from a specific file path
func LoadTemplateFromFile(path string) (*Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var tmpl Template
	if err := yaml.Unmarshal(content, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &tmpl, nil
}

// PromptUser prompts the user for variable values
func (e *Engine) PromptUser(tmpl *Template) (map[string]string, error) {
	values := make(map[string]string)

	fmt.Printf("\n=== %s ===\n", tmpl.Name)
	fmt.Println(tmpl.Description)
	fmt.Println()

	for _, v := range tmpl.Variables {
		var valueStr string

		switch v.Type {
		case "select":
			prompt := promptui.Select{
				Label: v.Description,
				Items: v.Options,
			}
			_, value, err := prompt.Run()
			if err != nil {
				return nil, fmt.Errorf("prompt failed: %w", err)
			}
			valueStr = value

		case "bool":
			prompt := promptui.Select{
				Label: v.Description,
				Items: []string{"Yes", "No"},
			}
			_, choice, err := prompt.Run()
			if err != nil {
				return nil, fmt.Errorf("prompt failed: %w", err)
			}
			valueStr = choice

		default: // string
			prompt := promptui.Prompt{
				Label:   v.Description,
				Default: v.Default,
			}
			value, err := prompt.Run()
			if err != nil {
				return nil, fmt.Errorf("prompt failed: %w", err)
			}
			valueStr = value
		}

		// Validate required fields
		if v.Required && valueStr == "" {
			return nil, fmt.Errorf("%s is required", v.Name)
		}

		values[v.Name] = valueStr
		fmt.Println()
	}

	return values, nil
}

// Generate creates project files from template
func (e *Engine) Generate(tmpl *Template, values map[string]string, targetDir string) error {
	// Add template metadata to values
	values["TemplateName"] = tmpl.Name
	values["TemplateType"] = string(tmpl.Type)

	// Create functions for templates
	funcMap := template.FuncMap{
		"toLower": strings.ToLower,
		"toUpper": strings.ToUpper,
		"title":   strings.Title,
		"replace": strings.ReplaceAll,
	}

	for _, file := range tmpl.Files {
		// Evaluate condition if present
		if file.Condition != "" {
			condTmpl, err := template.New("condition").Parse(file.Condition)
			if err != nil {
				return fmt.Errorf("invalid condition in file %s: %w", file.Path, err)
			}

			var condBuf bytes.Buffer
			if err := condTmpl.Execute(&condBuf, values); err != nil {
				return fmt.Errorf("failed to evaluate condition for file %s: %w", file.Path, err)
			}

			// Skip if condition is false/empty
			if strings.TrimSpace(condBuf.String()) != "true" {
				continue
			}
		}

		// Resolve file path
		pathTmpl, err := template.New("path").Funcs(funcMap).Parse(file.Path)
		if err != nil {
			return fmt.Errorf("invalid path template for file %s: %w", file.Path, err)
		}

		var pathBuf bytes.Buffer
		if err := pathTmpl.Execute(&pathBuf, values); err != nil {
			return fmt.Errorf("failed to resolve path for file %s: %w", file.Path, err)
		}

		filePath := filepath.Join(targetDir, pathBuf.String())

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
		}

		// Parse file content
		contentTmpl, err := template.New("content").Funcs(funcMap).Parse(file.Template)
		if err != nil {
			return fmt.Errorf("invalid content template for file %s: %w", file.Path, err)
		}

		var contentBuf bytes.Buffer
		if err := contentTmpl.Execute(&contentBuf, values); err != nil {
			return fmt.Errorf("failed to generate content for file %s: %w", file.Path, err)
		}

		// Write file
		if err := os.WriteFile(filePath, contentBuf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}

		fmt.Printf("✓ Created %s\n", filePath)
	}

	return nil
}

// RunPostInitCommands runs post-initialization commands
func (e *Engine) RunPostInitCommands(tmpl *Template, targetDir string) error {
	if len(tmpl.Commands) == 0 {
		return nil
	}

	fmt.Println()
	fmt.Println("Running post-init commands...")

	for _, cmd := range tmpl.Commands {
		if cmd.Message != "" {
			fmt.Printf("\n%s\n", cmd.Message)
		}

		// For now, just display the command
		// In production, you might want to execute these
		fmt.Printf("  $ %s\n", cmd.Run)

		// TODO: Actually execute commands in targetDir
		// This would require careful handling of shell commands
	}

	return nil
}
