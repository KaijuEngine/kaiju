/******************************************************************************/
/* editor_plugin_data.go                                                      */
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

package editor

import (
	"fmt"
	"io"
	"io/fs"
	"kaiju/build"
	"kaiju/editor/editor_plugin"
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/filesystem"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var editorPluginRegistry = map[string]editor_plugin.EditorPlugin{}

func RegisterPlugin(key string, plugin editor_plugin.EditorPlugin) {
	if _, ok := editorPluginRegistry[key]; ok {
		slog.Error("a plugin with the given key is already registered", "key", key)
		return
	}
	editorPluginRegistry[key] = plugin
}

func (ed *Editor) RecompileWithPlugins(plugins []editor_plugin.PluginInfo, onComplete func(err error)) error {
	// Copy editor source to build folder
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir, err := filesystem.GameDirectory()
	if err != nil {
		return err
	}
	to := filepath.Join(dir, "editor_build")
	os.MkdirAll(to, os.ModePerm)
	if err = copyEditorCodeForRecompile(to, project_file_system.EngineFS.EngineFileSystemInterface); err != nil {
		return err
	}
	registry, err := os.OpenFile(filepath.Join(to, "editor_plugin_registry.go"), os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer registry.Close()
	registry.WriteString("\nimport (\n")
	// Copy each enabled plugin into editor/editor_plugin/developer_plugins
	for i := range plugins {
		if !plugins[i].Config.Enabled {
			continue
		}
		dstName := plugins[i].Config.PackageName
		dst := filepath.Join(to, "editor/editor_plugin/developer_plugins", dstName)
		filesystem.CopyDirectory(plugins[i].Path, dst)
		registry.WriteString(fmt.Sprintf("\t_ \"kaiju/editor/editor_plugin/developer_plugins/%s\"\n", dstName))
		if err = editor_plugin.UpdatePluginConfigState(plugins[i]); err != nil {
			slog.Warn("failed to update the enabled state of the plugin",
				"name", plugins[i].Config.Name, "package", plugins[i].Config.PackageName, "error", err)
		}
	}
	registry.WriteString(")\n")
	// Run compile of the editor
	var cmd *exec.Cmd
	if build.Debug {
		cmd = exec.Command("go", "build", "-tags=debug,editor", "-o", filepath.Base(exe), ".")
	} else {
		cmd = exec.Command("go", "build", "-tags=editor", "-o", filepath.Base(exe), ".")
	}
	cmd.Dir = to
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		defer onComplete(err)
		// TODO:  Add better error messaging for failed compile
		if err = cmd.Wait(); err != nil {
			slog.Error("failed to compile the editor with the plugins", "error", err)
			return
		}
		// Launch the generators/plugin_installer/main.go
		toExe := filepath.Join(to, filepath.Base(exe))
		boot := exec.Command("go", "run", "generators/plugin_installer/main.go", exe, toExe)
		boot.Dir = to
		if err = boot.Start(); err != nil {
			slog.Error("failed to start the restart boot process", "error", err)
			return
		}
		slog.Info("attempting to restart editor with new build")
		ed.host.Close()
	}()
	return nil
}

func copyEditorCodeForRecompile(to string, efs project_file_system.EngineFileSystemInterface) error {
	const from = "."
	var err error
	var copyFolder func(path string) error
	os.RemoveAll(to)
	os.MkdirAll(to, os.ModePerm)
	copyFolder = func(path string) error {
		relPath, _ := filepath.Rel(from, path)
		folder := filepath.Join(to, relPath)
		if path != "." {
			if err := os.Mkdir(folder, os.ModePerm); err != nil {
				return err
			}
		}
		var dir []fs.DirEntry
		if dir, err = efs.ReadDir(path); err != nil {
			return err
		}
		for i := range dir {
			name := dir[i].Name()
			entryPath := filepath.ToSlash(filepath.Join(path, name))
			if dir[i].IsDir() {
				if copyFolder(entryPath); err != nil {
					return err
				} else {
					continue
				}
			}
			if strings.HasPrefix(name, "__") {
				continue
			}
			f, err := efs.Open(entryPath)
			if err != nil {
				return err
			}
			defer f.Close()
			t, err := os.Create(filepath.Join(folder, dir[i].Name()))
			if err != nil {
				return err
			}
			defer t.Close()
			if _, err := io.Copy(t, f); err != nil {
				return err
			}
		}
		return nil
	}
	copyFolder(from)
	return err
}
