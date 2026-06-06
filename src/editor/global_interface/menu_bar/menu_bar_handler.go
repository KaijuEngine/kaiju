/******************************************************************************/
/* menu_bar_handler.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package menu_bar

import (
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/rendering"
)

// WorkspaceTab is the data the menu bar template needs to render one tab in
// the workspace tab strip. The list is supplied by the editor on initial load
// and again after workspace registration / settings changes.
type WorkspaceTab struct {
	ID          string
	DisplayName string
}

type MenuBarHandler interface {
	BlurInterface()
	FocusInterface()
	Settings() *editor_settings.Settings
	Events() *editor_events.EditorEvents
	History() *memento.History
	Project() *project.Project
	ProjectFileSystem() *project_file_system.FileSystem
	// WorkspaceSelected is invoked when the user clicks a tab. The id is
	// the workspace id supplied via WorkspaceTab.ID, which is the same id
	// the workspace itself registered under.
	WorkspaceSelected(id string)
	StageView() *editor_stage_view.StageView
	Build(buildMode project.GameBuildMode)
	BuildAndRun(buildMode project.GameBuildMode)
	BuildAndRunCurrentStage()
	OpenCodeEditor()
	CreateNewStage()
	SaveCurrentStage()
	CreateNewCamera()
	CreateNewEntity()
	CreateNewLight()
	CreatePrimitive(primitive rendering.PrimitiveMesh)
	ConnectSelectedAsDistanceChain()
	ConnectSelectedAsRope()
	ConnectSelectedAsHingeChain()
	CreatePluginProject(path string)
	CreateHtmlUiFile(name string)
	CreateCssStylesheetFile(name string)
	SetGridVisible(visible bool)
}
