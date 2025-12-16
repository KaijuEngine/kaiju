/******************************************************************************/
/* editor_plugin_manager.go                                                   */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package editor_plugin

import (
	"encoding/json"
	"fmt"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"os"
	"path/filepath"
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
	"kaiju/editor"
	"kaiju/editor/editor_plugin"
)

// This key can be whatever you want, please make it unique so it doesn't
// collide with other's plugins. Using a URL or something unique like that
// is an option, but not required.
const pluginKey = "https://github.com/KaijuEngine/kaiju"

type Plugin struct {}

func init() { editor.RegisterPlugin(pluginKey, &Plugin{}) }

func (p *Plugin) Launch(ed editor_plugin.EditorInterface) error {
	// TODO:  Implement
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
			plugs = append(plugs, PluginInfo{
				Path:   folders[i],
				Config: cfg,
			})
		}
		f.Close()
	}

	return plugs
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
