package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type App struct {
	srcPath  string
	docsPath string
	logger   *slog.Logger
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if len(os.Args) < 2 {
		logger.Error("missing project path argument", "usage", os.Args[0]+" <project_path>")
		os.Exit(1)
	}

	app := NewApp(os.Args[1], logger)

	if err := app.Run(); err != nil {
		logger.Error("application failed", "error", err)
		os.Exit(1)
	}

	logger.Info("documentation generated successfully")
}

func NewApp(srcPath string, logger *slog.Logger) *App {
	docsPath := filepath.Join(filepath.Dir(srcPath), "docs", "api")
	return &App{
		srcPath:  srcPath,
		docsPath: docsPath,
		logger:   logger,
	}
}

func (a *App) Run() error {
	if err := os.MkdirAll(a.docsPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating docs directory: %w", err)
	}

	packages, err := findPackages(a.srcPath)
	if err != nil {
		return err
	}

	a.printPackages(packages)

	for name, path := range packages {
		if err := a.generatePackageDoc(name, path); err != nil {
			a.logger.Warn("failed to generate docs", "package", name, "error", err)
		}
	}

	return createIndex(packages, a.docsPath)
}

func (a *App) printPackages(packages map[string]string) {
	a.logger.Info("packages discovered", "count", len(packages))

	for name, path := range packages {
		a.logger.Info("package", "name", name, "path", path)
	}
}

func (a *App) generatePackageDoc(name, path string) error {
	outputFile := filepath.Join(a.docsPath, name+".md")

	a.logger.Info("generating docs", "package", name)

	if err := runGomarkdoc(a.srcPath, path, outputFile); err != nil {
		return err
	}

	a.logger.Info("generated", "file", outputFile)
	return nil
}

func findPackages(srcPath string) (map[string]string, error) {
	packages := make(map[string]string)

	entries, err := os.ReadDir(srcPath)
	if err != nil {
		return nil, err
	}

	if hasGoFiles(srcPath) {
		packages["root"] = "."
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if shouldSkip(name) {
			continue
		}

		packagePath := filepath.Join(srcPath, name)
		if hasGoFiles(packagePath) {
			packages[name] = name
		}
	}

	return packages, nil
}

func hasGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true
		}
	}
	return false
}

func shouldSkip(name string) bool {
	skipDirs := map[string]struct{}{
		".git": {}, "vendor": {}, "node_modules": {}, "build": {}, "dist": {}, "docs": {},
		"bullet3": {}, "soloud": {}, "libs": {}, "tools": {}, "file_templates": {},
		"generators": {}, "ollama": {}, "network": {},
	}

	_, exists := skipDirs[name]
	return exists
}

func runGomarkdoc(srcPath, packagePath, outputFile string) error {
	target := "."
	if packagePath != "." {
		target = "./" + packagePath
	}

	cmd := exec.Command("gomarkdoc", "--output", outputFile, target)
	cmd.Dir = srcPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gomarkdoc failed: %w | output: %s", err, string(output))
	}

	return nil
}

func createIndex(packages map[string]string, docsPath string) error {
	var builder strings.Builder

	builder.WriteString("# API Documentation\n\n")
	builder.WriteString("Auto-generated documentation using gomarkdoc.\n\n")

	categories := categorizePackages(packages)

	order := []string{"Core", "Engine", "Platform", "Rendering", "Registry", "Other"}

	for _, cat := range order {
		pkgs := categories[cat]
		if len(pkgs) == 0 {
			continue
		}

		builder.WriteString(fmt.Sprintf("## %s\n\n", cat))
		for _, pkg := range pkgs {
			displayName := pkg
			if pkg == "root" {
				displayName = "Root Package"
			}
			builder.WriteString(fmt.Sprintf("- [%s](%s.md)\n", displayName, pkg))
		}
		builder.WriteString("\n")
	}

	indexFile := filepath.Join(docsPath, "index.md")
	return os.WriteFile(indexFile, []byte(builder.String()), 0644)
}

func categorizePackages(packages map[string]string) map[string][]string {
	categories := map[string][]string{
		"Core":      {},
		"Engine":    {},
		"Platform":  {},
		"Rendering": {},
		"Registry":  {},
		"Other":     {},
	}

	for name := range packages {
		category := categorize(name)
		categories[category] = append(categories[category], name)
	}

	return categories
}

func categorize(packageName string) string {
	switch {
	case packageName == "bootstrap":
		return "Core"
	case strings.HasPrefix(packageName, "engine"):
		return "Engine"
	case strings.HasPrefix(packageName, "platform"):
		return "Platform"
	case packageName == "rendering":
		return "Rendering"
	case strings.Contains(packageName, "registry"):
		return "Registry"
	default:
		return "Other"
	}
}
