/******************************************************************************/
/* history.go                                                                 */
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

package memento

import (
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
)

type History struct {
	undoStack     []Memento
	position      int
	limit         int
	savedPosition int
	inAction      bool
}

func (h *History) Initialize(limit int) { h.limit = limit }

func (h *History) Add(m Memento) {
	defer tracing.NewRegion("History.Add").End()
	if h.inAction {
		return
	}
	for i := len(h.undoStack) - 1; i >= h.position; i-- {
		h.undoStack[i].Delete()
	}
	h.undoStack = h.undoStack[:h.position]
	h.undoStack = append(h.undoStack, m)
	h.position++
	if h.position > h.limit {
		h.position = h.limit
		h.undoStack[0].Exit()
		h.undoStack = h.undoStack[1:]
	}
}

func (h *History) Undo() {
	defer tracing.NewRegion("History.Undo").End()
	h.inAction = true
	if h.position == 0 {
		return
	}
	h.position--
	m := h.undoStack[h.position]
	m.Undo()
	h.inAction = false
}

func (h *History) Redo() {
	defer tracing.NewRegion("History.Redo").End()
	h.inAction = true
	if h.position == len(h.undoStack) {
		return
	}
	m := h.undoStack[h.position]
	m.Redo()
	h.position++
	h.inAction = false
}

func (h *History) Clear() {
	defer tracing.NewRegion("History.Clear").End()
	h.inAction = true
	for i := 0; i < len(h.undoStack); i++ {
		h.undoStack[i].Exit()
	}
	h.undoStack = klib.RemakeSlice(h.undoStack)
	h.position = 0
	h.inAction = false
}

func (h *History) SetSavePosition()        { h.savedPosition = h.position }
func (h *History) HasPendingChanges() bool { return h.savedPosition != h.position }
