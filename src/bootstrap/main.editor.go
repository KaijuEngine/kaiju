//go:build editor

/******************************************************************************/
/* main.editor.go                                                            */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package bootstrap

import (
	"kaiju/editor"
	"kaiju/editor/cache/editor_cache"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/profiler"
	"kaiju/systems/logging"
	tests "kaiju/tests/rendering_tests"
	"kaiju/tools/html_preview"
)

func addConsole(host *engine.Host) {
	html_preview.SetupConsole(host)
	profiler.SetupConsole(host)
	tests.SetupConsole(host)
}

func Main() {
	logStream := logging.Initialize(nil)
	container := host_container.New("Kaiju Editor", logStream)
	w := engine.DefaultWindowWidth
	h := engine.DefaultWindowHeight
	x, y := -1, -1
	if win, err := editor_cache.Window(editor_cache.MainWindow); err == nil {
		w = win.Width
		h = win.Height
		x = win.X
		y = win.Y
	}
	go container.Run(w, h, x, y)
	<-container.PrepLock
	container.RunFunction(func() {
		addConsole(container.Host)
	})
	editor := editor.New(container)
	container.RunFunction(func() {
		editor.SetupUI()
	})
	<-editor.Host().Done()
	x = editor.Host().Window.X()
	y = editor.Host().Window.Y()
	w = editor.Host().Window.Width()
	h = editor.Host().Window.Height()
	editor_cache.SetWindow(editor_cache.MainWindow, x, y, w, h)
}
