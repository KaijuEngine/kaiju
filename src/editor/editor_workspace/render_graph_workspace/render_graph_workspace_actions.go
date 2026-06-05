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
	ActionRenderGraphCenterView         editor_action.ActionID = "renderGraph.centerView"
	ActionRenderGraphFocusSelection     editor_action.ActionID = "renderGraph.focusSelection"
)

type CreateNodeActionArgs struct {
	NodeID      string  `json:"nodeId"`
	X           float32 `json:"x,omitempty"`
	Y           float32 `json:"y,omitempty"`
	UsePosition bool    `json:"usePosition,omitempty"`
}

func DefaultCreateNodeActionArgs() CreateNodeActionArgs {
	catalog := shaderGraphNodeCatalog()
	if len(catalog) == 0 {
		return CreateNodeActionArgs{}
	}
	return CreateNodeActionArgs{NodeID: catalog[0].ID}
}

func CreateNodeActionVariants() []editor_action.Variant {
	catalog := shaderGraphNodeCatalog()
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
