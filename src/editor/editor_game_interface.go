/******************************************************************************/
/* editor_plugins.go                                                          */
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
	"bytes"
	"image/png"
	"log/slog"
	"reflect"
	"runtime"
	"time"

	"kaijuengine.com/build"
	"kaijuengine.com/editor/editor_embedded_content"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

// EditorGame satisfies [bootstrap.GameInterface] and will allow the engine to
// bootstrap the editor (as it would a game).
type EditorGame struct{}

func (EditorGame) PluginRegistry() []reflect.Type {
	defer tracing.NewRegion("EditorGame.PluginRegistry").End()
	return []reflect.Type{}
}

func (EditorGame) ContentDatabase() (assets.Database, error) {
	return &editor_embedded_content.EditorContent{}, nil
}

func (EditorGame) Launch(host *engine.Host) {
	defer tracing.NewRegion("EditorGame.Launch").End()
	ed := &Editor{host: host}
	host.SetGame(ed)
	if err := ed.settings.Load(); err != nil {
		slog.Error("failed to load the settings for the editor", "error", err)
	}
	// goroutine
	go func() {
		data, err := host.AssetDatabase().Read("kiaju-icon.png")
		if err != nil {
			slog.Error("failed to read the editor application icon", "error", err)
			return
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			slog.Error("failed to decode the application icon file", "error", err)
			return
		}
		host.RunOnMainThread(func() {
			host.Window.SetIcon(img)
		})
	}()
	ed.UpdateSettings()
	ed.logging.Initialize(host, host.LogStream)
	ed.history.Initialize(512)
	ed.contentPreviewer.Initialize(ed)
	ed.earlyLoadUI()
	// Wait 2 frames to blur so the UI is updated properly before being disabled
	host.RunAfterFrames(2, func() {
		ed.BlurInterface()
		if build.Debug && engine.LaunchParams.AutoTest {
			// Auto-test mode: create a temporary test project automatically
			ed.createProject("AutoTest", "autotestproject", "")
		} else {
			ed.newProjectOverlay()
		}
	})
	if runtime.GOOS == "windows" {
		host.RunAfterTime(1*time.Second, func() {
			runNativeDialogDemo(host)
		})
	}
}

func runNativeDialogDemo(host *engine.Host) {
	defer tracing.NewRegion("runNativeDialogDemo").End()
	slog.Info("native dialog demo: opening file dialog")
	err := host.Window.OpenFileDialog("", []filesystem.DialogExtension{
		{Name: "Go Files", Extension: ".go"},
		{Name: "All Files", Extension: ".*"},
	}, func(path string) {
		slog.Info("native dialog demo: open file selected", "path", path)
		runNativeSaveDialogDemo(host)
	}, func() {
		slog.Info("native dialog demo: open file canceled")
		runNativeSaveDialogDemo(host)
	})
	if err != nil {
		slog.Error("native dialog demo: failed to show open file dialog", "error", err)
		runNativeSaveDialogDemo(host)
	}
}

func runNativeSaveDialogDemo(host *engine.Host) {
	defer tracing.NewRegion("runNativeSaveDialogDemo").End()
	slog.Info("native dialog demo: opening save file dialog")
	err := host.Window.SaveFileDialog("", "demo_output.txt", []filesystem.DialogExtension{
		{Name: "Text Files", Extension: ".txt"},
		{Name: "All Files", Extension: ".*"},
	}, func(path string) {
		slog.Info("native dialog demo: save file selected", "path", path)
		runNativeFolderDialogDemo(host)
	}, func() {
		slog.Info("native dialog demo: save file canceled")
		runNativeFolderDialogDemo(host)
	})
	if err != nil {
		slog.Error("native dialog demo: failed to show save file dialog", "error", err)
		runNativeFolderDialogDemo(host)
	}
}

func runNativeFolderDialogDemo(host *engine.Host) {
	defer tracing.NewRegion("runNativeFolderDialogDemo").End()
	slog.Info("native dialog demo: opening folder dialog")
	err := host.Window.OpenFolderDialog("", func(path string) {
		slog.Info("native dialog demo: folder selected", "path", path)
		runNativeAdvancedDialogDemo(host)
	}, func() {
		slog.Info("native dialog demo: folder dialog canceled")
		runNativeAdvancedDialogDemo(host)
	})
	if err != nil {
		slog.Error("native dialog demo: failed to show folder dialog", "error", err)
		runNativeAdvancedDialogDemo(host)
	}
}

func runNativeAdvancedDialogDemo(host *engine.Host) {
	defer tracing.NewRegion("runNativeAdvancedDialogDemo").End()

	root, err := filesystem.GameDirectory()
	if err != nil {
		slog.Error("native dialog demo: failed to resolve game directory for advanced dialog", "error", err)
		root = ""
	}

	request := filesystem.NativeDialogRequest{
		Mode:             filesystem.NativeDialogModeOpenFiles,
		Title:            "Advanced Native Dialog Demo (multi-select + options)",
		CurrentDirectory: root,
		Root:             root,
		ShowHidden:       true,
		Filters: []filesystem.DialogFilter{
			{Name: "Go and Text Files", Patterns: []string{"*.go", "*.txt"}},
			{Name: "Images", Patterns: []string{"*.png", "*.jpg", "*.jpeg"}},
			{Name: "All Files", Patterns: []string{"*.*"}},
		},
		Options: []filesystem.DialogCustomOption{
			{Name: "Recursive import", Default: 1},
			{Name: "Import mode", Values: []string{"Copy", "Reference", "Link"}, Default: 0},
		},
		WindowHandle: host.Window.PlatformWindow(),
	}

	slog.Info("native dialog demo: opening advanced dialog", "root", root)
	host.Window.DisableRawMouseInput()
	err = filesystem.OpenNativeDialogWindow(request, func(result filesystem.NativeDialogResult) {
		host.Window.EnableRawMouseInput()
		switch result.Status {
		case filesystem.NativeDialogStatusAccepted:
			slog.Info("native dialog demo: multi-select accepted",
				"count", len(result.Paths),
				"paths", result.Paths,
				"selectedFilterIndex", result.SelectedFilterIndex,
				"selectedOptions", result.SelectedOptions)
			for i := range result.Paths {
				slog.Info("native dialog demo: selected file", "index", i, "path", result.Paths[i])
			}
		case filesystem.NativeDialogStatusCancel:
			slog.Info("native dialog demo: advanced dialog canceled")
		case filesystem.NativeDialogStatusFailed:
			slog.Error("native dialog demo: advanced dialog failed",
				"error", result.Err,
				"hresult", result.HResult)
		}
	})
	if err != nil {
		host.Window.EnableRawMouseInput()
		slog.Error("native dialog demo: failed to open advanced dialog", "error", err)
	}
}
