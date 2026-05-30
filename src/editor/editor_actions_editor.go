/******************************************************************************/
/* editor_actions_editor.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/platform/hid"
)

const (
	ActionEditorSaveStage   editor_action.ActionID = "editor.saveStage"
	ActionEditorUndo        editor_action.ActionID = "editor.undo"
	ActionEditorRedo        editor_action.ActionID = "editor.redo"
	ActionEditorOpenPalette editor_action.ActionID = "editor.openActionPalette"
)

func init() {
	registerEditorActionProvider(registerEditorActions)
}

func registerEditorActions(ed *Editor, mustRegister editorActionRegistrar) {
	mustRegister(editor_action.Definition{
		ID:          ActionEditorSaveStage,
		Label:       "Save Stage",
		Description: "Saves the currently open stage.",
		Category:    "Editor",
		Tags:        []string{"save", "stage", "file"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionEditorSaveStage,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyS)}, CtrlOrMeta: true},
		}},
		UndoPolicy: editor_action.UndoPolicyNone,
		Visible:    true,
	}, ed.actionSaveStage, nil)
	mustRegister(editor_action.Definition{
		ID:          ActionEditorUndo,
		Label:       "Undo",
		Description: "Undoes the most recent editor action.",
		Category:    "Editor",
		Tags:        []string{"history"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionEditorUndo,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyZ)}, CtrlOrMeta: true},
		}},
		UndoPolicy: editor_action.UndoPolicyNone,
		Visible:    true,
	}, ed.actionUndo, nil)
	mustRegister(editor_action.Definition{
		ID:          ActionEditorRedo,
		Label:       "Redo",
		Description: "Redoes the next editor action.",
		Category:    "Editor",
		Tags:        []string{"history"},
		DefaultBindings: []editor_action.ActionBinding{
			{
				Action:  ActionEditorRedo,
				Enabled: true,
				Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyY)}, CtrlOrMeta: true},
			},
			{
				Action:  ActionEditorRedo,
				Enabled: true,
				Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyZ)}, CtrlOrMeta: true, Shift: true},
			},
		},
		UndoPolicy: editor_action.UndoPolicyNone,
		Visible:    true,
	}, ed.actionRedo, nil)
	mustRegister(editor_action.Definition{
		ID:          ActionEditorOpenPalette,
		Label:       "Open Action Palette",
		Description: "Opens the searchable action palette.",
		Category:    "Editor",
		Tags:        []string{"search", "command", "palette"},
		UndoPolicy:  editor_action.UndoPolicyNone,
		Visible:     true,
	}, ed.actionOpenPalette, nil)
}

func (ed *Editor) actionSaveStage(editor_action.Context, editor_action.Request) editor_action.Result {
	ed.SaveCurrentStage()
	return editor_action.Success("stage save requested")
}

func (ed *Editor) actionUndo(editor_action.Context, editor_action.Request) editor_action.Result {
	ed.history.Undo()
	return editor_action.Success("undo")
}

func (ed *Editor) actionRedo(editor_action.Context, editor_action.Request) editor_action.Result {
	ed.history.Redo()
	return editor_action.Success("redo")
}

func (ed *Editor) actionOpenPalette(editor_action.Context, editor_action.Request) editor_action.Result {
	ed.showActionPalette()
	return editor_action.Success("action palette opened")
}
