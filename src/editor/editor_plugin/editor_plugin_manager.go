/******************************************************************************/
/* editor_plugin_manager.go                                                   */
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
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	pluginConfigFile     = "plugin.json"
	pluginEntryPointFile = "plugin.go"
	pluginsFolder        = "plugins"
)

const editorPluginGo = `package rename_me

// If you would like to debug your plugin and are working from the editor source
// code, stub your plugin import "_" in the editor_plugin_registry.go file.

import (
	"kaijuengine.com/editor"
	"kaijuengine.com/editor/editor_plugin"
)

// This key can be whatever you want, please make it unique so it doesn't
// collide with other's plugins. Using a URL or something unique like that
// is an option, but not required.
const pluginKey = "https://github.com/KaijuEngine/kaiju"

type Plugin struct {}

func init() {
	editor.RegisterPlugin(pluginKey, &Plugin{})
	// To register a workspace tab, also call:
	//   editor_workspace_registry.Register(&MyWorkspace{})
	// where MyWorkspace implements editor_workspace.Workspace
	// (see the built-in workspaces under editor/editor_workspace/* for examples).
}

func (p *Plugin) Launch(ed editor_plugin.EditorInterface) error {
	// TODO:  Implement. The ed interface gives you access to the host,
	// project, settings, events, history, stage view, and workspace
	// registry. To switch to a different workspace use ed.SelectWorkspace(id),
	// to query another workspace use ed.Workspace(id) and type-assert to a
	// well-known interface.
	return nil
}
`

type PluginConfig struct {
	Name        string
	PackageName string
	Description string
	Version     float64
	Author      string
	Website     string
	Enabled     bool
	GitModule   string `json:",omitempty"`
}

type PluginInfo struct {
	Path   string
	Config PluginConfig
}

func CreatePluginProject(path string) error {
	defer tracing.NewRegion("editor_plugin.CreatePluginProject").End()
	if err := createPluginFolder(path); err != nil {
		return err
	}
	if err := createConfigFile(path); err != nil {
		return err
	}
	return createEntryPointFile(path)
}

func IsPluginFolder(path string) bool {
	defer tracing.NewRegion("editor_plugin.IsPluginFolder").End()
	if s, err := os.Stat(path); err != nil || !s.IsDir() {
		return false
	}
	cfgFile := filepath.Join(path, pluginConfigFile)
	if s, err := os.Stat(cfgFile); err != nil || s.IsDir() {
		return false
	}
	var cfg PluginConfig
	f, err := os.Open(cfgFile)
	if err != nil {
		return false
	}
	defer f.Close()
	if err = json.NewDecoder(f).Decode(&cfg); err != nil {
		return false
	}
	if cfg.PackageName == "" {
		return false
	}
	return true
}

func PluginsFolder() (string, error) {
	defer tracing.NewRegion("editor_plugin.PluginsFolder").End()
	dir, err := filesystem.GameDirectory()
	if err != nil {
		return "", err
	}
	folder := filepath.Join(dir, pluginsFolder)
	os.MkdirAll(folder, os.ModePerm)
	if s, err := os.Stat(folder); err != nil {
		return "", err
	} else if !s.IsDir() {
		return "", fmt.Errorf("path is not a folder: %s", folder)
	}
	return folder, nil
}

func AvailablePlugins() []PluginInfo {
	defer tracing.NewRegion("editor_plugin.AvailablePlugins").End()
	plugs := []PluginInfo{}

	plugFolder, err := PluginsFolder()
	if err != nil {
		return plugs
	}
	dir, err := os.ReadDir(plugFolder)
	if err != nil {
		return plugs
	}
	folders := make([]string, 0, len(dir))
	for i := range dir {
		if !dir[i].IsDir() {
			continue
		}
		folders = append(folders, filepath.Join(plugFolder, dir[i].Name()))
	}
	for i := range folders {
		cfgPath := filepath.Join(folders[i], pluginConfigFile)
		if s, err := os.Stat(cfgPath); err != nil || s.IsDir() {
			continue
		}
		f, err := os.Open(cfgPath)
		if err != nil {
			continue
		}
		var cfg PluginConfig
		if err = json.NewDecoder(f).Decode(&cfg); err == nil {
			path := folders[i]
			if cfg.GitModule != "" {
				path = "git://" + cfg.GitModule
			}
			plugs = append(plugs, PluginInfo{
				Path:   path,
				Config: cfg,
			})
		}
		f.Close()
	}

	return plugs
}

func UpdatePluginConfigState(info PluginInfo) error {
	// Skip Git plugins - they don't have physical config files to update
	if strings.HasPrefix(info.Path, "git://") {
		return nil
	}

	f, err := os.Create(filepath.Join(info.Path, pluginConfigFile))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(info.Config)
}

func createPluginFolder(path string) error {
	defer tracing.NewRegion("editor_plugin.createPluginFolder").End()
	if s, err := os.Stat(path); err == nil && !s.IsDir() {
		return fmt.Errorf("failed to create the plugin folder '%s', it's a file", path)
	}
	if dir, _ := os.ReadDir(path); len(dir) > 0 {
		return fmt.Errorf("failed to create the plugin folder, '%s' is not an empty folder", path)
	}
	os.MkdirAll(path, os.ModePerm)
	return nil
}

func createConfigFile(path string) error {
	defer tracing.NewRegion("editor_plugin.createConfigFile").End()
	cfg := PluginConfig{
		Name:        "RENAME ME",
		PackageName: "rename_me",
		Description: "My cool plugin does things",
		Version:     0.001,
		Author:      "Brent Farris",
		Website:     "https://github.com/KaijuEngine/kaiju",
		Enabled:     false,
	}
	f, err := os.Create(filepath.Join(path, pluginConfigFile))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(cfg)
}

func createEntryPointFile(path string) error {
	defer tracing.NewRegion("editor_plugin.createEntryPointFile").End()
	f, err := os.Create(filepath.Join(path, pluginEntryPointFile))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(editorPluginGo)
	return err
}
