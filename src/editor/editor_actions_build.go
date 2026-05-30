/******************************************************************************/
/* editor_actions_build.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"strings"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/platform/hid"
)

const (
	ActionEditorBuild       editor_action.ActionID = "editor.build"
	ActionEditorBuildAndRun editor_action.ActionID = "editor.buildAndRun"
)

type buildActionArgs struct {
	Mode string `json:"mode"`
}

type buildAndRunActionArgs struct {
	Mode         string `json:"mode"`
	CurrentStage bool   `json:"currentStage"`
}

func init() {
	registerEditorActionProvider(registerBuildActions)
}

func registerBuildActions(ed *Editor, mustRegister editorActionRegistrar) {
	mustRegister(editor_action.Definition{
		ID:            ActionEditorBuild,
		Label:         "Build Game",
		Description:   "Builds the game project.",
		Category:      "Editor",
		Tags:          []string{"build", "compile"},
		DefaultParams: editor_action.Params(buildActionArgs{Mode: "debug"}),
		NewParams:     func() any { return &buildActionArgs{} },
		Parameters: []editor_action.Parameter{
			{Name: "mode", Label: "Mode", Type: "string", Options: []string{"debug", "release"}},
		},
		UndoPolicy: editor_action.UndoPolicyNone,
		Visible:    true,
		Variants: []editor_action.Variant{
			{Label: "Build Game", Tags: []string{"debug"}, Params: editor_action.Params(buildActionArgs{Mode: "debug"})},
			{Label: "Build Game (Release)", Tags: []string{"release"}, Params: editor_action.Params(buildActionArgs{Mode: "release"})},
		},
	}, ed.actionBuild, nil)
	mustRegister(editor_action.Definition{
		ID:            ActionEditorBuildAndRun,
		Label:         "Build And Run Game",
		Description:   "Builds and runs the game project.",
		Category:      "Editor",
		Tags:          []string{"build", "run", "play"},
		DefaultParams: editor_action.Params(buildAndRunActionArgs{Mode: "debug"}),
		NewParams:     func() any { return &buildAndRunActionArgs{} },
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionEditorBuildAndRun,
			Params:  editor_action.Params(buildAndRunActionArgs{CurrentStage: true}),
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyF5)}},
		}, {
			Action:  ActionEditorBuildAndRun,
			Params:  editor_action.Params(buildAndRunActionArgs{Mode: "debug"}),
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyF5)}, CtrlOrMeta: true},
		}, {
			Action:  ActionEditorBuildAndRun,
			Params:  editor_action.Params(buildAndRunActionArgs{Mode: "release"}),
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyF5)}, CtrlOrMeta: true, Shift: true},
		}},
		Parameters: []editor_action.Parameter{
			{Name: "mode", Label: "Mode", Type: "string", Options: []string{"debug", "release"}},
			{Name: "currentStage", Label: "Current Stage", Type: "bool"},
		},
		UndoPolicy: editor_action.UndoPolicyNone,
		Visible:    true,
		Variants: []editor_action.Variant{
			{Label: "Build And Run Game", Tags: []string{"debug"}, Params: editor_action.Params(buildAndRunActionArgs{Mode: "debug"})},
			{Label: "Build And Run Game (Release)", Tags: []string{"release"}, Params: editor_action.Params(buildAndRunActionArgs{Mode: "release"})},
			{Label: "Run Current Stage", Tags: []string{"debug", "stage"}, Params: editor_action.Params(buildAndRunActionArgs{CurrentStage: true})},
		},
	}, ed.actionBuildAndRun, nil)
}

func (ed *Editor) actionBuild(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	args := buildActionArgs{Mode: "debug"}
	if params, ok := editor_action.Param[buildActionArgs](req); ok {
		args = params
	}
	mode, ok := parseBuildMode(args.Mode)
	if !ok {
		return editor_action.Failure("mode must be debug or release")
	}
	ed.Build(mode)
	return editor_action.Success("build requested")
}

func (ed *Editor) actionBuildAndRun(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	args := buildAndRunActionArgs{Mode: "debug"}
	if params, ok := editor_action.Param[buildAndRunActionArgs](req); ok {
		args = params
	}
	if args.CurrentStage {
		ed.BuildAndRunCurrentStage()
		return editor_action.Success("current stage run requested")
	}
	mode, ok := parseBuildMode(args.Mode)
	if !ok {
		return editor_action.Failure("mode must be debug or release")
	}
	ed.BuildAndRun(mode)
	return editor_action.Success("build and run requested")
}

func parseBuildMode(mode string) (project.GameBuildMode, bool) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "debug":
		return project.GameBuildModeDebug, true
	case "release":
		return project.GameBuildModeRelease, true
	default:
		return project.GameBuildModeDebug, false
	}
}
