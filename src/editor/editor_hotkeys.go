/******************************************************************************/
/* editor_hotkeys.go                                                          */
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

package editor

import (
	"kaiju/editor/viewport/tools/deleter"
	"kaiju/editor/viewport/tools/transform_tools"
	"kaiju/hid"
	"kaiju/klib"
)

func checkHotkeys(ed *Editor) {
	kb := &ed.Host().Window.Keyboard
	if kb.HasCtrl() {
		if kb.KeyDown(hid.KeyboardKeyZ) {
			ed.history.Undo()
		} else if kb.KeyDown(hid.KeyboardKeyY) {
			ed.history.Redo()
		} else if kb.KeyUp(hid.KeyboardKeyS) {
			ed.stageManager.Save(ed.statusBar)
		} else if kb.KeyUp(hid.KeyboardKeyP) {
			ed.selection.Parent(&ed.history)
			ed.statusBar.SetMessage("Parented entities")
			ed.ReloadTabs("Hierarchy")
		} else if kb.KeyUp(hid.KeyboardKeyF5) {
			ed.runProject(false)
		}
	} else if kb.HasShift() {
		if kb.KeyUp(hid.KeyboardKeyF5) {
			ed.killDebug()
		}
	} else if kb.KeyUp(hid.KeyboardKey1) {
		ed.tabContainers[0].Toggle()
	} else if kb.KeyUp(hid.KeyboardKey2) {
		ed.tabContainers[1].Toggle()
	} else if kb.KeyUp(hid.KeyboardKey3) {
		ed.tabContainers[2].Toggle()
	} else if kb.KeyUp(hid.KeyboardKeyF1) {
		klib.OpenWebsite("https://kaijuengine.org/")
	} else if kb.KeyUp(hid.KeyboardKeyF5) {
		ed.runProject(true)
	} else if kb.KeyDown(hid.KeyboardKeyF) && ed.selection.HasSelection() {
		ed.selection.Focus(ed.Host().Camera)
	} else if kb.KeyDown(hid.KeyboardKeyG) {
		ed.transformTool.Enable(transform_tools.ToolStateMove)
	} else if kb.KeyDown(hid.KeyboardKeyR) {
		ed.transformTool.Enable(transform_tools.ToolStateRotate)
	} else if kb.KeyDown(hid.KeyboardKeyS) {
		ed.transformTool.Enable(transform_tools.ToolStateScale)
	} else if kb.KeyDown(hid.KeyboardKeyDelete) {
		deleter.DeleteSelected(ed)
	}
}
