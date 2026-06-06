/******************************************************************************/
/* plugin_installer.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kaijuengine.com/platform/filesystem"
)

func AddGitPluginToStorage(modulePath string) error {
	plugFolder, err := PluginsFolder()
	if err != nil {
		return err
	}

	exists, err := gitPluginAlreadyStored(modulePath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	module := extractModule(modulePath)
	author, packageName := extractAuthorAndPackage(module)

	folderPath := filepath.Join(plugFolder, buildFolderName(modulePath))

	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return err
	}

	cfg := buildPluginConfig(modulePath, module, author, packageName)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return filesystem.WriteFile(filepath.Join(folderPath, pluginConfigFile), data)
}

func RemoveGitPluginFromStorage(modulePath string) error {
	plugFolder, err := PluginsFolder()
	if err != nil {
		return err
	}

	dirs, err := os.ReadDir(plugFolder)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		cfg, err := loadPluginConfig(plugFolder, dir.Name())
		if err != nil {
			continue
		}

		if cfg.GitModule == modulePath {
			return os.RemoveAll(filepath.Join(plugFolder, dir.Name()))
		}
	}

	return nil
}

func GetStoredGitPlugins() ([]string, error) {
	plugFolder, err := PluginsFolder()
	if err != nil {
		return nil, err
	}

	dirs, err := os.ReadDir(plugFolder)
	if err != nil {
		return nil, err
	}

	var plugins []string

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		cfg, err := loadPluginConfig(plugFolder, dir.Name())
		if err != nil {
			continue
		}

		if cfg.GitModule != "" {
			plugins = append(plugins, cfg.GitModule)
		}
	}

	return plugins, nil
}

func loadPluginConfig(basePath, dirName string) (PluginConfig, error) {
	cfgPath := filepath.Join(basePath, dirName, pluginConfigFile)

	info, err := os.Stat(cfgPath)
	if err != nil || info.IsDir() {
		return PluginConfig{}, fmt.Errorf("invalid config")
	}

	data, err := filesystem.ReadFile(cfgPath)
	if err != nil {
		return PluginConfig{}, err
	}

	var cfg PluginConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return PluginConfig{}, err
	}

	return cfg, nil
}

func gitPluginAlreadyStored(modulePath string) (bool, error) {
	plugins, err := GetStoredGitPlugins()
	if err != nil {
		return false, err
	}

	for _, plugin := range plugins {
		if plugin == modulePath {
			return true, nil
		}
	}

	return false, nil
}

func extractModule(modulePath string) string {
	return strings.Split(modulePath, "@")[0]
}

// For a module path like "github.com/author/package@ref", this returns "author" and "package"
func extractAuthorAndPackage(module string) (author, packageName string) {
	parts := strings.Split(module, "/")
	n := len(parts)
	return parts[n-2], parts[n-1]
}

func buildFolderName(modulePath string) string {
	replacer := strings.NewReplacer("/", "_", "@", "_", ":", "_")
	return "git_" + replacer.Replace(modulePath)
}

func buildPluginConfig(modulePath, module, author, packageName string) PluginConfig {
	return PluginConfig{
		Name:        packageName,
		PackageName: packageName,
		Description: fmt.Sprintf("Git plugin from %s", module),
		Version:     0.1,
		Author:      author,
		Website:     "https://" + module,
		Enabled:     true,
		GitModule:   modulePath,
	}
}

func parseGitURL(gitURL string) (modulePath, ref string) {
	clean := strings.TrimSpace(gitURL)

	if idx := strings.IndexAny(clean, "?#"); idx != -1 {
		clean = clean[:idx]
	}

	if strings.HasPrefix(clean, "git@") {
		clean = strings.TrimPrefix(clean, "git@")
		clean = strings.Replace(clean, ":", "/", 1)
	}

	clean = strings.TrimPrefix(clean, "https://")
	clean = strings.TrimPrefix(clean, "http://")
	clean = strings.TrimPrefix(clean, "git://")

	clean = strings.TrimSuffix(clean, ".git")
	clean = strings.TrimSuffix(clean, "/")

	ref = "latest"

	if idx := strings.LastIndex(clean, "@"); idx != -1 {
		candidate := clean[idx+1:]
		clean = clean[:idx]

		if candidate != "" {
			ref = candidate
		}
	}

	modulePath = clean
	return
}

func AddPluginFromGit(gitURL string) (string, error) {
	modulePath, ref := parseGitURL(gitURL)

	if strings.Contains(modulePath, "github.com/KaijuEngine/kaiju") {
		modulePath = "kaijuengine.com"
		ref = ""
	}

	fullModuleRef := modulePath
	if ref != "" {
		fullModuleRef = fmt.Sprintf("%s@%s", modulePath, ref)
	}

	if err := AddGitPluginToStorage(fullModuleRef); err != nil {
		return "", fmt.Errorf("failed to save Git plugin to storage: %w", err)
	}

	return fullModuleRef, nil
}

func AddPluginFromGitHub(githubURL string) (string, error) {
	return AddPluginFromGit(githubURL)
}
