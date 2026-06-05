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

func (p *shaderGraphPort) bindEvents() {
	if p == nil || p.dot == nil || p.graph == nil {
		return
	}
	p.dot.Base().AddEvent(ui.EventTypeDown, func() {
		if p.graph.isPanInputHeld() {
			return
		}
		p.graph.beginConnection(p)
	})
	p.dot.Base().AddEvent(ui.EventTypeUp, func() {
		p.graph.finishConnection(p)
	})
	p.dot.Base().AddEvent(ui.EventTypeDragEnd, func() {
		p.graph.finishConnection(p)
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

func shaderGraphPortTypeKey(portType string) string {
	return strings.ToLower(strings.TrimSpace(portType))
}
