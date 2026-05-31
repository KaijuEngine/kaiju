/******************************************************************************/
/* editor_plugins.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"bytes"
	"image/png"
	"log/slog"
	"reflect"

	"kaijuengine.com/build"
	"kaijuengine.com/editor/editor_embedded_content"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
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
	ed := &Editor{
		host:                   host,
		sessionDisabledPlugins: map[string]struct{}{},
	}
	host.SetGame(ed)
	if err := ed.settings.Load(); err != nil {
		slog.Error("failed to load the settings for the editor", "error", err)
	}
	ed.initializeActions()
	ed.initializeWebAPI()
	// goroutine
	go func() {
		data, err := host.AssetDatabase().Read("kaiju-icon.png")
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
			return
		}
		// validateCompiledPlugins inspects plugin.json vs the compiled-in
		// editorPluginRegistry. On match it invokes the onResolved callback
		// (newProjectOverlay) synchronously; on mismatch it shows a modal
		// that owns the resolution flow and calls onResolved itself.
		ed.validateCompiledPlugins(ed.newProjectOverlay)
	})
}
