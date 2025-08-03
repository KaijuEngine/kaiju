/******************************************************************************/
/* editor_settings_window.go                                                  */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package editor_settings_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/engine/host_container"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/plugins"
	"strconv"
)

type EditorSettingsWindow struct {
	uiMan            ui.Manager
	GoCompilerPath   string
	GridSnapping     float32
	RotationSnapping float32
}

func updateCompilerPath(elm *document.Element) {
	txt := elm.UI.ToInput().Text()
	editor_cache.SetEditorConfigValue(editor_cache.KaijuGoCompiler, txt)
}

func updateGridSnapping(elm *document.Element) {
	txt := elm.UI.ToInput().Text()
	if snap, err := strconv.ParseFloat(txt, 32); err == nil {
		editor_cache.SetEditorConfigValue(editor_cache.GridSnapping, snap)
	}
}

func updateRotationSnapping(elm *document.Element) {
	txt := elm.UI.ToInput().Text()
	if snap, err := strconv.ParseFloat(txt, 32); err == nil {
		editor_cache.SetEditorConfigValue(editor_cache.RotationSnapping, snap)
	}
}

func New() {
	const html = "editor/ui/editor_settings_window.html"
	esw := &EditorSettingsWindow{
		GridSnapping:     1,
		RotationSnapping: 15,
	}
	container := host_container.New("Editor Settings Window", nil)
	esw.uiMan.Init(container.Host)
	go container.Run(500, 300, -1, -1)
	<-container.PrepLock
	if v, ok := editor_cache.EditorConfigValue(editor_cache.KaijuGoCompiler); ok {
		esw.GoCompilerPath = v.(string)
	}
	if v, ok := editor_cache.EditorConfigValue(editor_cache.GridSnapping); ok {
		esw.GridSnapping = float32(v.(float64))
	}
	if v, ok := editor_cache.EditorConfigValue(editor_cache.RotationSnapping); ok {
		esw.RotationSnapping = float32(v.(float64))
	}
	container.RunFunction(func() {
		markup.DocumentFromHTMLAsset(&esw.uiMan, html, esw, map[string]func(*document.Element){
			"updateCompilerPath":     updateCompilerPath,
			"updateGridSnapping":     updateGridSnapping,
			"updateRotationSnapping": updateRotationSnapping,
			"regeneratePluginAPI":    func(*document.Element) { plugins.RegenerateAPI() },
		})
	})
}
