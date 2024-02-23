//go:build editor

/******************************************************************************/
/* host.ed.go                                                                */
/******************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/******************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/******************************************************************************/

package engine

import "kaiju/klib"

type editorEntities []*Entity

func newEditorEntities() editorEntities {
	return make([]*Entity, 0)
}

func (e *editorEntities) remove(entity *Entity) {
	for i, t := range *e {
		if t == entity {
			*e = klib.RemoveUnordered(*e, i)
			break
		}
	}
}

func (e editorEntities) contains(entity *Entity) bool {
	for _, t := range e {
		if t == entity {
			return true
		}
	}
	return false
}

func (e *editorEntities) tickCleanup() {
	back := len(*e)
	for i, t := range *e {
		if t.TickCleanup() {
			(*e)[i] = (*e)[back-1]
			back--
		}
	}
	*e = (*e)[:back]
}

func (e editorEntities) resetDirty() {
	for _, t := range e {
		t.Transform.ResetDirty()
	}
}

func (host *Host) addEntity(entity *Entity) {
	if host.inEditorEntity > 0 {
		host.editorEntities = append(host.editorEntities, entity)
	} else {
		host.entities = append(host.entities, entity)
	}
}

func (host *Host) addEntities(entities ...*Entity) {
	if host.inEditorEntity > 0 {
		host.editorEntities = append(host.editorEntities, entities...)
	} else {
		host.entities = append(host.entities, entities...)
	}
}
