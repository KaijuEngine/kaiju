/******************************************************************************/
/* window_listing.go                                                          */
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

package editor_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/klib"
	"slices"
	"sync"
)

type Listing struct {
	windows []EditorWindow
	mutex   sync.Mutex
}

func New() Listing {
	return Listing{
		windows: make([]EditorWindow, 0),
		mutex:   sync.Mutex{},
	}
}

func (l *Listing) Add(w EditorWindow) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.windows = append(l.windows, w)
	w.Container().Host.OnClose.Add(func() {
		l.mutex.Lock()
		if slices.Contains(l.windows, w) {
			saveLayout(w, false)
			w.Closed()
			l.Remove(w)
		}
		l.mutex.Unlock()
	})
}

func (l *Listing) Remove(w EditorWindow) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for i, win := range l.windows {
		if win == w {
			l.windows = klib.RemoveUnordered(l.windows, i)
			break
		}
	}
}

func (l *Listing) CloseAll() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	cpy := slices.Clone(l.windows)
	l.windows = make([]EditorWindow, 0)
	for _, win := range cpy {
		saveLayout(win, true)
		win.Container().Host.Close()
	}
}

func saveLayout(win EditorWindow, isOpen bool) {
	host := win.Container().Host
	x := host.Window.X()
	y := host.Window.Y()
	w := host.Window.Width()
	h := host.Window.Height()
	editor_cache.SetWindow(win.Tag(), x, y, w, h, isOpen)
}
