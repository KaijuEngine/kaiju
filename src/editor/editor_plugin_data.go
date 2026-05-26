/******************************************************************************/
/* editor_plugin_data.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"kaijuengine.com/build"
	"kaijuengine.com/editor/editor_plugin"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/filesystem"
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
	// O_APPEND alone opens the file read-only on macOS/Linux, so every
	// WriteString below silently fails with "bad file descriptor" and no
	// plugin imports ever land in the build's editor_plugin_registry.go.
	// Add O_WRONLY so the appends actually write.
	registry, err := os.OpenFile(filepath.Join(to, "editor_plugin_registry.go"), os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer registry.Close()
	if _, err := registry.WriteString("\nimport (\n"); err != nil {
		return err
	}
	for i := range plugins {
		if !plugins[i].Config.Enabled {
			continue
		}
		if strings.HasPrefix(plugins[i].Path, "git://") {
			moduleRef := strings.TrimPrefix(plugins[i].Path, "git://")
			modulePath := strings.Split(moduleRef, "@")[0]

			getCmd := exec.Command("go", "get", moduleRef)
			getCmd.Dir = to
			if output, getErr := getCmd.CombinedOutput(); getErr != nil {
				slog.Error("failed to get git plugin module", "module", moduleRef, "error", getErr, "output", string(output))
				return getErr
			}

			if _, err := registry.WriteString(fmt.Sprintf("\t_ \"%s\"\n", modulePath)); err != nil {
				return err
			}
			continue
		}
		dstName := plugins[i].Config.PackageName
		dst := filepath.Join(to, "editor/editor_plugin/developer_plugins", dstName)
		if err = filesystem.CopyDirectory(plugins[i].Path, dst); err != nil {
			slog.Error("failed to copy plugin directory", "source", plugins[i].Path, "destination", dst, "error", err)
			return err
		}
		os.Remove(filepath.Join(dst, "go.mod"))
		os.Remove(filepath.Join(dst, "go.sum"))

		if _, err := registry.WriteString(fmt.Sprintf("\t_ \"kaijuengine.com/editor/editor_plugin/developer_plugins/%s\"\n", dstName)); err != nil {
			return err
		}
		if err = editor_plugin.UpdatePluginConfigState(plugins[i]); err != nil {
			slog.Warn("failed to update the enabled state of the plugin",
				"name", plugins[i].Config.Name, "package", plugins[i].Config.PackageName, "error", err)
		}
	}
	if _, err := registry.WriteString(")\n"); err != nil {
		return err
	}

	var cmd *exec.Cmd
	if build.Debug {
		cmd = exec.Command("go", "build", "-tags=debug,editor,filedrop", "-o", filepath.Base(exe), ".")
	} else {
		cmd = exec.Command("go", "build", "-tags=editor,filedrop", "-o", filepath.Base(exe), ".")
	}
	cmd.Dir = to

	var buildOutput strings.Builder
	cmd.Stdout = &buildOutput
	cmd.Stderr = &buildOutput

	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		defer onComplete(err)
		err = cmd.Wait()
		if err != nil {
			slog.Error("failed to compile the editor with the plugins", "error", err, "output", buildOutput.String())
			return
		}
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
			if err := os.MkdirAll(folder, os.ModePerm); err != nil {
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
