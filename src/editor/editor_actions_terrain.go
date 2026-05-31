/******************************************************************************/
/* editor_actions_terrain.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_workspace/terrain_workspace"
	"kaijuengine.com/platform/hid"
)

const (
	ActionTerrainDecreaseBrushRadius   editor_action.ActionID = "terrain.decreaseBrushRadius"
	ActionTerrainIncreaseBrushRadius   editor_action.ActionID = "terrain.increaseBrushRadius"
	ActionTerrainDecreaseBrushStrength editor_action.ActionID = "terrain.decreaseBrushStrength"
	ActionTerrainIncreaseBrushStrength editor_action.ActionID = "terrain.increaseBrushStrength"
)

func init() {
	registerEditorActionProvider(registerTerrainActions)
}

func registerTerrainActions(ed *Editor, mustRegister editorActionRegistrar) {
	mustRegister(editor_action.Definition{
		ID:          ActionTerrainDecreaseBrushRadius,
		Label:       "Decrease Brush Radius",
		Description: "Reduces the active terrain brush radius.",
		Category:    "Terrain",
		Tags:        []string{"terrain", "brush", "radius", "decrease"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionTerrainDecreaseBrushRadius,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyOpenBracket)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: terrain_workspace.ID,
	}, ed.actionTerrainDecreaseBrushRadius, ed.terrainCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionTerrainIncreaseBrushRadius,
		Label:       "Increase Brush Radius",
		Description: "Increases the active terrain brush radius.",
		Category:    "Terrain",
		Tags:        []string{"terrain", "brush", "radius", "increase"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionTerrainIncreaseBrushRadius,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyCloseBracket)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: terrain_workspace.ID,
	}, ed.actionTerrainIncreaseBrushRadius, ed.terrainCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionTerrainDecreaseBrushStrength,
		Label:       "Decrease Brush Strength",
		Description: "Reduces the active terrain brush strength or texture opacity.",
		Category:    "Terrain",
		Tags:        []string{"terrain", "brush", "strength", "opacity", "decrease"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionTerrainDecreaseBrushStrength,
			Enabled: true,
			Chord: editor_action.KeyChord{
				Keys:  []int{int(hid.KeyboardKeyOpenBracket)},
				Shift: true,
			},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: terrain_workspace.ID,
	}, ed.actionTerrainDecreaseBrushStrength, ed.terrainCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionTerrainIncreaseBrushStrength,
		Label:       "Increase Brush Strength",
		Description: "Increases the active terrain brush strength or texture opacity.",
		Category:    "Terrain",
		Tags:        []string{"terrain", "brush", "strength", "opacity", "increase"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionTerrainIncreaseBrushStrength,
			Enabled: true,
			Chord: editor_action.KeyChord{
				Keys:  []int{int(hid.KeyboardKeyCloseBracket)},
				Shift: true,
			},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: terrain_workspace.ID,
	}, ed.actionTerrainIncreaseBrushStrength, ed.terrainCanRun)
}

func (ed *Editor) terrainCanRun(editor_action.Context, editor_action.Request) editor_action.Result {
	if _, ok := ed.terrainWorkspace(); !ok {
		return editor_action.Failure("terrain workspace is not available")
	}
	return editor_action.Success("")
}

func (ed *Editor) actionTerrainDecreaseBrushRadius(editor_action.Context, editor_action.Request) editor_action.Result {
	return ed.actionTerrainBrushRadius(-1)
}

func (ed *Editor) actionTerrainIncreaseBrushRadius(editor_action.Context, editor_action.Request) editor_action.Result {
	return ed.actionTerrainBrushRadius(1)
}

func (ed *Editor) actionTerrainBrushRadius(direction int) editor_action.Result {
	w, ok := ed.terrainWorkspace()
	if !ok {
		return editor_action.Failure("terrain workspace is not available")
	}
	w.AdjustBrushRadius(direction)
	return editor_action.Success("terrain brush radius changed")
}

func (ed *Editor) actionTerrainDecreaseBrushStrength(editor_action.Context, editor_action.Request) editor_action.Result {
	return ed.actionTerrainBrushStrength(-1)
}

func (ed *Editor) actionTerrainIncreaseBrushStrength(editor_action.Context, editor_action.Request) editor_action.Result {
	return ed.actionTerrainBrushStrength(1)
}

func (ed *Editor) actionTerrainBrushStrength(direction int) editor_action.Result {
	w, ok := ed.terrainWorkspace()
	if !ok {
		return editor_action.Failure("terrain workspace is not available")
	}
	w.AdjustBrushStrength(direction)
	return editor_action.Success("terrain brush strength changed")
}

func (ed *Editor) terrainWorkspace() (*terrain_workspace.TerrainWorkspace, bool) {
	w, ok := ed.Workspace(terrain_workspace.ID)
	if !ok {
		return nil, false
	}
	tw, ok := w.(*terrain_workspace.TerrainWorkspace)
	return tw, ok
}
