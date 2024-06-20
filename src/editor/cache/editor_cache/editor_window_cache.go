/******************************************************************************/
/* editor_window_cache.go                                                     */
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

package editor_cache

import (
	"encoding/json"
	"errors"
	"kaiju/filesystem"
	"path/filepath"
)

const (
	windowsFile = "layout.json"
)

type WindowInfo struct {
	X      int
	Y      int
	Width  int
	Height int
	Open   bool
}

var windows = map[string]WindowInfo{}

func SaveWindowCache() error {
	cache, err := cacheFolder()
	if err != nil {
		return err
	}
	str, err := json.Marshal(windows)
	if err != nil {
		return err
	}
	return filesystem.WriteTextFile(filepath.Join(cache, windowsFile), string(str))
}

func readCache() error {
	cache, err := cacheFolder()
	if err != nil {
		return err
	}
	str, err := filesystem.ReadTextFile(filepath.Join(cache, windowsFile))
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(str), &windows)
}

func SetWindow(key string, x, y, w, h int, isOpen bool) {
	windows[key] = WindowInfo{x, y, w, h, isOpen}
}

func Window(key string) (WindowInfo, error) {
	if w, ok := windows[key]; ok {
		return w, nil
	}
	if err := readCache(); err != nil {
		return WindowInfo{}, err
	}
	if w, ok := windows[key]; ok {
		return w, nil
	}
	return WindowInfo{}, errors.New("window info not found")
}

func WindowWasOpen(key string) bool {
	if w, err := Window(key); err == nil && w.Open {
		return true
	}
	return false
}
