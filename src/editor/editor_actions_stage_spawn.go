/******************************************************************************/
/* editor_actions_stage_spawn.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"fmt"
	"strings"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_workspace/stage_workspace"
	"kaijuengine.com/rendering"
)

const (
	ActionStageSpawnEntity    editor_action.ActionID = "stage.spawnEntity"
	ActionStageSpawnCamera    editor_action.ActionID = "stage.spawnCamera"
	ActionStageSpawnLight     editor_action.ActionID = "stage.spawnLight"
	ActionStageSpawnPrimitive editor_action.ActionID = "stage.spawnPrimitive"
)

type primitiveActionArgs struct {
	Primitive string `json:"primitive"`
}

func init() {
	registerEditorActionProvider(registerStageSpawnActions)
}

func registerStageSpawnActions(ed *Editor, mustRegister editorActionRegistrar) {
	mustRegister(editor_action.Definition{
		ID:                ActionStageSpawnEntity,
		Label:             "Spawn Entity",
		Description:       "Creates an empty entity at the stage view focus point.",
		Category:          "Stage",
		Tags:              []string{"create", "new", "entity"},
		UndoPolicy:        editor_action.UndoPolicyTransaction,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionSpawnEntity, ed.stageCanRun)
	mustRegister(editor_action.Definition{
		ID:                ActionStageSpawnCamera,
		Label:             "Spawn Camera",
		Description:       "Creates a camera entity.",
		Category:          "Stage",
		Tags:              []string{"create", "new", "camera"},
		UndoPolicy:        editor_action.UndoPolicyTransaction,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionSpawnCamera, ed.stageCanRun)
	mustRegister(editor_action.Definition{
		ID:                ActionStageSpawnLight,
		Label:             "Spawn Light",
		Description:       "Creates a light entity.",
		Category:          "Stage",
		Tags:              []string{"create", "new", "light"},
		UndoPolicy:        editor_action.UndoPolicyTransaction,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionSpawnLight, ed.stageCanRun)
	mustRegister(editor_action.Definition{
		ID:            ActionStageSpawnPrimitive,
		Label:         "Spawn Primitive",
		Description:   "Creates a primitive mesh entity.",
		Category:      "Stage",
		Tags:          []string{"create", "new", "primitive", "mesh"},
		DefaultParams: editor_action.Params(primitiveActionArgs{Primitive: "cube"}),
		NewParams:     func() any { return &primitiveActionArgs{} },
		Parameters: []editor_action.Parameter{{
			Name: "primitive", Label: "Primitive", Type: "string", Required: true,
			Options: []string{"cube", "sphere", "plane", "capsule", "cylinder", "cone", "arrow"},
		}},
		UndoPolicy:        editor_action.UndoPolicyManaged,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
		Variants: []editor_action.Variant{
			{Label: "Spawn Cube", Tags: []string{"cube"}, Params: editor_action.Params(primitiveActionArgs{Primitive: "cube"})},
			{Label: "Spawn Sphere", Tags: []string{"sphere"}, Params: editor_action.Params(primitiveActionArgs{Primitive: "sphere"})},
			{Label: "Spawn Plane", Tags: []string{"plane"}, Params: editor_action.Params(primitiveActionArgs{Primitive: "plane"})},
			{Label: "Spawn Capsule", Tags: []string{"capsule"}, Params: editor_action.Params(primitiveActionArgs{Primitive: "capsule"})},
			{Label: "Spawn Cylinder", Tags: []string{"cylinder"}, Params: editor_action.Params(primitiveActionArgs{Primitive: "cylinder"})},
			{Label: "Spawn Cone", Tags: []string{"cone"}, Params: editor_action.Params(primitiveActionArgs{Primitive: "cone"})},
			{Label: "Spawn Arrow", Tags: []string{"arrow"}, Params: editor_action.Params(primitiveActionArgs{Primitive: "arrow"})},
		},
	}, ed.actionSpawnPrimitive, ed.stageCanRun)
}

func (ed *Editor) actionSpawnEntity(editor_action.Context, editor_action.Request) editor_action.Result {
	e, ok := ed.StageWorkspace().CreateNewEntity()
	if !ok {
		return editor_action.Failure("failed to create entity")
	}
	man := ed.stageView.Manager()
	man.ClearSelection()
	man.SelectEntity(e)
	return stageResult("entity created", e, man.Selection())
}

func (ed *Editor) actionSpawnCamera(editor_action.Context, editor_action.Request) editor_action.Result {
	e, ok := ed.StageWorkspace().CreateNewCamera()
	if !ok {
		return editor_action.Failure("failed to create camera")
	}
	return stageResult("camera created", e, ed.stageView.Manager().Selection())
}

func (ed *Editor) actionSpawnLight(editor_action.Context, editor_action.Request) editor_action.Result {
	e, ok := ed.StageWorkspace().CreateNewLight()
	if !ok {
		return editor_action.Failure("failed to create light")
	}
	return stageResult("light created", e, ed.stageView.Manager().Selection())
}

func (ed *Editor) actionSpawnPrimitive(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	args := primitiveActionArgs{Primitive: "cube"}
	if params, ok := editor_action.Param[primitiveActionArgs](req); ok {
		args = params
	}
	primitive, ok := parsePrimitive(args.Primitive)
	if !ok {
		return editor_action.Failure("primitive must be cube, sphere, plane, capsule, cylinder, cone, or arrow")
	}
	e, ok := ed.StageWorkspace().CreatePrimitive(primitive)
	if !ok {
		return editor_action.Failure(fmt.Sprintf("failed to create %s", args.Primitive))
	}
	return stageResult("primitive created", e, ed.stageView.Manager().Selection())
}

func parsePrimitive(value string) (rendering.PrimitiveMesh, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "cube", "texturable_cube":
		return rendering.PrimitiveMeshTexturableCube, true
	case "sphere", string(rendering.PrimitiveMeshSphere):
		return rendering.PrimitiveMeshSphere, true
	case "plane":
		return rendering.PrimitiveMeshPlane, true
	case "capsule", string(rendering.PrimitiveMeshCapsule):
		return rendering.PrimitiveMeshCapsule, true
	case "cylinder", string(rendering.PrimitiveMeshCylinder):
		return rendering.PrimitiveMeshCylinder, true
	case "cone", string(rendering.PrimitiveMeshCone):
		return rendering.PrimitiveMeshCone, true
	case "arrow", string(rendering.PrimitiveMeshArrow):
		return rendering.PrimitiveMeshArrow, true
	default:
		return rendering.PrimitiveMeshTexturableCube, false
	}
}
