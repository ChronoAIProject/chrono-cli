// Package detector provides project type and tech stack detection
package detector

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectType represents the type of project
type ProjectType string

const (
	ProjectTypeFrontend ProjectType = "frontend"
	ProjectTypeBackend  ProjectType = "backend"
	ProjectTypeFullstack ProjectType = "fullstack"
	ProjectTypeUnknown  ProjectType = "unknown"
)

// TechStack represents detected technology stack
type TechStack struct {
	Framework    string            `yaml:"framework" json:"framework"`
	Language     string            `yaml:"language" json:"language"`
	Version      string            `yaml:"version,omitempty" json:"version,omitempty"`
	Port         int               `yaml:"port,omitempty" json:"port,omitempty"`
	HasDockerfile bool             `yaml:"has_dockerfile" json:"has_dockerfile"`
	Dockerfile   string            `yaml:"dockerfile_path,omitempty" json:"dockerfile_path,omitempty"`
	BuildCmd     string            `yaml:"build_command,omitempty" json:"build_command,omitempty"`
	StartCmd     string            `yaml:"start_command,omitempty" json:"start_command,omitempty"`
	EnvVars      map[string]string `yaml:"env_vars,omitempty" json:"env_vars,omitempty"`
}

// Middleware represents detected middleware dependencies
type Middleware struct {
	MongoDB bool `yaml:"mongodb" json:"mongodb"`
	Redis   bool `yaml:"redis" json:"redis"`
	Postgres bool `yaml:"postgres" json:"postgres"`
	MySQL   bool `yaml:"mysql" json:"mysql"`
}

// Metadata represents complete project metadata
type Metadata struct {
	Project    ProjectMetadata `yaml:"project" json:"project"`
	TechStack  TechStackMap    `yaml:"tech_stack" json:"tech_stack"`
	Middleware Middleware      `yaml:"middleware" json:"middleware"`
	Platform   PlatformInfo    `yaml:"platform" json:"platform"`
}

// ProjectMetadata contains basic project info
type ProjectMetadata struct {
	Name       string    `yaml:"name" json:"name"`
	Type       ProjectType `yaml:"type" json:"type"`
	DetectedAt string    `yaml:"detected_at" json:"detected_at"`
}

// TechStackMap contains frontend and/or backend stacks
type TechStackMap struct {
	Frontend *TechStack `yaml:"frontend,omitempty" json:"frontend,omitempty"`
	Backend  *TechStack `yaml:"backend,omitempty" json:"backend,omitempty"`
}

// PlatformInfo contains platform-specific info
type PlatformInfo struct {
	CreatedBy string `yaml:"created_by,omitempty" json:"created_by,omitempty"`
	Template  string `yaml:"template,omitempty" json:"template,omitempty"`
}

// Detector analyzes project directories
type Detector struct {
	rootDir string
}

// NewDetector creates a new detector for the given directory
func NewDetector(rootDir string) *Detector {
	return &Detector{rootDir: rootDir}
}

// Detect analyzes the project and returns metadata
func (d *Detector) Detect() (*Metadata, error) {
	projectName := filepath.Base(d.rootDir)

	// First, try to detect at root level
	frontend := d.detectFrontend()
	backend := d.detectBackend()

	// If nothing found at root, check for monorepo structure
	if frontend == nil && backend == nil {
		frontend, backend = d.detectMonorepo()
	}

	middleware := d.detectMiddleware()

	var projType ProjectType
	switch {
	case frontend != nil && backend != nil:
		projType = ProjectTypeFullstack
	case frontend != nil:
		projType = ProjectTypeFrontend
	case backend != nil:
		projType = ProjectTypeBackend
	default:
		projType = ProjectTypeUnknown
	}

	return &Metadata{
		Project: ProjectMetadata{
			Name:       projectName,
			Type:       projType,
			DetectedAt: timestamp(),
		},
		TechStack: TechStackMap{
			Frontend: frontend,
			Backend:  backend,
		},
		Middleware: middleware,
	}, nil
}

// detectMonorepo checks for common monorepo directory structures
func (d *Detector) detectMonorepo() (frontend, backend *TechStack) {
	// Common monorepo directory patterns
	frontendDirs := []string{"frontend", "web", "client", "ui"}
	backendDirs := []string{"backend", "api", "server", "services"}

	// Check frontend directories
	for _, dir := range frontendDirs {
		dirPath := filepath.Join(d.rootDir, dir)
		if d.fileExists(dirPath) {
			subDetector := &Detector{rootDir: dirPath}
			if f := subDetector.detectFrontend(); f != nil {
				f.Dockerfile = filepath.Join(dir, f.Dockerfile)
				frontend = f
				break
			}
		}
	}

	// Check backend directories
	for _, dir := range backendDirs {
		dirPath := filepath.Join(d.rootDir, dir)
		if d.fileExists(dirPath) {
			subDetector := &Detector{rootDir: dirPath}
			if b := subDetector.detectBackend(); b != nil {
				b.Dockerfile = filepath.Join(dir, b.Dockerfile)
				backend = b
				break
			}
		}
	}

	return frontend, backend
}

// detectFrontend detects frontend framework and configuration
func (d *Detector) detectFrontend() *TechStack {
	// Check for package.json
	pkgPath := filepath.Join(d.rootDir, "package.json")
	if _, err := os.Stat(pkgPath); err != nil {
		return nil
	}

	content, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}

	deps := d.parseDependencies(content)

	// Detect framework
	var framework, language string
	var port int = 3000 // default

	switch {
	case d.hasDep(deps, "next"):
		framework = "nextjs"
		port = 3000
	case d.hasDep(deps, "react") || d.hasDep(deps, "react-dom"):
		framework = "react"
		port = 3000
	case d.hasDep(deps, "vue"):
		framework = "vue"
		port = 5173
	case d.hasDep(deps, "vite"):
		framework = "vite"
		port = 5173
	case d.hasDep(deps, "svelte"):
		framework = "svelte"
		port = 5173
	default:
		return nil
	}

	// Check for TypeScript
	language = "javascript"
	if d.hasDep(deps, "typescript") || d.hasDep(deps, "@types/react") || d.hasDep(deps, "@types/node") {
		language = "typescript"
	}

	// Find Dockerfile
	dockerfile, hasDocker := d.findDockerfile([]string{
		"frontend/Dockerfile",
		"Dockerfile",
		"Dockerfile.frontend",
	})

	// Detect environment variables
	envVars := d.detectEnvVars([]string{
		"frontend/.env",
		"frontend/.env.local",
		".env",
		".env.local",
	})

	return &TechStack{
		Framework:    framework,
		Language:     language,
		Port:         port,
		HasDockerfile: hasDocker,
		Dockerfile:   dockerfile,
		BuildCmd:     "npm run build",
		StartCmd:     "npm start",
		EnvVars:      envVars,
	}
}

// detectBackend detects backend framework and configuration
func (d *Detector) detectBackend() *TechStack {
	// Check for Go
	if goMod := filepath.Join(d.rootDir, "go.mod"); d.fileExists(goMod) {
		return d.detectGoBackend()
	}

	// Check for Python
	if reqTxt := filepath.Join(d.rootDir, "requirements.txt"); d.fileExists(reqTxt) {
		return d.detectPythonBackend()
	}
	if pyProject := filepath.Join(d.rootDir, "pyproject.toml"); d.fileExists(pyProject) {
		return d.detectPythonBackend()
	}

	// Check for Node.js backend
	pkgPath := filepath.Join(d.rootDir, "package.json")
	if content, err := os.ReadFile(pkgPath); err == nil {
		deps := d.parseDependencies(content)
		if d.hasDep(deps, "express") || d.hasDep(deps, "fastify") || d.hasDep(deps, "koa") {
			return d.detectNodeBackend(deps)
		}
	}

	return nil
}

// detectGoBackend detects Go backend configuration
func (d *Detector) detectGoBackend() *TechStack {
	// Find Dockerfile
	dockerfile, hasDocker := d.findDockerfile([]string{
		"backend/Dockerfile",
		"Dockerfile",
		"Dockerfile.backend",
	})

	// Detect middleware
	envVars := d.detectEnvVars([]string{
		"backend/.env",
		".env",
	})

	// Default port for Go is 8080
	port := 8080

	return &TechStack{
		Framework:    "gin",
		Language:     "go",
		Version:      "1.22",
		Port:         port,
		HasDockerfile: hasDocker,
		Dockerfile:   dockerfile,
		BuildCmd:     "go build -o bin/server ./cmd/server",
		StartCmd:     "./bin/server",
		EnvVars:      envVars,
	}
}

// detectPythonBackend detects Python backend configuration
func (d *Detector) detectPythonBackend() *TechStack {
	dockerfile, hasDocker := d.findDockerfile([]string{
		"backend/Dockerfile",
		"Dockerfile",
	})

	envVars := d.detectEnvVars([]string{
		"backend/.env",
		".env",
	})

	return &TechStack{
		Framework:    "fastapi",
		Language:     "python",
		Port:         8000,
		HasDockerfile: hasDocker,
		Dockerfile:   dockerfile,
		BuildCmd:     "pip install -r requirements.txt",
		StartCmd:     "uvicorn main:app --host 0.0.0.0 --port 8000",
		EnvVars:      envVars,
	}
}

// detectNodeBackend detects Node.js backend configuration
func (d *Detector) detectNodeBackend(deps map[string]string) *TechStack {
	framework := "express"
	if d.hasDep(deps, "fastify") {
		framework = "fastify"
	} else if d.hasDep(deps, "koa") {
		framework = "koa"
	}

	dockerfile, hasDocker := d.findDockerfile([]string{
		"backend/Dockerfile",
		"api/Dockerfile",
		"Dockerfile",
	})

	envVars := d.detectEnvVars([]string{
		"backend/.env",
		".env",
	})

	return &TechStack{
		Framework:    framework,
		Language:     "nodejs",
		Port:         8080,
		HasDockerfile: hasDocker,
		Dockerfile:   dockerfile,
		BuildCmd:     "npm run build",
		StartCmd:     "npm start",
		EnvVars:      envVars,
	}
}

// detectMiddleware detects middleware dependencies from code
func (d *Detector) detectMiddleware() Middleware {
	var m Middleware

	// Check in package.json
	pkgPath := filepath.Join(d.rootDir, "package.json")
	if content, err := os.ReadFile(pkgPath); err == nil {
		deps := d.parseDependencies(content)

		// Check for MongoDB clients
		if d.hasDep(deps, "mongodb") || d.hasDep(deps, "mongoose") || d.hasDep(deps, "@prisma/client") {
			// Check if prisma uses MongoDB by checking schema
			if !d.checkPrismaPostgres() {
				m.MongoDB = true
			}
		}

		// Check for Redis
		if d.hasDep(deps, "redis") || d.hasDep(deps, "ioredis") || d.hasDep(deps, "@redis/client") {
			m.Redis = true
		}

		// Check for Postgres
		if d.hasDep(deps, "pg") || d.hasDep(deps, "postgres") || d.hasDep(deps, "prisma") {
			if d.checkPrismaPostgres() {
				m.Postgres = true
			}
		}
	}

	// Check in go.mod
	goModPath := filepath.Join(d.rootDir, "go.mod")
	if content, err := os.ReadFile(goModPath); err == nil {
		contentStr := string(content)

		if strings.Contains(contentStr, "go.mongodb.org/mongo-driver") {
			m.MongoDB = true
		}
		if strings.Contains(contentStr, "github.com/redis/go-redis") || strings.Contains(contentStr, "github.com/go-redis/redis") {
			m.Redis = true
		}
		if strings.Contains(contentStr, "github.com/lib/pq") || strings.Contains(contentStr, "github.com/jackc/pgx") {
			m.Postgres = true
		}
	}

	// Check in requirements.txt
	reqPath := filepath.Join(d.rootDir, "requirements.txt")
	if content, err := os.ReadFile(reqPath); err == nil {
		contentStr := string(content)

		if strings.Contains(contentStr, "pymongo") || strings.Contains(contentStr, "motor") {
			m.MongoDB = true
		}
		if strings.Contains(contentStr, "redis") || strings.Contains(contentStr, "aioredis") {
			m.Redis = true
		}
		if strings.Contains(contentStr, "psycopg2") || strings.Contains(contentStr, "asyncpg") {
			m.Postgres = true
		}
	}

	return m
}

// checkPrismaPostgres checks if Prisma schema uses PostgreSQL
func (d *Detector) checkPrismaPostgres() bool {
	schemaPath := filepath.Join(d.rootDir, "prisma/schema.prisma")
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), `provider = "postgresql"`)
}

// detectEnvVars detects environment variables from .env files
func (d *Detector) detectEnvVars(paths []string) map[string]string {
	envVars := make(map[string]string)

	for _, path := range paths {
		fullPath := filepath.Join(d.rootDir, path)
		file, err := os.Open(fullPath)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				// Skip secrets - just store the key
				if !strings.Contains(strings.ToLower(key), "secret") &&
				   !strings.Contains(strings.ToLower(key), "password") &&
				   !strings.Contains(strings.ToLower(key), "key") {
					envVars[key] = "[VALUE]"
				}
			}
		}
	}

	return envVars
}

// Helper functions

func (d *Detector) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (d *Detector) findDockerfile(paths []string) (string, bool) {
	for _, path := range paths {
		fullPath := filepath.Join(d.rootDir, path)
		if d.fileExists(fullPath) {
			// Check if it's a valid Dockerfile
			if content, err := os.ReadFile(fullPath); err == nil {
				if d.isValidDockerfile(content) {
					return path, true
				}
			}
		}
	}
	return "", false
}

func (d *Detector) isValidDockerfile(content []byte) bool {
	contentStr := string(content)
	return strings.Contains(contentStr, "FROM") &&
		   (strings.Contains(contentStr, "EXPOSE") || strings.Contains(contentStr, "CMD") || strings.Contains(contentStr, "ENTRYPOINT"))
}

func (d *Detector) parseDependencies(content []byte) map[string]string {
	deps := make(map[string]string)

	lines := bytes.Split(content, []byte("\n"))
	inDeps := false

	for _, line := range lines {
		lineStr := strings.TrimSpace(string(line))

		if strings.HasPrefix(lineStr, `"dependencies":`) || strings.HasPrefix(lineStr, `'dependencies':`) {
			inDeps = true
			continue
		}

		if strings.HasPrefix(lineStr, `"devDependencies":`) || strings.HasPrefix(lineStr, `'devDependencies':`) {
			inDeps = true
			continue
		}

		if inDeps && (strings.HasPrefix(lineStr, "}") || strings.HasPrefix(lineStr, "]")) {
			inDeps = false
		}

		if inDeps {
			parts := strings.SplitN(lineStr, ":", 2)
			if len(parts) == 2 {
				name := strings.Trim(strings.TrimSpace(parts[0]), `"'`)
				if name != "" {
					deps[name] = strings.Trim(strings.TrimSpace(parts[1]), `"',`)
				}
			}
		}
	}

	return deps
}

func (d *Detector) hasDep(deps map[string]string, name string) bool {
	// Check exact match
	if _, ok := deps[name]; ok {
		return true
	}

	// Check prefix match
	for dep := range deps {
		if strings.HasPrefix(dep, name) {
			return true
		}
	}

	return false
}

func timestamp() string {
	return fmt.Sprintf("%d", 0) // Placeholder - would use actual timestamp
}
