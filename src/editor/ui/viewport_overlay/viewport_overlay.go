/******************************************************************************/
/* viewport_overlay.go                                                        */
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

package viewport_overlay

import (
	"kaiju/editor/editor_interface"
	"kaiju/editor/viewport/controls"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/matrix"
)

type ViewportOverlay struct {
	ed  editor_interface.Editor
	doc *document.Document
}

func (v *ViewportOverlay) updateSelectCameraModeColors(e *document.Element) {
	modePanels := v.doc.GetElementsByGroup("camMode")
	for i := range modePanels {
		modePanels[i].UI.ToPanel().SetColor(matrix.ColorBlack())
		modePanels[i].Children[0].UI.ToLabel().SetBGColor(matrix.ColorBlack())
	}
	e.UI.ToPanel().SetColor(matrix.ColorDarkBG())
	e.Children[0].UI.ToLabel().SetBGColor(matrix.ColorDarkBG())
}

func (v *ViewportOverlay) setCameraMode3d(e *document.Element) {
	v.ed.Camera().SetMode(controls.EditorCameraMode3d, v.ed.Host())
	v.updateSelectCameraModeColors(e)
}

func (v *ViewportOverlay) setCameraMode2d(e *document.Element) {
	v.ed.Camera().SetMode(controls.EditorCameraMode2d, v.ed.Host())
	v.updateSelectCameraModeColors(e)
}

func New(ed editor_interface.Editor, uiMan *ui.Manager) {
	const html = "editor/ui/viewport_overlay/viewport.html"
	v := &ViewportOverlay{ed, nil}
	host := ed.Host()
	host.CreatingEditorEntities()
	v.doc, _ = markup.DocumentFromHTMLAsset(uiMan, html, nil, map[string]func(*document.Element){
		"setCameraMode3d": v.setCameraMode3d,
		"setCameraMode2d": v.setCameraMode2d,
	})
	host.DoneCreatingEditorEntities()
	ed.Camera().OnModeChange.Add(func() {
		switch ed.Camera().Mode() {
		case controls.EditorCameraMode3d:
			if e, ok := v.doc.GetElementById("camMode3d"); ok {
				v.updateSelectCameraModeColors(e)
			}
		case controls.EditorCameraMode2d:
			if e, ok := v.doc.GetElementById("camMode2d"); ok {
				v.updateSelectCameraModeColors(e)
			}
		}
	})
}
