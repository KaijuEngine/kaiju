package viewport_overlay

import (
	"kaiju/editor/interfaces"
	"kaiju/editor/viewport/controls"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/matrix"
	"kaiju/engine/ui"
)

type ViewportOverlay struct {
	ed  interfaces.Editor
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

func New(ed interfaces.Editor, uiMan *ui.Manager) {
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
