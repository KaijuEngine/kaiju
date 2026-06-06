/******************************************************************************/
/* render_graph_workspace_actions.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/matrix"
)

const (
	ActionRenderGraphShowCreateNodeMenu editor_action.ActionID = "renderGraph.showCreateNodeMenu"
	ActionRenderGraphCreateNode         editor_action.ActionID = "renderGraph.createNode"
	ActionRenderGraphCreateComment      editor_action.ActionID = "renderGraph.createComment"
	ActionRenderGraphCenterView         editor_action.ActionID = "renderGraph.centerView"
	ActionRenderGraphFocusSelection     editor_action.ActionID = "renderGraph.focusSelection"
	ActionRenderGraphSave               editor_action.ActionID = "renderGraph.save"
	ActionRenderGraphDeleteSelection    editor_action.ActionID = "renderGraph.deleteSelection"
)

type CreateNodeActionArgs struct {
	NodeID            string  `json:"nodeId"`
	X                 float32 `json:"x,omitempty"`
	Y                 float32 `json:"y,omitempty"`
	UsePosition       bool    `json:"usePosition,omitempty"`
	UseConnection     bool    `json:"useConnection,omitempty"`
	ConnectFromNodeID string  `json:"connectFromNodeId,omitempty"`
	ConnectFromPort   int     `json:"connectFromPort,omitempty"`
	ConnectFromOutput bool    `json:"connectFromOutput,omitempty"`
	ConnectFromType   string  `json:"connectFromType,omitempty"`
}

type CreateCommentActionArgs struct {
	Label       string  `json:"label,omitempty"`
	X           float32 `json:"x,omitempty"`
	Y           float32 `json:"y,omitempty"`
	Width       float32 `json:"width,omitempty"`
	Height      float32 `json:"height,omitempty"`
	UsePosition bool    `json:"usePosition,omitempty"`
	UseSize     bool    `json:"useSize,omitempty"`
}

func DefaultCreateNodeActionArgs() CreateNodeActionArgs {
	catalog := renderGraphNodeCatalog()
	if len(catalog) == 0 {
		return CreateNodeActionArgs{}
	}
	return CreateNodeActionArgs{NodeID: catalog[0].ID}
}

func DefaultCreateCommentActionArgs() CreateCommentActionArgs {
	return CreateCommentActionArgs{
		Label:  "Comment",
		Width:  renderGraphCommentDefaultWidth,
		Height: renderGraphCommentDefaultHeight,
	}
}

func CreateNodeActionVariants() []editor_action.Variant {
	catalog := renderGraphNodeCatalog()
	variants := make([]editor_action.Variant, 0, len(catalog))
	for i := range catalog {
		entry := catalog[i]
		variants = append(variants, editor_action.Variant{
			Label:       "Create " + entry.Name + " Node",
			Description: entry.Description,
			Tags:        append([]string{"render", "graph", "node", "create"}, entry.Tags...),
			Params:      editor_action.Params(CreateNodeActionArgs{NodeID: entry.ID}),
		})
	}
	return variants
}

func (a CreateNodeActionArgs) position(fallback matrix.Vec2) matrix.Vec2 {
	if !a.UsePosition {
		return fallback
	}
	return matrix.NewVec2(matrix.Float(a.X), matrix.Float(a.Y))
}

func (a CreateCommentActionArgs) position(fallback matrix.Vec2) matrix.Vec2 {
	if !a.UsePosition {
		return fallback
	}
	return matrix.NewVec2(matrix.Float(a.X), matrix.Float(a.Y))
}

func (a CreateCommentActionArgs) size() matrix.Vec2 {
	if !a.UseSize {
		return matrix.NewVec2(
			matrix.Float(renderGraphCommentDefaultWidth),
			matrix.Float(renderGraphCommentDefaultHeight),
		)
	}
	return renderGraphCommentSizeOrDefault(matrix.NewVec2(matrix.Float(a.Width), matrix.Float(a.Height)))
}
