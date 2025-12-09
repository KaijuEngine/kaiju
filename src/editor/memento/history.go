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

package memento

import (
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"reflect"
)

type History struct {
	undoStack     []Memento
	transaction   *HistoryTransaction
	position      int
	limit         int
	savedPosition int
	lockAdditions bool
}

// Initialize sets the max number of undo entries that the history will retain.
func (h *History) Initialize(limit int) { h.limit = limit }

// LockAdditions prevents new mementos from being added to the history.
func (h *History) LockAdditions() { h.lockAdditions = true }

// UnlockAdditions re-enables adding new mementos after a lock.
func (h *History) UnlockAdditions() { h.lockAdditions = false }

// IsInTransaction reports whether a history transaction is currently active.
func (h *History) IsInTransaction() bool { return h.transaction != nil }

// BeginTransaction starts a new transaction. Subsequent Add calls will be
// queued in the transaction until it is committed or cancelled. If committed
// all of the undos will be joined together into a single undo/redo operation.
// You should start a transaction when calling common methods that would create
// their own internal history.
//
// For example, when you delete selected entities, the clear function is called
// on the selection (to update UI, visuals, and other things). This clear call
// will generate history, thus creating 2 history entries (clear and delete). By
// starting a transaction, both of those will be within the same undo/redo call.
func (h *History) BeginTransaction() {
	h.transaction = &HistoryTransaction{}
}

// CommitTransaction finalizes the current transaction, adding all queued
// mementos to the history as a single atomic operation.
func (h *History) CommitTransaction() {
	t := h.transaction
	h.transaction = nil
	h.Add(t)
}

// CancelTransaction aborts the current transaction, discarding any queued
// mementos.
func (h *History) CancelTransaction() { h.transaction = nil }

// Add inserts a new memento into the history. If a transaction is active, the
// memento is queued instead of being added immediately.
func (h *History) Add(m Memento) {
	defer tracing.NewRegion("History.Add").End()
	if h.IsInTransaction() {
		h.transaction.stack = append(h.transaction.stack, m)
		return
	}
	if h.lockAdditions {
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

func (h *History) Last() (Memento, bool) {
	if h.position == 0 {
		return nil, false
	}
	return h.undoStack[h.position-1], true
}

// AddOrReplaceLast inserts a new memento into the history. If the most recent
// memento (the one at h.position‑1) has the same concrete type as the supplied
// memento, it is replaced instead of creating a new entry. This is useful for
// collapsing consecutive operations of the same kind (e.g. repeated brush
// strokes) into a single undo step.
func (h *History) AddOrReplaceLast(m Memento) {
	defer tracing.NewRegion("History.AddOrReplaceLast").End()
	if h.IsInTransaction() {
		h.transaction.stack = append(h.transaction.stack, m)
		return
	}
	if h.lockAdditions {
		return
	}
	if h.position > 0 {
		lastIdx := h.position - 1
		last := h.undoStack[lastIdx]
		if reflect.TypeOf(last) == reflect.TypeOf(m) {
			last.Delete()
			h.undoStack[lastIdx] = m
			return
		}
	}
	h.Add(m)
}

// Undo reverts the most recent memento, moving the current position back.
func (h *History) Undo() {
	defer tracing.NewRegion("History.Undo").End()
	h.LockAdditions()
	defer h.UnlockAdditions()
	if h.position == 0 {
		return
	}
	h.position--
	m := h.undoStack[h.position]
	m.Undo()
}

// Redo reapplies the next memento in the stack, moving the position forward.
func (h *History) Redo() {
	defer tracing.NewRegion("History.Redo").End()
	h.LockAdditions()
	defer h.UnlockAdditions()
	if h.position == len(h.undoStack) {
		return
	}
	m := h.undoStack[h.position]
	m.Redo()
	h.position++
}

// Clear removes all mementos from the history and resets the position.
func (h *History) Clear() {
	defer tracing.NewRegion("History.Clear").End()
	h.LockAdditions()
	defer h.UnlockAdditions()
	for i := 0; i < len(h.undoStack); i++ {
		h.undoStack[i].Exit()
	}
	h.undoStack = klib.RemakeSlice(h.undoStack)
	h.position = 0
}

// SetSavePosition records the current position as the saved state.
func (h *History) SetSavePosition() { h.savedPosition = h.position }

// HasPendingChanges reports whether the history has changes since the last
// saved position.
func (h *History) HasPendingChanges() bool { return h.savedPosition != h.position }
