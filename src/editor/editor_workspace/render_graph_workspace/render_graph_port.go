/******************************************************************************/
/* shader_graph_port.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"strings"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
)

type shaderGraphPort struct {
	graph       *shaderGraph
	node        *shaderGraphNode
	spec        shaderGraphPortSpec
	output      bool
	index       int
	hit         *ui.Panel
	dot         *ui.Panel
	label       *ui.Label
	localAnchor matrix.Vec2
}

func (p *shaderGraphPort) Anchor() matrix.Vec2 {
	if p == nil || p.node == nil {
		return matrix.Vec2Zero()
	}
	return p.node.position.Add(p.localAnchor)
}

func (p *shaderGraphPort) CanConnect(other *shaderGraphPort) bool {
	return shaderGraphPortsCanConnect(p, other)
}

func (p *shaderGraphPort) Color() matrix.Color {
	if p == nil {
		return matrix.ColorWhite()
	}
	return shaderGraphPortColor(p.spec.Type, p.output)
}

func shaderGraphPortRef(port *shaderGraphPort) (RenderGraphPortRef, bool) {
	if port == nil || port.node == nil || port.node.id == "" {
		return RenderGraphPortRef{}, false
	}
	return RenderGraphPortRef{Node: port.node.id, Port: port.index}, true
}

func (p *shaderGraphPort) bindEvents() {
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

func (p *shaderGraphPort) bindTargetEvents(target *ui.UI) {
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

func shaderGraphPortsCanConnect(a, b *shaderGraphPort) bool {
	return a != nil && b != nil &&
		a.output != b.output &&
		shaderGraphPortTypeKey(a.spec.Type) == shaderGraphPortTypeKey(b.spec.Type)
}

func shaderGraphConnectionPorts(a, b *shaderGraphPort) (output, input *shaderGraphPort, ok bool) {
	if !shaderGraphPortsCanConnect(a, b) {
		return nil, nil, false
	}
	if a.output {
		return a, b, true
	}
	return b, a, true
}

func shaderGraphFirstCompatibleNodePort(node *shaderGraphNode, source *shaderGraphPort) *shaderGraphPort {
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

func shaderGraphPortTypeKey(portType string) string {
	return strings.ToLower(strings.TrimSpace(portType))
}
