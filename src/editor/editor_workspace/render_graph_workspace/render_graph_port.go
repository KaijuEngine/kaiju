/******************************************************************************/
/* render_graph_port.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"strings"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
)

type renderGraphPort struct {
	graph       *renderGraph
	node        *renderGraphNode
	spec        renderGraphPortSpec
	output      bool
	index       int
	hit         *ui.Panel
	dot         *ui.Panel
	label       *ui.Label
	localAnchor matrix.Vec2
}

func (p *renderGraphPort) Anchor() matrix.Vec2 {
	if p == nil || p.node == nil {
		return matrix.Vec2Zero()
	}
	return p.node.position.Add(p.localAnchor)
}

func (p *renderGraphPort) CanConnect(other *renderGraphPort) bool {
	return renderGraphPortsCanConnect(p, other)
}

func (p *renderGraphPort) Color() matrix.Color {
	if p == nil {
		return matrix.ColorWhite()
	}
	return renderGraphPortColor(p.spec.Type, p.output)
}

func renderGraphPortRef(port *renderGraphPort) (RenderGraphPortRef, bool) {
	if port == nil || port.node == nil || port.node.id == "" {
		return RenderGraphPortRef{}, false
	}
	return RenderGraphPortRef{Node: port.node.id, Port: port.index}, true
}

func (p *renderGraphPort) bindEvents() {
	if p == nil || p.graph == nil {
		return
	}
	if p.hit != nil {
		p.bindTargetEvents(p.hit.Base())
		return
	}
	if p.dot != nil {
		p.bindTargetEvents(p.dot.Base())
	}
	if p.label != nil {
		p.bindTargetEvents(p.label.Base())
	}
}

func (p *renderGraphPort) bindTargetEvents(target *ui.UI) {
	if target == nil || p == nil || p.graph == nil {
		return
	}
	target.AddEvent(ui.EventTypeDown, func() {
		if p.graph.isPanInputHeld() {
			return
		}
		if p.graph.isAltInputHeld() {
			p.graph.DisconnectPort(p)
			return
		}
		p.graph.beginConnection(p)
	})
	target.AddEvent(ui.EventTypeUp, func() {
		p.graph.finishConnection(p)
	})
	target.AddEvent(ui.EventTypeDragEnd, func() {
		p.graph.finishConnection(nil)
	})
}

func renderGraphPortsCanConnect(a, b *renderGraphPort) bool {
	return a != nil && b != nil &&
		a.output != b.output &&
		renderGraphPortTypeKey(a.spec.Type) == renderGraphPortTypeKey(b.spec.Type)
}

func renderGraphConnectionPorts(a, b *renderGraphPort) (output, input *renderGraphPort, ok bool) {
	if !renderGraphPortsCanConnect(a, b) {
		return nil, nil, false
	}
	if a.output {
		return a, b, true
	}
	return b, a, true
}

func renderGraphFirstCompatibleNodePort(node *renderGraphNode, source *renderGraphPort) *renderGraphPort {
	if node == nil || source == nil {
		return nil
	}
	if source.output {
		for i := range node.inputs {
			if source.CanConnect(node.inputs[i]) {
				return node.inputs[i]
			}
		}
		return nil
	}
	for i := range node.outputs {
		if source.CanConnect(node.outputs[i]) {
			return node.outputs[i]
		}
	}
	return nil
}

func renderGraphPortTypeKey(portType string) string {
	return strings.ToLower(strings.TrimSpace(portType))
}
