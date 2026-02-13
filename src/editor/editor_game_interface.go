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
	"kaiju/build"
	"kaiju/editor/editor_embedded_content"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"reflect"
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
}
