/******************************************************************************/
/* schema_workspace_layout.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

const (
	schemaWorkspaceMenuBarHeight   = float32(24.0)
	schemaWorkspaceStatusBarHeight = float32(20.8)
	schemaWorkspaceActionBarHeight = float32(48.0)
)

func (w *SchemaWorkspace) applyLayout() {
	if w.Host == nil || w.Host.Window == nil {
		return
	}
	windowWidth := float32(w.Host.Window.Width())
	windowHeight := float32(w.Host.Window.Height())
	contentHeight := max(1, windowHeight-schemaWorkspaceMenuBarHeight-schemaWorkspaceStatusBarHeight)
	canvasHeight := max(1, contentHeight-schemaWorkspaceActionBarHeight)
	w.graph.SetViewport(0, schemaWorkspaceMenuBarHeight, windowWidth, canvasHeight)
}
