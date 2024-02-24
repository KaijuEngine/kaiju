/******************************************************************************/
/* history.go                                                                 */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).    */
/* Copyright (c) 2015-2023 Brent Farris.                                      */
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

package memento

type History struct {
	undoStack []Memento
	position  int
	limit     int
}

func NewHistory(limit int) History {
	return History{
		undoStack: make([]Memento, 0),
		limit:     limit,
	}
}

func (h *History) Add(m Memento) {
	for i := h.position; i < len(h.undoStack); i++ {
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
	if h.position == 0 {
		return
	}
	h.position--
	m := h.undoStack[h.position]
	m.Undo()
}

func (h *History) Redo() {
	if h.position == len(h.undoStack) {
		return
	}
	m := h.undoStack[h.position]
	m.Redo()
	h.position++
}

func (h *History) Clear() {
	for i := 0; i < len(h.undoStack); i++ {
		h.undoStack[i].Exit()
	}
	h.undoStack = h.undoStack[:0]
	h.position = 0
}
