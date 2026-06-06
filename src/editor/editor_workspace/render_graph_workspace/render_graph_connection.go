/******************************************************************************/
/* render_graph_connection.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"fmt"
	"log/slog"
	"sync/atomic"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const (
	shaderGraphSplineSegments = 48
	shaderGraphSplineStroke   = matrix.Float(2.0)
	shaderGraphSplineDepth    = matrix.Float(4.35)
)

var shaderGraphSplineID atomic.Uint64

type shaderGraphConnection struct {
	output *shaderGraphPort
	input  *shaderGraphPort
	visual shaderGraphSpline
}

func (c *shaderGraphConnection) Initialize(host *engine.Host, parent *ui.Panel, output, input *shaderGraphPort) {
	c.output = output
	c.input = input
	c.visual.Initialize(host, parent)
	c.Update()
}

func (c *shaderGraphConnection) Update() {
	if c.output == nil || c.input == nil {
		c.visual.Hide()
		return
	}
	c.visual.Show()
	c.visual.SetColor(c.output.Color())
	graph := c.output.graph
	if graph == nil {
		graph = c.input.graph
	}
	start := c.output.Anchor()
	end := c.input.Anchor()
	if graph != nil {
		start = graph.viewPosition(start)
		end = graph.viewPosition(end)
	}
	c.visual.Update(start, end)
}

func (c *shaderGraphConnection) Hide() {
	c.visual.Hide()
}

func (c *shaderGraphConnection) Destroy() {
	c.visual.Destroy()
}

func (c *shaderGraphConnection) matches(outputRef, inputRef RenderGraphPortRef) bool {
	if c == nil {
		return false
	}
	currentOutput, currentInput, ok := shaderGraphConnectionRefs(c.output, c.input)
	return ok && currentOutput == outputRef && currentInput == inputRef
}

func (c *shaderGraphConnection) renderConnection() (RenderGraphConnection, bool) {
	if c == nil {
		return RenderGraphConnection{}, false
	}
	outputRef, inputRef, ok := shaderGraphConnectionRefs(c.output, c.input)
	if !ok {
		return RenderGraphConnection{}, false
	}
	return RenderGraphConnection{Output: outputRef, Input: inputRef}, true
}

func (c *shaderGraphConnection) touchesPort(port *shaderGraphPort) bool {
	portRef, ok := shaderGraphPortRef(port)
	if c == nil || !ok {
		return false
	}
	outputRef, inputRef, ok := shaderGraphConnectionRefs(c.output, c.input)
	if !ok {
		return false
	}
	if port.output {
		return outputRef == portRef
	}
	return inputRef == portRef
}

func (c *shaderGraphConnection) touchesNode(id string) bool {
	if c == nil || id == "" {
		return false
	}
	outputRef, inputRef, ok := shaderGraphConnectionRefs(c.output, c.input)
	return ok && (outputRef.Node == id || inputRef.Node == id)
}

func shaderGraphConnectionRefs(output, input *shaderGraphPort) (RenderGraphPortRef, RenderGraphPortRef, bool) {
	if output == nil || input == nil ||
		output.node == nil || input.node == nil ||
		output.node.id == "" || input.node.id == "" {
		return RenderGraphPortRef{}, RenderGraphPortRef{}, false
	}
	return RenderGraphPortRef{
			Node: output.node.id,
			Port: output.index,
		}, RenderGraphPortRef{
			Node: input.node.id,
			Port: input.index,
		}, true
}

type shaderGraphSpline struct {
	host      *engine.Host
	root      *ui.Panel
	mesh      *rendering.Mesh
	shader    *ui.ShaderData
	transform matrix.Transform
	verts     []rendering.Vertex
	points    []matrix.Vec2
	key       string
	color     matrix.Color
}

func (s *shaderGraphSpline) Initialize(host *engine.Host, parent *ui.Panel) {
	if host == nil || parent == nil {
		return
	}
	s.host = host
	s.root = parent
	s.color = matrix.ColorWhite()
	s.key = fmt.Sprintf("editor_shading_graph_spline_%d", shaderGraphSplineID.Add(1))
	s.verts = make([]rendering.Vertex, (shaderGraphSplineSegments+1)*2)
	s.points = make([]matrix.Vec2, shaderGraphSplineSegments+1)
	indexes := make([]uint32, shaderGraphSplineSegments*6)
	for i := 0; i < shaderGraphSplineSegments; i++ {
		startBottom := uint32(i * 2)
		startTop := startBottom + 1
		endBottom := uint32((i + 1) * 2)
		endTop := endBottom + 1
		ii := i * 6
		indexes[ii+0] = startBottom
		indexes[ii+1] = endTop
		indexes[ii+2] = startTop
		indexes[ii+3] = startBottom
		indexes[ii+4] = endBottom
		indexes[ii+5] = endTop
	}
	s.hideVertices()
	s.mesh = host.MeshCache().DynamicMesh(s.key, s.verts, indexes)
	s.transform.Initialize(host.WorkGroup())
	s.transform.SetPosition(matrix.NewVec3(0, 0, shaderGraphSplineDepth))
	s.shader = s.newShaderData()
	s.addDrawing()
	s.Hide()
}

func (s *shaderGraphSpline) newShaderData() *ui.ShaderData {
	shader := &ui.ShaderData{
		ShaderDataBase: rendering.NewShaderDataBase(),
		UVs:            matrix.Vec4{0, 0, 1, 1},
		FgColor:        s.color,
		BgColor:        matrix.ColorClear(),
		Scissor:        matrix.Vec4{-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax},
		Size2D:         matrix.Vec4{1, 1, 1, 1},
		BorderLen:      matrix.Vec2{8, 8},
	}
	return shader
}

func (s *shaderGraphSpline) addDrawing() {
	material, err := s.host.MaterialCache().Material(assets.MaterialDefinitionUITransparent)
	if err != nil {
		slog.Error("failed to load shader graph spline material",
			"material", assets.MaterialDefinitionUITransparent, "error", err)
		return
	}
	texture, err := s.host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterNearest)
	if err != nil {
		slog.Error("failed to load shader graph spline texture",
			"texture", assets.TextureSquare, "error", err)
		return
	}
	s.host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material.CreateInstance([]*rendering.Texture{texture}),
		Mesh:       s.mesh,
		ShaderData: s.shader,
		Transform:  &s.transform,
		Sort:       435,
		Layer:      rendering.RenderLayerUI,
		ViewCuller: &s.host.Cameras.UI,
	})
}

func (s *shaderGraphSpline) SetColor(color matrix.Color) {
	if matrix.Vec4Approx(matrix.Vec4(s.color), matrix.Vec4(color)) {
		return
	}
	s.color = color
	if s.shader != nil {
		s.shader.FgColor = color
	}
}

func (s *shaderGraphSpline) Show() {
	if s.shader == nil {
		return
	}
	s.shader.Activate()
}

func (s *shaderGraphSpline) Hide() {
	if s.shader != nil {
		s.shader.Deactivate()
	}
}

func (s *shaderGraphSpline) Update(start, end matrix.Vec2) {
	if s.host == nil || s.mesh == nil || s.shader == nil {
		return
	}
	controlDistance := max(matrix.Abs(end.X()-start.X())*0.55, matrix.Float(70.0))
	c1 := start.Add(matrix.NewVec2(controlDistance, 0))
	c2 := end.Subtract(matrix.NewVec2(controlDistance, 0))
	s.updateMesh(start, c1, c2, end)
	s.shader.Scissor = s.panelScissor()
	// Dynamic vertex edits change the mesh bounds while the transform stays fixed.
	// Mark it dirty so render culling recalculates against the new bounds.
	s.transform.SetDirty()
	s.host.MeshCache().UpdateMeshVertices(s.mesh.Key(), s.verts)
}

func (s *shaderGraphSpline) Destroy() {
	if s.shader != nil {
		s.shader.Destroy()
	}
	s.shader = nil
	s.mesh = nil
	s.verts = nil
	s.points = nil
}

func (s *shaderGraphSpline) updateMesh(p0, p1, p2, p3 matrix.Vec2) {
	for i := range s.points {
		t := matrix.Float(i) / matrix.Float(shaderGraphSplineSegments)
		s.points[i] = s.localToUIWorld(shaderGraphBezierPoint(p0, p1, p2, p3, t))
	}
	for i, point := range s.points {
		tangent := s.splineTangent(i)
		normal := matrix.NewVec2(-tangent.Y(), tangent.X()).Normal().Scale(shaderGraphSplineStroke * 0.5)
		t := matrix.Float(i) / matrix.Float(shaderGraphSplineSegments)
		base := i * 2
		s.verts[base+0] = shaderGraphSplineVertex(point.Subtract(normal), matrix.NewVec2(t, 1))
		s.verts[base+1] = shaderGraphSplineVertex(point.Add(normal), matrix.NewVec2(t, 0))
	}
}

func (s *shaderGraphSpline) splineTangent(index int) matrix.Vec2 {
	last := len(s.points) - 1
	if last <= 0 {
		return matrix.NewVec2(1, 0)
	}
	var tangent matrix.Vec2
	switch {
	case index <= 0:
		tangent = s.points[1].Subtract(s.points[0])
	case index >= last:
		tangent = s.points[last].Subtract(s.points[last-1])
	default:
		tangent = s.points[index+1].Subtract(s.points[index-1])
	}
	if tangent.IsZero() {
		return matrix.NewVec2(1, 0)
	}
	return tangent
}

func (s *shaderGraphSpline) hideVertices() {
	for i := range s.verts {
		s.verts[i] = rendering.Vertex{
			Position: matrix.NewVec3(-matrix.FloatMax, -matrix.FloatMax, 0),
			UV0:      matrix.NewVec2(0, 0),
			Color:    matrix.ColorClear(),
		}
	}
}

func (s *shaderGraphSpline) localToUIWorld(point matrix.Vec2) matrix.Vec2 {
	if s.host == nil || s.host.Window == nil || s.root == nil {
		return point
	}
	offset := s.root.Base().Layout().Offset()
	windowWidth := matrix.Float(s.host.Window.Width())
	windowHeight := matrix.Float(s.host.Window.Height())
	return matrix.NewVec2(
		offset.X()+point.X()-windowWidth*0.5,
		windowHeight*0.5-offset.Y()-point.Y(),
	)
}

func (s *shaderGraphSpline) panelScissor() matrix.Vec4 {
	if s.host == nil || s.host.Window == nil || s.root == nil {
		return matrix.Vec4{-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax}
	}
	layout := s.root.Base().Layout()
	offset := layout.Offset()
	size := layout.PixelSize()
	windowWidth := matrix.Float(s.host.Window.Width())
	windowHeight := matrix.Float(s.host.Window.Height())
	minX := offset.X() - windowWidth*0.5
	maxX := minX + size.X()
	maxY := windowHeight*0.5 - offset.Y()
	minY := maxY - size.Y()
	return matrix.Vec4{minX, minY, maxX, maxY}
}

func shaderGraphBezierPoint(p0, p1, p2, p3 matrix.Vec2, t matrix.Float) matrix.Vec2 {
	mt := 1 - t
	return p0.Scale(mt * mt * mt).
		Add(p1.Scale(3 * mt * mt * t)).
		Add(p2.Scale(3 * mt * t * t)).
		Add(p3.Scale(t * t * t))
}

func shaderGraphSplineVertex(position matrix.Vec2, uv matrix.Vec2) rendering.Vertex {
	return rendering.Vertex{
		Position: matrix.NewVec3(position.X(), position.Y(), 0),
		UV0:      uv,
		Color:    matrix.ColorWhite(),
	}
}
