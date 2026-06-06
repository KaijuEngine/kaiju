/******************************************************************************/
/* render_graph_output_compiler.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const (
	renderGraphOutputFloat = "float"
	renderGraphOutputVec2  = "vec2"
	renderGraphOutputVec3  = "vec3"
	renderGraphOutputVec4  = "vec4"
	renderGraphOutputColor = "color"
	renderGraphOutputTex2D = "texture2d"
)

type renderGraphCompiledOutput struct {
	VertexSource   string
	FragmentSource string
	SamplerLabels  []string
	Textures       []rendering.MaterialTextureData
}

type renderGraphOutputSurface struct {
	BaseColor           string
	Metallic            string
	Roughness           string
	Normal              string
	Occlusion           string
	EmissionColor       string
	EmissionStrength    string
	Alpha               string
	Specular            string
	UseAlphaInput       bool
	UseTextureMetallic  bool
	UseTextureRoughness bool
	UseTextureNormal    bool
	UseTextureOcclusion bool
	UseTextureEmission  bool
}

type renderGraphOutputExpression struct {
	Type       string
	Value      string
	ColorSpace string
}

type renderGraphOutputCompiler struct {
	document     RenderGraphDocument
	vertexStage  bool
	nodes        map[string]RenderGraphNode
	specs        map[string]renderGraphNodeSpec
	incoming     map[RenderGraphPortRef]RenderGraphPortRef
	cache        map[RenderGraphPortRef]renderGraphOutputExpression
	visiting     map[RenderGraphPortRef]bool
	textures     []rendering.MaterialTextureData
	textureSlots map[string]int
}

func compileRenderGraphDocumentOutput(document RenderGraphDocument) (renderGraphCompiledOutput, error) {
	compiler, err := newRenderGraphOutputCompiler(document)
	if err != nil {
		return renderGraphCompiledOutput{}, err
	}
	surface, err := compiler.compileMaterialSurface()
	if err != nil {
		return renderGraphCompiledOutput{}, err
	}
	vertexCompiler, err := newRenderGraphOutputCompiler(document)
	if err != nil {
		return renderGraphCompiledOutput{}, err
	}
	vertexCompiler.vertexStage = true
	displacement, err := vertexCompiler.compileMaterialDisplacement()
	if err != nil {
		return renderGraphCompiledOutput{}, err
	}
	return renderGraphCompiledOutput{
		VertexSource:   renderGraphPBRVertexSource(displacement),
		FragmentSource: renderGraphPBRFragmentSource(surface, len(compiler.textures)),
		SamplerLabels:  renderGraphSamplerLabels(compiler.textures),
		Textures:       append([]rendering.MaterialTextureData(nil), compiler.textures...),
	}, nil
}

func newRenderGraphOutputCompiler(document RenderGraphDocument) (*renderGraphOutputCompiler, error) {
	if err := validateRenderGraphDocument(document); err != nil {
		return nil, err
	}
	compiler := &renderGraphOutputCompiler{
		document:     document,
		nodes:        make(map[string]RenderGraphNode, len(document.Nodes)),
		specs:        make(map[string]renderGraphNodeSpec, len(document.Nodes)),
		incoming:     make(map[RenderGraphPortRef]RenderGraphPortRef, len(document.Connections)),
		cache:        map[RenderGraphPortRef]renderGraphOutputExpression{},
		visiting:     map[RenderGraphPortRef]bool{},
		textures:     renderGraphDefaultTextureSlots(),
		textureSlots: map[string]int{},
	}
	for i := range document.Nodes {
		node := document.Nodes[i]
		spec, _ := renderGraphNodeCatalogSpec(node.Type)
		compiler.nodes[node.ID] = node
		compiler.specs[node.ID] = spec
	}
	for i := range document.Connections {
		connection := document.Connections[i]
		outputNode, outputOK := compiler.nodes[connection.Output.Node]
		inputNode, inputOK := compiler.nodes[connection.Input.Node]
		if !outputOK || !inputOK {
			return nil, fmt.Errorf("render graph connection %d references missing nodes", i)
		}
		outputSpec := compiler.specs[outputNode.ID]
		inputSpec := compiler.specs[inputNode.ID]
		if connection.Output.Port < 0 || connection.Output.Port >= len(outputSpec.Outputs) {
			return nil, fmt.Errorf("render graph connection %d references invalid output port %d on node %q",
				i, connection.Output.Port, outputNode.ID)
		}
		if connection.Input.Port < 0 || connection.Input.Port >= len(inputSpec.Inputs) {
			return nil, fmt.Errorf("render graph connection %d references invalid input port %d on node %q",
				i, connection.Input.Port, inputNode.ID)
		}
		outputPort := outputSpec.Outputs[connection.Output.Port]
		inputPort := inputSpec.Inputs[connection.Input.Port]
		if renderGraphPortTypeKey(outputPort.Type) != renderGraphPortTypeKey(inputPort.Type) {
			return nil, fmt.Errorf("render graph connection %d links %q output to %q input",
				i, outputPort.Type, inputPort.Type)
		}
		if _, exists := compiler.incoming[connection.Input]; exists {
			return nil, fmt.Errorf("render graph input %q port %d has multiple connections",
				connection.Input.Node, connection.Input.Port)
		}
		compiler.incoming[connection.Input] = connection.Output
	}
	return compiler, nil
}

func (c *renderGraphOutputCompiler) compileMaterialSurface() (renderGraphOutputSurface, error) {
	output, err := c.materialOutputNode()
	if err != nil {
		return renderGraphOutputSurface{}, err
	}
	surfaceInput := RenderGraphPortRef{Node: output.ID, Port: 0}
	surfaceRef, ok := c.incoming[surfaceInput]
	if !ok {
		return renderGraphOutputSurface{}, fmt.Errorf("material output surface is disconnected")
	}
	surfaceNode, ok := c.nodes[surfaceRef.Node]
	if !ok {
		return renderGraphOutputSurface{}, fmt.Errorf("material output references missing surface node %q", surfaceRef.Node)
	}
	if surfaceNode.Type != "principled-bsdf" || surfaceRef.Port != 0 {
		return renderGraphOutputSurface{}, fmt.Errorf("material output surface must come from principled-bsdf")
	}
	return c.compilePrincipledSurface(surfaceNode)
}

func (c *renderGraphOutputCompiler) compileMaterialDisplacement() (string, error) {
	output, err := c.materialOutputNode()
	if err != nil {
		return "", err
	}
	displacementRef, ok := c.incoming[RenderGraphPortRef{Node: output.ID, Port: 1}]
	if !ok {
		return "0.0", nil
	}
	expr, err := c.emitExpression(displacementRef, renderGraphOutputFloat)
	if err != nil {
		return "", err
	}
	if len(c.textures) > len(renderGraphDefaultTextureSlots()) {
		return "", fmt.Errorf("render graph vertex displacement does not support texture sampling yet")
	}
	return expr.Value, nil
}

func (c *renderGraphOutputCompiler) materialOutputNode() (RenderGraphNode, error) {
	var output RenderGraphNode
	count := 0
	for i := range c.document.Nodes {
		if c.document.Nodes[i].Type == "material-output" {
			output = c.document.Nodes[i]
			count++
		}
	}
	if count == 0 {
		return output, fmt.Errorf("render graph is missing a material output node")
	}
	if count > 1 {
		return output, fmt.Errorf("render graph has multiple material output nodes")
	}
	return output, nil
}

func (c *renderGraphOutputCompiler) compilePrincipledSurface(node RenderGraphNode) (renderGraphOutputSurface, error) {
	surface := renderGraphOutputSurface{
		BaseColor:           "fragColor",
		Metallic:            "fragMetallic",
		Roughness:           "fragRoughness",
		Normal:              "pbrNormal(geometricNormal)",
		Occlusion:           "1.0",
		EmissionColor:       "vec3(0.0)",
		EmissionStrength:    "fragEmissive",
		Alpha:               "baseSample.a * graphBaseColor.a",
		Specular:            "1.0",
		UseTextureMetallic:  true,
		UseTextureRoughness: true,
		UseTextureNormal:    true,
		UseTextureOcclusion: true,
		UseTextureEmission:  true,
	}
	if colorRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 0}]; ok {
		expr, err := c.emitExpression(colorRef, renderGraphOutputColor)
		if err != nil {
			return surface, err
		}
		surface.BaseColor = "(" + expr.Value + " * fragColor)"
	}
	if roughnessRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 1}]; ok {
		expr, err := c.emitExpression(roughnessRef, renderGraphOutputFloat)
		if err != nil {
			return surface, err
		}
		surface.Roughness = expr.Value
		surface.UseTextureRoughness = false
	}
	if normalRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 2}]; ok {
		expr, err := c.emitExpression(normalRef, renderGraphOutputVec3)
		if err != nil {
			return surface, err
		}
		surface.Normal = "safeNormalize(" + expr.Value + ", geometricNormal)"
		surface.UseTextureNormal = false
	}
	if metallicRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 3}]; ok {
		expr, err := c.emitExpression(metallicRef, renderGraphOutputFloat)
		if err != nil {
			return surface, err
		}
		surface.Metallic = expr.Value
		surface.UseTextureMetallic = false
	}
	if occlusionRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 4}]; ok {
		expr, err := c.emitExpression(occlusionRef, renderGraphOutputFloat)
		if err != nil {
			return surface, err
		}
		surface.Occlusion = expr.Value
		surface.UseTextureOcclusion = false
	}
	if emissionColorRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 5}]; ok {
		expr, err := c.emitExpression(emissionColorRef, renderGraphOutputColor)
		if err != nil {
			return surface, err
		}
		surface.EmissionColor = "(" + expr.Value + ").rgb"
		surface.UseTextureEmission = false
	}
	if emissionStrengthRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 6}]; ok {
		expr, err := c.emitExpression(emissionStrengthRef, renderGraphOutputFloat)
		if err != nil {
			return surface, err
		}
		surface.EmissionStrength = expr.Value
	}
	if alphaRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 7}]; ok {
		expr, err := c.emitExpression(alphaRef, renderGraphOutputFloat)
		if err != nil {
			return surface, err
		}
		surface.Alpha = expr.Value
		surface.UseAlphaInput = true
	}
	if specularRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 8}]; ok {
		expr, err := c.emitExpression(specularRef, renderGraphOutputFloat)
		if err != nil {
			return surface, err
		}
		surface.Specular = expr.Value
	}
	return surface, nil
}

func (c *renderGraphOutputCompiler) emitExpression(ref RenderGraphPortRef, wantType string) (renderGraphOutputExpression, error) {
	expr, err := c.emitOutput(ref)
	if err != nil {
		return expr, err
	}
	if expr.Type != wantType {
		return expr, fmt.Errorf("render graph node %q output %d is %q, want %q",
			ref.Node, ref.Port, expr.Type, wantType)
	}
	return expr, nil
}

func (c *renderGraphOutputCompiler) emitOutput(ref RenderGraphPortRef) (renderGraphOutputExpression, error) {
	if cached, ok := c.cache[ref]; ok {
		return cached, nil
	}
	if c.visiting[ref] {
		return renderGraphOutputExpression{}, fmt.Errorf("render graph contains a cycle at node %q", ref.Node)
	}
	node, ok := c.nodes[ref.Node]
	if !ok {
		return renderGraphOutputExpression{}, fmt.Errorf("render graph references missing node %q", ref.Node)
	}
	spec := c.specs[ref.Node]
	if ref.Port < 0 || ref.Port >= len(spec.Outputs) {
		return renderGraphOutputExpression{}, fmt.Errorf("render graph references invalid output port %d on node %q", ref.Port, ref.Node)
	}
	c.visiting[ref] = true
	defer delete(c.visiting, ref)

	expr, err := c.emitNodeOutput(node, ref.Port)
	if err != nil {
		return expr, err
	}
	wantType := renderGraphPortTypeKey(spec.Outputs[ref.Port].Type)
	if expr.Type != wantType {
		return expr, fmt.Errorf("render graph node %q emitted %q for %q output", node.ID, expr.Type, wantType)
	}
	c.cache[ref] = expr
	return expr, nil
}

func (c *renderGraphOutputCompiler) emitNodeOutput(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	switch node.Type {
	case "value":
		value, err := c.floatField(node, "value")
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: value}, err
	case "color":
		value := c.fieldValue(node, "color").Color
		return renderGraphOutputExpression{
			Type:  renderGraphOutputColor,
			Value: glslColor(value),
		}, nil
	case "vector":
		value, err := c.vectorField(node, "vector")
		return renderGraphOutputExpression{Type: renderGraphOutputVec3, Value: value}, err
	case "vector2":
		value, err := c.vector2Field(node, "vector")
		return renderGraphOutputExpression{Type: renderGraphOutputVec2, Value: value}, err
	case "vector4":
		value, err := c.vector4Field(node, "vector")
		if port == 1 {
			return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: value}, err
		}
		if port != 0 {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph vector4 node %q has invalid output %d", node.ID, port)
		}
		return renderGraphOutputExpression{Type: renderGraphOutputVec4, Value: value}, err
	case "combine-vec2":
		return c.emitCombineVector(node, port, "vec2", renderGraphOutputVec2, 2)
	case "combine-vec3":
		return c.emitCombineVector(node, port, "vec3", renderGraphOutputVec3, 3)
	case "combine-vec4":
		return c.emitCombineVector(node, port, "vec4", renderGraphOutputVec4, 4)
	case "split-vec2":
		return c.emitSplitVector(node, port, renderGraphOutputVec2, 2)
	case "split-vec3":
		return c.emitSplitVector(node, port, renderGraphOutputVec3, 3)
	case "split-vec4":
		return c.emitSplitVector(node, port, renderGraphOutputVec4, 4)
	case "swizzle-vec2":
		return c.emitSwizzleVector(node, port, renderGraphOutputVec2, "vec2", 2)
	case "swizzle-vec3":
		return c.emitSwizzleVector(node, port, renderGraphOutputVec3, "vec3", 3)
	case "swizzle-vec4":
		return c.emitSwizzleVector(node, port, renderGraphOutputVec4, "vec4", 4)
	case "texture-2d":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return c.emitTexture2D(node)
	case "sample-texture-2d":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return c.emitSampleTexture2D(node, port)
	case "uv":
		return renderGraphOutputExpression{Type: renderGraphOutputVec2, Value: c.uvExpression()}, nil
	case "uv-transform":
		return c.emitUVTransform(node)
	case "split-rgba":
		return c.emitSplitRGBA(node, port)
	case "channel-mask":
		return c.emitChannelMask(node)
	case "texel-size":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return c.emitTexelSize(node, port)
	case "normal-map":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return c.emitNormalMap(node, port)
	case "normal-strength":
		return c.emitNormalStrength(node)
	case "blend-normals":
		return c.emitBlendNormals(node)
	case "orm-mra-unpack":
		return c.emitPackedPBRMap(node, port)
	case "height-bump":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return c.emitHeightBump(node)
	case "parallax":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return c.emitParallax(node, port)
	case "triplanar":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return c.emitTriplanar(node, port)
	case "detail-texture":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return c.emitDetailTexture(node)
	case "time":
		return emitTime(node, port)
	case "world-position":
		return emitVec3Context(node, port, c.worldPositionExpression())
	case "normal-vector":
		return emitVec3Context(node, port, c.geometricNormalExpression())
	case "tangent-vector":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return emitVec3Context(node, port, "safeNormalize(cotangentFrame("+
			graphGeometricNormalExpression()+", fragPos, fragTexCoords)[0], vec3(1.0, 0.0, 0.0))")
	case "bitangent-vector":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return emitVec3Context(node, port, "safeNormalize(cotangentFrame("+
			graphGeometricNormalExpression()+", fragPos, fragTexCoords)[1], vec3(0.0, 0.0, 1.0))")
	case "view-direction":
		return emitVec3Context(node, port, c.viewDirectionExpression())
	case "camera-position":
		return emitVec3Context(node, port, "cameraPosition.xyz")
	case "screen-position":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return emitScreenPosition(node, port)
	case "vertex-color":
		return emitVertexColor(node, port)
	case "noise":
		return c.emitNoise(node, port)
	case "voronoi":
		return c.emitVoronoi(node, port)
	case "checker":
		return c.emitChecker(node, port)
	case "gradient":
		return c.emitGradient(node, port)
	case "remap":
		return c.emitRemap(node)
	case "posterize":
		return c.emitPosterize(node)
	case "posterize-color":
		return c.emitPosterizeColor(node)
	case "fresnel":
		return c.emitFresnel(node)
	case "rim-light":
		return c.emitRimLight(node, port)
	case "fwidth", "ddx", "ddy":
		if c.vertexStage {
			return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q cannot be used for vertex displacement", node.Type)
		}
		return c.emitDerivative(node)
	case "add", "subtract", "multiply", "divide", "minimum", "maximum", "power":
		return c.emitFloatBinary(node)
	case "add-vec2", "subtract-vec2", "multiply-vec2", "divide-vec2":
		return c.emitVectorArithmetic(node, port, renderGraphOutputVec2)
	case "add-vec3", "subtract-vec3", "multiply-vec3", "divide-vec3":
		return c.emitVectorArithmetic(node, port, renderGraphOutputVec3)
	case "add-vec4", "subtract-vec4", "multiply-vec4", "divide-vec4":
		return c.emitVectorArithmetic(node, port, renderGraphOutputVec4)
	case "absolute", "one-minus", "floor", "ceiling", "fraction", "sine", "cosine", "tangent", "square-root":
		return c.emitFloatUnary(node)
	case "clamp", "lerp":
		return c.emitFloatTernary(node)
	case "step", "smoothstep":
		return c.emitFloatStep(node)
	case "dot-product":
		return c.emitVectorBinaryFloat(node, "dot")
	case "cross-product":
		return c.emitVectorBinaryVec3(node, "cross")
	case "normalize":
		value, err := c.inputExpression(node, 0, renderGraphOutputVec3)
		return renderGraphOutputExpression{Type: renderGraphOutputVec3, Value: "normalize(" + value + ")"}, err
	case "length":
		value, err := c.inputExpression(node, 0, renderGraphOutputVec3)
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "length(" + value + ")"}, err
	case "mix-color":
		return c.emitMixColor(node)
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph node type %q is not supported for material generation", node.Type)
	}
}

func (c *renderGraphOutputCompiler) emitFloatBinary(node RenderGraphNode) (renderGraphOutputExpression, error) {
	a, err := c.inputExpression(node, 0, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	b, err := c.inputExpression(node, 1, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	value := ""
	switch node.Type {
	case "add":
		value = "(" + a + " + " + b + ")"
	case "subtract":
		value = "(" + a + " - " + b + ")"
	case "multiply":
		value = "(" + a + " * " + b + ")"
	case "divide":
		value = "(" + a + " / " + b + ")"
	case "minimum":
		value = "min(" + a + ", " + b + ")"
	case "maximum":
		value = "max(" + a + ", " + b + ")"
	case "power":
		value = "pow(" + a + ", " + b + ")"
	}
	return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: value}, nil
}

func (c *renderGraphOutputCompiler) emitVectorArithmetic(node RenderGraphNode, port int, vectorType string) (renderGraphOutputExpression, error) {
	if port != 0 {
		if vectorType == renderGraphOutputVec4 && port == 1 {
			value, err := c.emitVectorArithmeticValue(node, vectorType)
			return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: value}, err
		}
		return renderGraphOutputExpression{}, fmt.Errorf("render graph vector arithmetic node %q has invalid output %d", node.ID, port)
	}
	value, err := c.emitVectorArithmeticValue(node, vectorType)
	return renderGraphOutputExpression{Type: vectorType, Value: value}, err
}

func (c *renderGraphOutputCompiler) emitVectorArithmeticValue(node RenderGraphNode, vectorType string) (string, error) {
	a, err := c.inputExpression(node, 0, vectorType)
	if err != nil {
		return "", err
	}
	b, err := c.inputExpression(node, 1, vectorType)
	if err != nil {
		return "", err
	}
	operator := ""
	switch {
	case strings.HasPrefix(node.Type, "add-"):
		operator = "+"
	case strings.HasPrefix(node.Type, "subtract-"):
		operator = "-"
	case strings.HasPrefix(node.Type, "multiply-"):
		operator = "*"
	case strings.HasPrefix(node.Type, "divide-"):
		operator = "/"
	default:
		return "", fmt.Errorf("render graph vector arithmetic node %q is not supported", node.Type)
	}
	return "(" + a + " " + operator + " " + b + ")", nil
}

func (c *renderGraphOutputCompiler) emitFloatUnary(node RenderGraphNode) (renderGraphOutputExpression, error) {
	value, err := c.inputExpression(node, 0, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	switch node.Type {
	case "absolute":
		value = "abs(" + value + ")"
	case "one-minus":
		value = "(1.0 - " + value + ")"
	case "floor":
		value = "floor(" + value + ")"
	case "ceiling":
		value = "ceil(" + value + ")"
	case "fraction":
		value = "fract(" + value + ")"
	case "sine":
		value = "sin(" + value + ")"
	case "cosine":
		value = "cos(" + value + ")"
	case "tangent":
		value = "tan(" + value + ")"
	case "square-root":
		value = "sqrt(" + value + ")"
	}
	return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: value}, nil
}

func (c *renderGraphOutputCompiler) emitFloatTernary(node RenderGraphNode) (renderGraphOutputExpression, error) {
	a, err := c.inputExpression(node, 0, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	b, err := c.inputExpression(node, 1, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	d, err := c.inputExpression(node, 2, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	value := ""
	switch node.Type {
	case "clamp":
		value = "clamp(" + a + ", " + b + ", " + d + ")"
	case "lerp":
		value = "mix(" + a + ", " + b + ", " + d + ")"
	}
	return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: value}, nil
}

func (c *renderGraphOutputCompiler) emitFloatStep(node RenderGraphNode) (renderGraphOutputExpression, error) {
	a, err := c.inputExpression(node, 0, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	b, err := c.inputExpression(node, 1, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	if node.Type == "step" {
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "step(" + a + ", " + b + ")"}, nil
	}
	d, err := c.inputExpression(node, 2, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "smoothstep(" + a + ", " + b + ", " + d + ")"}, nil
}

func (c *renderGraphOutputCompiler) emitVectorBinaryFloat(node RenderGraphNode, fn string) (renderGraphOutputExpression, error) {
	a, err := c.inputExpression(node, 0, renderGraphOutputVec3)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	b, err := c.inputExpression(node, 1, renderGraphOutputVec3)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: fn + "(" + a + ", " + b + ")"}, nil
}

func (c *renderGraphOutputCompiler) emitVectorBinaryVec3(node RenderGraphNode, fn string) (renderGraphOutputExpression, error) {
	a, err := c.inputExpression(node, 0, renderGraphOutputVec3)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	b, err := c.inputExpression(node, 1, renderGraphOutputVec3)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{Type: renderGraphOutputVec3, Value: fn + "(" + a + ", " + b + ")"}, nil
}

func (c *renderGraphOutputCompiler) emitMixColor(node RenderGraphNode) (renderGraphOutputExpression, error) {
	factor, err := c.inputExpression(node, 0, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	a, err := c.inputExpression(node, 1, renderGraphOutputColor)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	b, err := c.inputExpression(node, 2, renderGraphOutputColor)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	mode := c.fieldValue(node, "mode").Option
	if mode == "" {
		mode = "mix"
	}
	if c.fieldValue(node, "clamp").Bool {
		factor = "clamp(" + factor + ", 0.0, 1.0)"
	}
	value := ""
	switch mode {
	case "mix":
		value = "mix(" + a + ", " + b + ", " + factor + ")"
	case "add":
		value = "mix(" + a + ", (" + a + " + " + b + "), " + factor + ")"
	case "multiply":
		value = "mix(" + a + ", (" + a + " * " + b + "), " + factor + ")"
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("mix-color node %q has unsupported mode %q", node.ID, mode)
	}
	if c.fieldValue(node, "clamp").Bool {
		value = "clamp(" + value + ", vec4(0.0), vec4(1.0))"
	}
	return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: value}, nil
}

func (c *renderGraphOutputCompiler) emitTexture2D(node RenderGraphNode) (renderGraphOutputExpression, error) {
	if index, ok := c.textureSlots[node.ID]; ok {
		return renderGraphOutputExpression{
			Type:       renderGraphOutputTex2D,
			Value:      fmt.Sprintf("textures[%d]", index),
			ColorSpace: c.textureColorSpace(node),
		}, nil
	}
	texture := strings.TrimSpace(c.fieldValue(node, "texture").Text)
	if texture == "" {
		texture = assets.TextureSquare
	}
	texture = filepath.ToSlash(texture)
	texture = strings.ReplaceAll(texture, "\\", "/")
	label := strings.TrimSpace(c.fieldValue(node, "label").Text)
	if label == "" {
		label = "Texture"
	}
	filter := c.fieldValue(node, "filter").Option
	switch filter {
	case "Nearest", "Linear":
	default:
		filter = "Linear"
	}
	index := len(c.textures)
	c.textures = append(c.textures, rendering.MaterialTextureData{
		Label:   c.uniqueTextureLabel(label),
		Texture: texture,
		Filter:  filter,
	})
	c.textureSlots[node.ID] = index
	return renderGraphOutputExpression{
		Type:       renderGraphOutputTex2D,
		Value:      fmt.Sprintf("textures[%d]", index),
		ColorSpace: c.textureColorSpace(node),
	}, nil
}

func (c *renderGraphOutputCompiler) emitSampleTexture2D(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	texRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 0}]
	if !ok {
		return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q texture input is disconnected", node.ID)
	}
	texture, err := c.emitExpression(texRef, renderGraphOutputTex2D)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	uv, err := c.optionalInputExpression(node, 1, renderGraphOutputVec2, "fragTexCoords")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	sample := "texture(" + texture.Value + ", " + uv + ")"
	if texture.ColorSpace == "srgb" {
		sample = "graphSrgbToLinear(" + sample + ")"
	}
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: sample}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputVec3, Value: sample + ".rgb"}, nil
	case 2:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: sample + ".r"}, nil
	case 3:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: sample + ".g"}, nil
	case 4:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: sample + ".b"}, nil
	case 5:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: sample + ".a"}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph sample texture node %q has invalid output %d", node.ID, port)
	}
}

func (c *renderGraphOutputCompiler) emitUVTransform(node RenderGraphNode) (renderGraphOutputExpression, error) {
	uv, err := c.optionalInputExpression(node, 0, renderGraphOutputVec2, c.uvExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	tiling, err := c.vector2Field(node, "tiling")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	offset, err := c.vector2Field(node, "offset")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{Type: renderGraphOutputVec2, Value: "((" + uv + ") * " + tiling + " + " + offset + ")"}, nil
}

func (c *renderGraphOutputCompiler) emitSplitRGBA(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	color, err := c.inputExpression(node, 0, renderGraphOutputColor)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	components := []string{".r", ".g", ".b", ".a"}
	if port < 0 || port >= len(components) {
		return renderGraphOutputExpression{}, fmt.Errorf("render graph split rgba node %q has invalid output %d", node.ID, port)
	}
	return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + color + ")" + components[port]}, nil
}

func (c *renderGraphOutputCompiler) emitChannelMask(node RenderGraphNode) (renderGraphOutputExpression, error) {
	color, err := c.inputExpression(node, 0, renderGraphOutputColor)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	switch c.fieldValue(node, "channel").Option {
	case "g":
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + color + ").g"}, nil
	case "b":
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + color + ").b"}, nil
	case "a":
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + color + ").a"}, nil
	case "luma":
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "dot((" + color + ").rgb, vec3(0.2126, 0.7152, 0.0722))"}, nil
	default:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + color + ").r"}, nil
	}
}

func (c *renderGraphOutputCompiler) emitTexelSize(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	texRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 0}]
	if !ok {
		return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q texture input is disconnected", node.ID)
	}
	texture, err := c.emitExpression(texRef, renderGraphOutputTex2D)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	size := "(1.0 / vec2(textureSize(" + texture.Value + ", 0)))"
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputVec2, Value: size}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: size + ".x"}, nil
	case 2:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: size + ".y"}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph texel size node %q has invalid output %d", node.ID, port)
	}
}

func (c *renderGraphOutputCompiler) emitNormalMap(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	rgb, err := c.inputExpression(node, 0, renderGraphOutputVec3)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	uv, err := c.optionalInputExpression(node, 1, renderGraphOutputVec2, "fragTexCoords")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	strength, err := c.optionalInputOrFloatField(node, 2, "strength", "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	flipY := "1.0"
	if c.fieldValue(node, "y").Option == "directx" {
		flipY = "-1.0"
	}
	tangentNormal := "graphTangentNormalFromMap(" + rgb + ", " + strength + ", " + flipY + ")"
	switch port {
	case 0:
		return renderGraphOutputExpression{
			Type: renderGraphOutputVec3,
			Value: "graphWorldNormalFromTangent(" + tangentNormal + ", " + uv + ", " +
				graphGeometricNormalExpression() + ")",
		}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputVec3, Value: tangentNormal}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph normal map node %q has invalid output %d", node.ID, port)
	}
}

func (c *renderGraphOutputCompiler) emitNormalStrength(node RenderGraphNode) (renderGraphOutputExpression, error) {
	normal, err := c.optionalInputExpression(node, 0, renderGraphOutputVec3, c.geometricNormalExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	strength, err := c.optionalInputOrFloatField(node, 1, "strength", "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{
		Type: renderGraphOutputVec3,
		Value: "graphApplyNormalStrength(" + normal + ", " + strength + ", " +
			c.geometricNormalExpression() + ")",
	}, nil
}

func (c *renderGraphOutputCompiler) emitBlendNormals(node RenderGraphNode) (renderGraphOutputExpression, error) {
	base, err := c.optionalInputExpression(node, 0, renderGraphOutputVec3, c.geometricNormalExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	detail, err := c.optionalInputExpression(node, 1, renderGraphOutputVec3, c.geometricNormalExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	strength, err := c.optionalInputOrFloatField(node, 2, "strength", "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{
		Type: renderGraphOutputVec3,
		Value: "graphBlendNormals(" + base + ", " + detail + ", " + strength + ", " +
			c.geometricNormalExpression() + ")",
	}, nil
}

func (c *renderGraphOutputCompiler) emitPackedPBRMap(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	packed, err := c.inputExpression(node, 0, renderGraphOutputColor)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	layout := c.fieldValue(node, "layout").Option
	components := []string{".r", ".g", ".b"}
	if layout == "mra" {
		components = []string{".b", ".g", ".r"}
	}
	if port < 0 || port >= len(components) {
		return renderGraphOutputExpression{}, fmt.Errorf("render graph packed pbr map node %q has invalid output %d", node.ID, port)
	}
	return renderGraphOutputExpression{
		Type:  renderGraphOutputFloat,
		Value: "clamp((" + packed + ")" + components[port] + ", 0.0, 1.0)",
	}, nil
}

func (c *renderGraphOutputCompiler) emitHeightBump(node RenderGraphNode) (renderGraphOutputExpression, error) {
	height, err := c.optionalInputExpression(node, 0, renderGraphOutputFloat, "0.5")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	strength, err := c.optionalInputOrFloatField(node, 1, "strength", "0.05")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{
		Type: renderGraphOutputVec3,
		Value: "graphBumpNormal(" + height + ", " + strength + ", " +
			graphGeometricNormalExpression() + ")",
	}, nil
}

func (c *renderGraphOutputCompiler) emitParallax(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	uv, err := c.optionalInputExpression(node, 0, renderGraphOutputVec2, "fragTexCoords")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	height, err := c.optionalInputExpression(node, 1, renderGraphOutputFloat, "0.5")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	scale, err := c.optionalInputOrFloatField(node, 2, "scale", "0.05")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	parallaxUV := "graphParallaxUV(" + uv + ", " + height + ", " + scale + ", " +
		graphGeometricNormalExpression() + ")"
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputVec2, Value: parallaxUV}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputVec2, Value: "(" + parallaxUV + " - " + uv + ")"}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph parallax node %q has invalid output %d", node.ID, port)
	}
}

func (c *renderGraphOutputCompiler) emitTriplanar(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	texRef, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: 0}]
	if !ok {
		return renderGraphOutputExpression{}, fmt.Errorf("render graph node %q texture input is disconnected", node.ID)
	}
	texture, err := c.emitExpression(texRef, renderGraphOutputTex2D)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	position, err := c.optionalInputExpression(node, 1, renderGraphOutputVec3, "fragPos")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	normal, err := c.optionalInputExpression(node, 2, renderGraphOutputVec3, graphGeometricNormalExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	scale, err := c.optionalInputOrFloatField(node, 3, "scale", "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	blend, err := c.optionalInputOrFloatField(node, 4, "blend", "4.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	sample := "graphTriplanarSample(" + texture.Value + ", " + position + ", " + normal + ", " + scale + ", " + blend + ")"
	if texture.ColorSpace == "srgb" {
		sample = "graphSrgbToLinear(" + sample + ")"
	}
	return emitColorLikeOutput(node, port, sample, "triplanar")
}

func (c *renderGraphOutputCompiler) emitDetailTexture(node RenderGraphNode) (renderGraphOutputExpression, error) {
	base, err := c.optionalInputExpression(node, 0, renderGraphOutputColor, "vec4(1.0)")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	detail, err := c.optionalInputExpression(node, 1, renderGraphOutputColor, "vec4(1.0)")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	mask, err := c.optionalInputExpression(node, 2, renderGraphOutputFloat, "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	strength, err := c.optionalInputOrFloatField(node, 3, "strength", "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	factor := "clamp((" + mask + ") * (" + strength + "), 0.0, 1.0)"
	value := ""
	switch c.fieldValue(node, "mode").Option {
	case "add":
		value = "mix(" + base + ", vec4((" + base + ").rgb + ((" + detail + ").rgb - vec3(0.5)), (" + base + ").a), " + factor + ")"
	case "overlay":
		value = "mix(" + base + ", graphOverlayColor(" + base + ", " + detail + "), " + factor + ")"
	default:
		value = "mix(" + base + ", vec4((" + base + ").rgb * (" + detail + ").rgb, (" + base + ").a), " + factor + ")"
	}
	if c.fieldValue(node, "clamp").Bool {
		value = "clamp(" + value + ", vec4(0.0), vec4(1.0))"
	}
	return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: value}, nil
}

func (c *renderGraphOutputCompiler) emitNoise(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	uv, err := c.optionalInputExpression(node, 0, renderGraphOutputVec2, c.uvExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	scale, err := c.optionalInputOrFloatField(node, 1, "scale", "8.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	detail, err := c.optionalInputOrFloatField(node, 2, "detail", "4.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	roughness, err := c.optionalInputOrFloatField(node, 3, "roughness", "0.5")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	value := "graphFBM2D((" + uv + ") * " + scale + ", " + detail + ", " + roughness + ")"
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: value}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: "vec4(vec3(" + value + "), 1.0)"}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph noise node %q has invalid output %d", node.ID, port)
	}
}

func (c *renderGraphOutputCompiler) emitVoronoi(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	uv, err := c.optionalInputExpression(node, 0, renderGraphOutputVec2, c.uvExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	scale, err := c.optionalInputOrFloatField(node, 1, "scale", "8.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	jitter, err := c.optionalInputOrFloatField(node, 2, "jitter", "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	value := "graphVoronoi2D((" + uv + ") * " + scale + ", " + jitter + ")"
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + value + ").x"}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + value + ").y"}, nil
	case 2:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + value + ").z"}, nil
	case 3:
		return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: "vec4(vec3((" + value + ").y), 1.0)"}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph voronoi node %q has invalid output %d", node.ID, port)
	}
}

func (c *renderGraphOutputCompiler) emitChecker(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	uv, err := c.optionalInputExpression(node, 0, renderGraphOutputVec2, c.uvExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	scale, err := c.optionalInputOrFloatField(node, 1, "scale", "8.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	mask := "graphCheckerMask(" + uv + ", " + scale + ")"
	switch port {
	case 0:
		a := c.colorFieldExpression(node, "color-a")
		b := c.colorFieldExpression(node, "color-b")
		return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: "mix(" + a + ", " + b + ", " + mask + ")"}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: mask}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph checker node %q has invalid output %d", node.ID, port)
	}
}

func (c *renderGraphOutputCompiler) emitGradient(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	uv, err := c.optionalInputExpression(node, 0, renderGraphOutputVec2, c.uvExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	angle, err := c.optionalInputOrFloatField(node, 1, "angle", "0.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	radial := "0.0"
	if c.fieldValue(node, "mode").Option == "radial" {
		radial = "1.0"
	}
	factor := "graphGradientFactor(" + uv + ", " + angle + ", " + radial + ")"
	switch port {
	case 0:
		a := c.colorFieldExpression(node, "color-a")
		b := c.colorFieldExpression(node, "color-b")
		return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: "mix(" + a + ", " + b + ", " + factor + ")"}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: factor}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph gradient node %q has invalid output %d", node.ID, port)
	}
}

func (c *renderGraphOutputCompiler) emitRemap(node RenderGraphNode) (renderGraphOutputExpression, error) {
	value, err := c.optionalInputExpression(node, 0, renderGraphOutputFloat, "0.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	inMin, err := c.optionalInputOrFloatField(node, 1, "in-min", "0.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	inMax, err := c.optionalInputOrFloatField(node, 2, "in-max", "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	outMin, err := c.optionalInputOrFloatField(node, 3, "out-min", "0.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	outMax, err := c.optionalInputOrFloatField(node, 4, "out-max", "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	result := "graphRemap(" + value + ", " + inMin + ", " + inMax + ", " + outMin + ", " + outMax + ")"
	if c.fieldValue(node, "clamp").Bool {
		result = "clamp(" + result + ", min(" + outMin + ", " + outMax + "), max(" + outMin + ", " + outMax + "))"
	}
	return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: result}, nil
}

func (c *renderGraphOutputCompiler) emitPosterize(node RenderGraphNode) (renderGraphOutputExpression, error) {
	value, err := c.optionalInputExpression(node, 0, renderGraphOutputFloat, "0.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	steps, err := c.optionalInputOrFloatField(node, 1, "steps", "4.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{
		Type:  renderGraphOutputFloat,
		Value: "graphPosterize(" + value + ", " + steps + ")",
	}, nil
}

func (c *renderGraphOutputCompiler) emitPosterizeColor(node RenderGraphNode) (renderGraphOutputExpression, error) {
	color, err := c.optionalInputExpression(node, 0, renderGraphOutputColor, "vec4(0.0, 0.0, 0.0, 1.0)")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	steps, err := c.optionalInputOrFloatField(node, 1, "steps", "4.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{
		Type:  renderGraphOutputColor,
		Value: "graphPosterizeColor(" + color + ", " + steps + ")",
	}, nil
}

func (c *renderGraphOutputCompiler) emitFresnel(node RenderGraphNode) (renderGraphOutputExpression, error) {
	normal, err := c.optionalInputExpression(node, 0, renderGraphOutputVec3, c.geometricNormalExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	view, err := c.optionalInputExpression(node, 1, renderGraphOutputVec3, c.viewDirectionExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	power, err := c.optionalInputOrFloatField(node, 2, "power", "5.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	bias, err := c.optionalInputOrFloatField(node, 3, "bias", "0.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	scale, err := c.optionalInputOrFloatField(node, 4, "scale", "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{
		Type:  renderGraphOutputFloat,
		Value: "graphFresnel(" + normal + ", " + view + ", " + power + ", " + bias + ", " + scale + ")",
	}, nil
}

func (c *renderGraphOutputCompiler) emitRimLight(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	normal, err := c.optionalInputExpression(node, 0, renderGraphOutputVec3, c.geometricNormalExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	view, err := c.optionalInputExpression(node, 1, renderGraphOutputVec3, c.viewDirectionExpression())
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	power, err := c.optionalInputOrFloatField(node, 2, "power", "3.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	intensity, err := c.optionalInputOrFloatField(node, 3, "intensity", "1.0")
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	factor := "graphRimFactor(" + normal + ", " + view + ", " + power + ", " + intensity + ")"
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: factor}, nil
	case 1:
		color, err := c.optionalInputOrColorField(node, 4, "color", matrix.ColorWhite())
		if err != nil {
			return renderGraphOutputExpression{}, err
		}
		return renderGraphOutputExpression{
			Type:  renderGraphOutputColor,
			Value: "vec4((" + color + ").rgb * " + factor + ", (" + color + ").a)",
		}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph rim light node %q has invalid output %d", node.ID, port)
	}
}

func (c *renderGraphOutputCompiler) emitDerivative(node RenderGraphNode) (renderGraphOutputExpression, error) {
	value, err := c.inputExpression(node, 0, renderGraphOutputFloat)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	fn := "fwidth"
	switch node.Type {
	case "ddx":
		fn = "dFdx"
	case "ddy":
		fn = "dFdy"
	}
	return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: fn + "(" + value + ")"}, nil
}

func emitColorLikeOutput(node RenderGraphNode, port int, value, label string) (renderGraphOutputExpression, error) {
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: value}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputVec3, Value: value + ".rgb"}, nil
	case 2:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: value + ".r"}, nil
	case 3:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: value + ".g"}, nil
	case 4:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: value + ".b"}, nil
	case 5:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: value + ".a"}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph %s node %q has invalid output %d", label, node.ID, port)
	}
}

func (c *renderGraphOutputCompiler) emitCombineVector(node RenderGraphNode, port int, constructor, outputType string, count int) (renderGraphOutputExpression, error) {
	components := make([]string, count)
	for i := range components {
		value, err := c.optionalInputExpression(node, i, renderGraphOutputFloat, "0.0")
		if err != nil {
			return renderGraphOutputExpression{}, err
		}
		components[i] = value
	}
	value := constructor + "(" + strings.Join(components, ", ") + ")"
	if outputType == renderGraphOutputVec4 && port == 1 {
		return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: value}, nil
	}
	if port != 0 {
		return renderGraphOutputExpression{}, fmt.Errorf("render graph combine vector node %q has invalid output %d", node.ID, port)
	}
	return renderGraphOutputExpression{Type: outputType, Value: value}, nil
}

func (c *renderGraphOutputCompiler) emitSplitVector(node RenderGraphNode, port int, inputType string, count int) (renderGraphOutputExpression, error) {
	if port < 0 || port >= count {
		return renderGraphOutputExpression{}, fmt.Errorf("render graph split vector node %q has invalid output %d", node.ID, port)
	}
	value, err := c.inputExpression(node, 0, inputType)
	if err != nil {
		return renderGraphOutputExpression{}, err
	}
	return renderGraphOutputExpression{
		Type:  renderGraphOutputFloat,
		Value: "(" + value + ")." + "xyzw"[port:port+1],
	}, nil
}

func (c *renderGraphOutputCompiler) emitSwizzleVector(node RenderGraphNode, port int, inputType, constructor string, count int) (renderGraphOutputExpression, error) {
	if port != 0 {
		if inputType == renderGraphOutputVec4 && port == 1 {
			value, err := c.emitSwizzleVectorValue(node, inputType, constructor, count)
			return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: value}, err
		}
		return renderGraphOutputExpression{}, fmt.Errorf("render graph swizzle node %q has invalid output %d", node.ID, port)
	}
	value, err := c.emitSwizzleVectorValue(node, inputType, constructor, count)
	return renderGraphOutputExpression{Type: inputType, Value: value}, err
}

func (c *renderGraphOutputCompiler) emitSwizzleVectorValue(node RenderGraphNode, inputType, constructor string, count int) (string, error) {
	value, err := c.inputExpression(node, 0, inputType)
	if err != nil {
		return "", err
	}
	fields := []string{"x", "y", "z", "w"}[:count]
	components := make([]string, count)
	for i := range components {
		selection := c.fieldValue(node, fields[i]).Option
		if selection == "" {
			selection = fields[i]
		}
		component, err := renderGraphSwizzleComponentExpression(value, selection, count)
		if err != nil {
			return "", fmt.Errorf("render graph swizzle node %q component %q: %w", node.ID, fields[i], err)
		}
		components[i] = component
	}
	return constructor + "(" + strings.Join(components, ", ") + ")", nil
}

func renderGraphSwizzleComponentExpression(value, selection string, count int) (string, error) {
	switch strings.ToLower(strings.TrimSpace(selection)) {
	case "x", "r":
		return "(" + value + ").x", nil
	case "y", "g":
		if count < 2 {
			return "", fmt.Errorf("selection %q is unavailable for vec%d", selection, count)
		}
		return "(" + value + ").y", nil
	case "z", "b":
		if count < 3 {
			return "", fmt.Errorf("selection %q is unavailable for vec%d", selection, count)
		}
		return "(" + value + ").z", nil
	case "w", "a":
		if count < 4 {
			return "", fmt.Errorf("selection %q is unavailable for vec%d", selection, count)
		}
		return "(" + value + ").w", nil
	case "0":
		return "0.0", nil
	case "1":
		return "1.0", nil
	default:
		return "", fmt.Errorf("unsupported selection %q", selection)
	}
}

func emitTime(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "time"}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph time node %q has invalid output %d", node.ID, port)
	}
}

func emitVec3Context(node RenderGraphNode, port int, value string) (renderGraphOutputExpression, error) {
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputVec3, Value: value}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + value + ").x"}, nil
	case 2:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + value + ").y"}, nil
	case 3:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "(" + value + ").z"}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph context node %q has invalid output %d", node.ID, port)
	}
}

func emitScreenPosition(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	uv := "(gl_FragCoord.xy / max(screenSize, vec2(1.0)))"
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputVec2, Value: uv}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputVec2, Value: "gl_FragCoord.xy"}, nil
	case 2:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: uv + ".x"}, nil
	case 3:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: uv + ".y"}, nil
	case 4:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "gl_FragCoord.z"}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph screen position node %q has invalid output %d", node.ID, port)
	}
}

func emitVertexColor(node RenderGraphNode, port int) (renderGraphOutputExpression, error) {
	switch port {
	case 0:
		return renderGraphOutputExpression{Type: renderGraphOutputColor, Value: "fragColor"}, nil
	case 1:
		return renderGraphOutputExpression{Type: renderGraphOutputVec3, Value: "fragColor.rgb"}, nil
	case 2:
		return renderGraphOutputExpression{Type: renderGraphOutputFloat, Value: "fragColor.a"}, nil
	default:
		return renderGraphOutputExpression{}, fmt.Errorf("render graph vertex color node %q has invalid output %d", node.ID, port)
	}
}

func graphGeometricNormalExpression() string {
	return "safeNormalize(fragNormal, vec3(0.0, 1.0, 0.0))"
}

func (c *renderGraphOutputCompiler) uvExpression() string {
	if c.vertexStage {
		return "graphVertexUV"
	}
	return "fragTexCoords"
}

func (c *renderGraphOutputCompiler) worldPositionExpression() string {
	if c.vertexStage {
		return "graphVertexWorldPosition.xyz"
	}
	return "fragPos"
}

func (c *renderGraphOutputCompiler) geometricNormalExpression() string {
	if c.vertexStage {
		return "graphVertexWorldNormal"
	}
	return graphGeometricNormalExpression()
}

func (c *renderGraphOutputCompiler) viewDirectionExpression() string {
	return "safeNormalize(cameraPosition.xyz - " + c.worldPositionExpression() + ", " + c.geometricNormalExpression() + ")"
}

func (c *renderGraphOutputCompiler) inputExpression(node RenderGraphNode, input int, wantType string) (string, error) {
	ref, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: input}]
	if !ok {
		return "", fmt.Errorf("render graph node %q input %d is disconnected", node.ID, input)
	}
	expr, err := c.emitExpression(ref, wantType)
	if err != nil {
		return "", err
	}
	return expr.Value, nil
}

func (c *renderGraphOutputCompiler) optionalInputExpression(node RenderGraphNode, input int, wantType, fallback string) (string, error) {
	ref, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: input}]
	if !ok {
		return fallback, nil
	}
	expr, err := c.emitExpression(ref, wantType)
	if err != nil {
		return "", err
	}
	return expr.Value, nil
}

func (c *renderGraphOutputCompiler) optionalInputOrFloatField(node RenderGraphNode, input int, field, fallback string) (string, error) {
	ref, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: input}]
	if ok {
		expr, err := c.emitExpression(ref, renderGraphOutputFloat)
		if err != nil {
			return "", err
		}
		return expr.Value, nil
	}
	if strings.TrimSpace(field) == "" {
		return fallback, nil
	}
	value, err := c.floatField(node, field)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (c *renderGraphOutputCompiler) optionalInputOrColorField(node RenderGraphNode, input int, field string, fallback matrix.Color) (string, error) {
	ref, ok := c.incoming[RenderGraphPortRef{Node: node.ID, Port: input}]
	if ok {
		expr, err := c.emitExpression(ref, renderGraphOutputColor)
		if err != nil {
			return "", err
		}
		return expr.Value, nil
	}
	if strings.TrimSpace(field) == "" {
		return glslColor(fallback), nil
	}
	return c.colorFieldExpression(node, field), nil
}

func (c *renderGraphOutputCompiler) colorFieldExpression(node RenderGraphNode, field string) string {
	return glslColor(c.fieldValue(node, field).Color)
}

func (c *renderGraphOutputCompiler) uniqueTextureLabel(label string) string {
	used := map[string]bool{}
	for i := range c.textures {
		used[c.textures[i].Label] = true
	}
	if !used[label] {
		return label
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s %d", label, i)
		if !used[candidate] {
			return candidate
		}
	}
}

func (c *renderGraphOutputCompiler) textureColorSpace(node RenderGraphNode) string {
	if c.fieldValue(node, "color-space").Option == "linear" {
		return "linear"
	}
	return "srgb"
}

func (c *renderGraphOutputCompiler) fieldValue(node RenderGraphNode, id string) renderGraphNodeFieldValue {
	if value, ok := node.Values[id]; ok {
		out := renderGraphNodeFieldValue{
			Text:   value.Text,
			Parts:  append([]string(nil), value.Parts...),
			Option: value.Option,
		}
		if value.Bool != nil {
			out.Bool = *value.Bool
		}
		if value.Color != nil {
			out.Color = *value.Color
		}
		return out
	}
	spec := c.specs[node.ID]
	for i := range spec.Fields {
		if spec.Fields[i].ID == id {
			return renderGraphDefaultFieldValue(spec.Fields[i])
		}
	}
	return renderGraphNodeFieldValue{}
}

func (c *renderGraphOutputCompiler) floatField(node RenderGraphNode, id string) (string, error) {
	return glslFloatFromText(c.fieldValue(node, id).Text)
}

func (c *renderGraphOutputCompiler) vectorField(node RenderGraphNode, id string) (string, error) {
	parts := c.fieldValue(node, id).Parts
	if len(parts) < 3 {
		parts = renderGraphFieldParts(parts, 3)
	}
	x, err := glslFloatFromText(parts[0])
	if err != nil {
		return "", err
	}
	y, err := glslFloatFromText(parts[1])
	if err != nil {
		return "", err
	}
	z, err := glslFloatFromText(parts[2])
	if err != nil {
		return "", err
	}
	return "vec3(" + x + ", " + y + ", " + z + ")", nil
}

func (c *renderGraphOutputCompiler) vector2Field(node RenderGraphNode, id string) (string, error) {
	parts := c.fieldValue(node, id).Parts
	if len(parts) < 2 {
		parts = renderGraphFieldParts(parts, 2)
	}
	x, err := glslFloatFromText(parts[0])
	if err != nil {
		return "", err
	}
	y, err := glslFloatFromText(parts[1])
	if err != nil {
		return "", err
	}
	return "vec2(" + x + ", " + y + ")", nil
}

func (c *renderGraphOutputCompiler) vector4Field(node RenderGraphNode, id string) (string, error) {
	parts := c.fieldValue(node, id).Parts
	if len(parts) < 4 {
		parts = renderGraphFieldParts(parts, 4)
	}
	x, err := glslFloatFromText(parts[0])
	if err != nil {
		return "", err
	}
	y, err := glslFloatFromText(parts[1])
	if err != nil {
		return "", err
	}
	z, err := glslFloatFromText(parts[2])
	if err != nil {
		return "", err
	}
	w, err := glslFloatFromText(parts[3])
	if err != nil {
		return "", err
	}
	return "vec4(" + x + ", " + y + ", " + z + ", " + w + ")", nil
}

func glslColor(value matrix.Color) string {
	return fmt.Sprintf("vec4(%s, %s, %s, %s)",
		glslFloat(float64(value.R())),
		glslFloat(float64(value.G())),
		glslFloat(float64(value.B())),
		glslFloat(float64(value.A())),
	)
}

func glslFloatFromText(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "0"
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", fmt.Errorf("invalid numeric shader graph value %q: %w", value, err)
	}
	if math.IsInf(parsed, 0) || math.IsNaN(parsed) {
		return "", fmt.Errorf("invalid non-finite shader graph value %q", value)
	}
	return glslFloat(parsed), nil
}

func glslFloat(value float64) string {
	if value == 0 {
		value = 0
	}
	out := strconv.FormatFloat(value, 'f', -1, 64)
	if !strings.ContainsAny(out, ".eE") {
		out += ".0"
	}
	return out
}

func renderGraphDefaultTextureSlots() []rendering.MaterialTextureData {
	return []rendering.MaterialTextureData{
		{Label: "Diffuse", Texture: assets.TextureSquare, Filter: "Linear"},
		{Label: "Normal", Texture: assets.TexturePBRDefaultNormal, Filter: "Linear"},
		{Label: "Metallic Roughness", Texture: assets.TexturePBRDefaultMetallicRough, Filter: "Linear"},
		{Label: "Emissive", Texture: assets.TextureBlankSquare, Filter: "Linear"},
	}
}

func renderGraphSamplerLabels(textures []rendering.MaterialTextureData) []string {
	labels := make([]string, len(textures))
	for i := range textures {
		labels[i] = textures[i].Label
	}
	return labels
}

func renderGraphPBRVertexSource(displacement string) string {
	displacement = strings.TrimSpace(displacement)
	if displacement == "" {
		displacement = "0.0"
	}
	return fmt.Sprintf(`#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR 0
#define LAYOUT_VERT_METALLIC_ROUGHNESS_EMISSIVE_ALBEDO 1
#define LAYOUT_VERT_FLAGS 2
#define LAYOUT_VERT_LIGHT_IDS 3

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_FLAGS 1
#define LAYOUT_FRAG_POS 2
#define LAYOUT_FRAG_TEX_COORDS 3
#define LAYOUT_FRAG_NORMAL 4
#define LAYOUT_FRAG_METALLIC 5
#define LAYOUT_FRAG_ROUGHNESS 6
#define LAYOUT_FRAG_EMISSIVE 7

#define LAYOUT_ALL_LIGHT_REQUIREMENTS 8

#include "kaiju.glsl"

vec2 graphVertexUV;
vec3 graphVertexWorldNormal;
vec4 graphVertexWorldPosition;

vec3 safeNormalize(vec3 v, vec3 fallback) {
	float len2 = dot(v, v);
	if (len2 <= 0.00000001) {
		return fallback;
	}
	return v * inversesqrt(len2);
}

vec3 graphApplyNormalStrength(vec3 normal, float strength, vec3 geometricNormal) {
	return safeNormalize(mix(geometricNormal, normal, max(strength, 0.0)), geometricNormal);
}

vec3 graphBlendNormals(vec3 base, vec3 detail, float strength, vec3 geometricNormal) {
	vec3 safeBase = safeNormalize(base, geometricNormal);
	vec3 safeDetail = safeNormalize(detail, geometricNormal);
	return safeNormalize(safeBase + (safeDetail - geometricNormal) * max(strength, 0.0), safeBase);
}

float graphHash12(vec2 p) {
	vec3 p3 = fract(vec3(p.xyx) * 0.1031);
	p3 += dot(p3, p3.yzx + 33.33);
	return fract((p3.x + p3.y) * p3.z);
}

vec2 graphHash22(vec2 p) {
	vec3 p3 = fract(vec3(p.xyx) * vec3(0.1031, 0.1030, 0.0973));
	p3 += dot(p3, p3.yzx + 33.33);
	return fract((p3.xx + p3.yz) * p3.zy);
}

float graphNoise2D(vec2 p) {
	vec2 i = floor(p);
	vec2 f = fract(p);
	vec2 u = f * f * (3.0 - 2.0 * f);
	float a = graphHash12(i + vec2(0.0, 0.0));
	float b = graphHash12(i + vec2(1.0, 0.0));
	float c = graphHash12(i + vec2(0.0, 1.0));
	float d = graphHash12(i + vec2(1.0, 1.0));
	return mix(mix(a, b, u.x), mix(c, d, u.x), u.y);
}

float graphFBM2D(vec2 p, float detail, float roughness) {
	float layers = clamp(floor(detail), 1.0, 8.0);
	float amplitude = 0.5;
	float frequency = 1.0;
	float sum = 0.0;
	float weight = 0.0;
	for (int i = 0; i < 8; ++i) {
		if (float(i) >= layers) {
			break;
		}
		sum += graphNoise2D(p * frequency) * amplitude;
		weight += amplitude;
		frequency *= 2.0;
		amplitude *= clamp(roughness, 0.0, 1.0);
	}
	return clamp(sum / max(weight, 0.0001), 0.0, 1.0);
}

vec3 graphVoronoi2D(vec2 p, float jitter) {
	vec2 baseCell = floor(p);
	vec2 local = fract(p);
	float best = 8.0;
	float second = 8.0;
	float cellValue = 0.0;
	for (int y = -1; y <= 1; ++y) {
		for (int x = -1; x <= 1; ++x) {
			vec2 cell = vec2(float(x), float(y));
			vec2 offset = mix(vec2(0.5), graphHash22(baseCell + cell), clamp(jitter, 0.0, 1.0));
			vec2 delta = cell + offset - local;
			float dist = dot(delta, delta);
			if (dist < best) {
				second = best;
				best = dist;
				cellValue = graphHash12(baseCell + cell);
			} else if (dist < second) {
				second = dist;
			}
		}
	}
	float nearest = sqrt(best);
	float edge = max(sqrt(second) - nearest, 0.0);
	return vec3(nearest, cellValue, edge);
}

float graphCheckerMask(vec2 uv, float scale) {
	vec2 cell = floor(uv * max(abs(scale), 0.0001));
	return mod(cell.x + cell.y, 2.0);
}

float graphGradientFactor(vec2 uv, float angle, float radialMode) {
	vec2 centered = uv - vec2(0.5);
	vec2 direction = vec2(cos(angle), sin(angle));
	float linear = dot(centered, direction) + 0.5;
	float radial = length(centered) * 2.0;
	return clamp(mix(linear, radial, step(0.5, radialMode)), 0.0, 1.0);
}

float graphRemap(float value, float inMin, float inMax, float outMin, float outMax) {
	float denom = inMax - inMin;
	if (abs(denom) <= 0.00000001) {
		return outMin;
	}
	float t = (value - inMin) / denom;
	return mix(outMin, outMax, t);
}

float graphPosterize(float value, float steps) {
	float levels = max(floor(steps), 2.0);
	float v = clamp(value, 0.0, 1.0);
	return floor(v * (levels - 1.0) + 0.5) / (levels - 1.0);
}

vec4 graphPosterizeColor(vec4 color, float steps) {
	return vec4(
		graphPosterize(color.r, steps),
		graphPosterize(color.g, steps),
		graphPosterize(color.b, steps),
		graphPosterize(color.a, steps)
	);
}

float graphFresnel(vec3 normal, vec3 viewDir, float power, float bias, float scale) {
	vec3 n = safeNormalize(normal, vec3(0.0, 1.0, 0.0));
	vec3 v = safeNormalize(viewDir, vec3(0.0, 0.0, 1.0));
	float facing = max(dot(n, v), 0.0);
	return clamp(bias + scale * pow(1.0 - facing, max(power, 0.0001)), 0.0, 1.0);
}

float graphRimFactor(vec3 normal, vec3 viewDir, float power, float intensity) {
	return clamp(graphFresnel(normal, viewDir, power, 0.0, 1.0) * max(intensity, 0.0), 0.0, 1.0);
}

vec4 graphWorldPositionFromLocal(vec3 localPosition) {
#ifdef SKINNING
	return skinMatrix() * vec4(localPosition, 1.0);
#else
	return model * vec4(localPosition, 1.0);
#endif
}

void main() {
	fragMetallic = meRoEmAo.r;
	fragRoughness = meRoEmAo.g;
	fragEmissive = meRoEmAo.b;
	fragFlags = flags;
	fragTexCoords = UV0;
	fragColor = color * Color;

	graphVertexUV = UV0;
	graphVertexWorldNormal = safeNormalize(transpose(inverse(mat3(model))) * Normal, vec3(0.0, 1.0, 0.0));
	graphVertexWorldPosition = graphWorldPositionFromLocal(Position);

	vec3 graphVertexLocalNormal = safeNormalize(Normal, vec3(0.0, 1.0, 0.0));
	float graphDisplacement = %s;
	vec3 graphDisplacedPosition = Position + graphVertexLocalNormal * graphDisplacement;
	graphVertexWorldPosition = graphWorldPositionFromLocal(graphDisplacedPosition);

	fragPos = graphVertexWorldPosition.xyz;
	gl_Position = projection * view * graphVertexWorldPosition;
	calcVertexLightInformation();
}
`, displacement)
}

func renderGraphPBRFragmentSource(surface renderGraphOutputSurface, samplerCount int) string {
	if samplerCount < 4 {
		samplerCount = 4
	}
	metallicExpr := "clamp(mrSample.b * max(" + surface.Metallic + ", 0.0), 0.0, 1.0)"
	if !surface.UseTextureMetallic {
		metallicExpr = "clamp(" + surface.Metallic + ", 0.0, 1.0)"
	}
	roughnessExpr := "clamp(mrSample.g * max(" + surface.Roughness + ", MIN_ROUGHNESS), MIN_ROUGHNESS, 1.0)"
	if !surface.UseTextureRoughness {
		roughnessExpr = "clamp(" + surface.Roughness + ", MIN_ROUGHNESS, 1.0)"
	}
	occlusionExpr := "clamp(mrSample.r, 0.0, 1.0)"
	if !surface.UseTextureOcclusion {
		occlusionExpr = "clamp(" + surface.Occlusion + ", 0.0, 1.0)"
	}
	normalExpr := surface.Normal
	if surface.UseTextureNormal {
		normalExpr = "pbrNormal(geometricNormal)"
	}
	emissionColorExpr := "srgbToLinear(texture(textures[3], fragTexCoords).rgb)"
	if !surface.UseTextureEmission {
		emissionColorExpr = surface.EmissionColor
	}
	emissionExpr := "max(" + emissionColorExpr + ", vec3(0.0)) * max(" + surface.EmissionStrength + ", 0.0)"
	alphaExpr := surface.Alpha
	if surface.UseAlphaInput {
		alphaExpr = "clamp(" + surface.Alpha + ", 0.0, 1.0)"
	}
	specularExpr := "clamp(" + surface.Specular + ", 0.0, 1.0)"
	return fmt.Sprintf(`#version 460
#define FRAGMENT_SHADER
#define HAS_GBUFFER

#define SAMPLER_COUNT   %d
#define SHADOW_SAMPLERS

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_FLAGS 1
#define LAYOUT_FRAG_POS 2
#define LAYOUT_FRAG_TEX_COORDS 3
#define LAYOUT_FRAG_NORMAL 4
#define LAYOUT_FRAG_METALLIC 5
#define LAYOUT_FRAG_ROUGHNESS 6
#define LAYOUT_FRAG_EMISSIVE 7

#define LAYOUT_ALL_LIGHT_REQUIREMENTS 8

#include "kaiju.glsl"

const float MIN_ROUGHNESS = 0.045;
const float DEFAULT_AMBIENT_STRENGTH = 0.03;
const float MIN_TBN_DERIVATIVE_LEN2 = 1e-20;

vec3 safeNormalize(vec3 v, vec3 fallback) {
	float len2 = dot(v, v);
	if (len2 <= 0.00000001) {
		return fallback;
	}
	return v * inversesqrt(len2);
}

vec3 srgbToLinear(vec3 color) {
	return pow(max(color, vec3(0.0)), vec3(2.2));
}

vec3 linearToSrgb(vec3 color) {
	return pow(max(color, vec3(0.0)), vec3(1.0 / 2.2));
}

vec4 graphSrgbToLinear(vec4 color) {
	return vec4(srgbToLinear(color.rgb), color.a);
}

vec3 acesTonemap(vec3 color) {
	const float a = 2.51;
	const float b = 0.03;
	const float c = 2.43;
	const float d = 0.59;
	const float e = 0.14;
	return clamp((color * (a * color + b)) / (color * (c * color + d) + e), 0.0, 1.0);
}

mat3 fallbackTBN(vec3 n) {
	vec3 up = abs(n.z) < 0.999 ? vec3(0.0, 0.0, 1.0) : vec3(0.0, 1.0, 0.0);
	vec3 t = normalize(cross(up, n));
	vec3 b = cross(n, t);
	return mat3(t, b, n);
}

mat3 cotangentFrame(vec3 n, vec3 pos, vec2 uv) {
	vec3 dp1 = dFdx(pos);
	vec3 dp2 = dFdy(pos);
	vec2 duv1 = dFdx(uv);
	vec2 duv2 = dFdy(uv);
	vec3 dp2Perp = cross(dp2, n);
	vec3 dp1Perp = cross(n, dp1);
	vec3 t = dp2Perp * duv1.x + dp1Perp * duv2.x;
	vec3 b = dp2Perp * duv1.y + dp1Perp * duv2.y;
	float maxLen = max(dot(t, t), dot(b, b));
	if (maxLen <= MIN_TBN_DERIVATIVE_LEN2) {
		return fallbackTBN(n);
	}
	float invMax = inversesqrt(maxLen);
	return mat3(t * invMax, b * invMax, n);
}

vec3 graphTangentNormalFromMap(vec3 sampleRGB, float strength, float flipY) {
	vec3 tangentNormal = sampleRGB * 2.0 - 1.0;
	tangentNormal.y *= flipY;
	tangentNormal.xy *= max(strength, 0.0);
	if (dot(tangentNormal, tangentNormal) <= 0.0001) {
		return vec3(0.0, 0.0, 1.0);
	}
	return normalize(tangentNormal);
}

vec3 graphWorldNormalFromTangent(vec3 tangentNormal, vec2 uv, vec3 geometricNormal) {
	mat3 tbn = cotangentFrame(geometricNormal, fragPos, uv);
	return safeNormalize(tbn * safeNormalize(tangentNormal, vec3(0.0, 0.0, 1.0)), geometricNormal);
}

vec3 graphApplyNormalStrength(vec3 normal, float strength, vec3 geometricNormal) {
	return safeNormalize(mix(geometricNormal, normal, max(strength, 0.0)), geometricNormal);
}

vec3 graphBlendNormals(vec3 base, vec3 detail, float strength, vec3 geometricNormal) {
	vec3 safeBase = safeNormalize(base, geometricNormal);
	vec3 safeDetail = safeNormalize(detail, geometricNormal);
	return safeNormalize(safeBase + (safeDetail - geometricNormal) * max(strength, 0.0), safeBase);
}

vec3 graphBumpNormal(float height, float strength, vec3 geometricNormal) {
	vec3 dpdx = dFdx(fragPos);
	vec3 dpdy = dFdy(fragPos);
	float dhdx = dFdx(height);
	float dhdy = dFdy(height);
	vec3 r1 = cross(dpdy, geometricNormal);
	vec3 r2 = cross(geometricNormal, dpdx);
	float det = dot(dpdx, r1);
	vec3 gradient = sign(det) * (dhdx * r1 + dhdy * r2);
	return safeNormalize(abs(det) * geometricNormal - max(strength, 0.0) * gradient, geometricNormal);
}

vec2 graphParallaxUV(vec2 uv, float height, float scale, vec3 geometricNormal) {
	vec3 viewDir = safeNormalize(cameraPosition.xyz - fragPos, geometricNormal);
	mat3 tbn = cotangentFrame(geometricNormal, fragPos, uv);
	vec3 tangentView = transpose(tbn) * viewDir;
	float denom = max(abs(tangentView.z), 0.05);
	float centeredHeight = height - 0.5;
	return uv - (tangentView.xy / denom) * centeredHeight * scale;
}

vec4 graphTriplanarSample(sampler2D tex, vec3 position, vec3 normal, float scale, float blendPower) {
	vec3 weights = pow(max(abs(safeNormalize(normal, vec3(0.0, 1.0, 0.0))), vec3(0.0001)), vec3(max(blendPower, 0.0001)));
	weights /= max(weights.x + weights.y + weights.z, 0.0001);
	vec3 p = position * max(abs(scale), 0.0001);
	vec4 xSample = texture(tex, p.zy);
	vec4 ySample = texture(tex, p.xz);
	vec4 zSample = texture(tex, p.xy);
	return xSample * weights.x + ySample * weights.y + zSample * weights.z;
}

vec4 graphOverlayColor(vec4 base, vec4 detail) {
	vec3 low = 2.0 * base.rgb * detail.rgb;
	vec3 high = 1.0 - 2.0 * (1.0 - base.rgb) * (1.0 - detail.rgb);
	return vec4(mix(low, high, step(vec3(0.5), base.rgb)), base.a);
}

float graphHash12(vec2 p) {
	vec3 p3 = fract(vec3(p.xyx) * 0.1031);
	p3 += dot(p3, p3.yzx + 33.33);
	return fract((p3.x + p3.y) * p3.z);
}

vec2 graphHash22(vec2 p) {
	vec3 p3 = fract(vec3(p.xyx) * vec3(0.1031, 0.1030, 0.0973));
	p3 += dot(p3, p3.yzx + 33.33);
	return fract((p3.xx + p3.yz) * p3.zy);
}

float graphNoise2D(vec2 p) {
	vec2 i = floor(p);
	vec2 f = fract(p);
	vec2 u = f * f * (3.0 - 2.0 * f);
	float a = graphHash12(i + vec2(0.0, 0.0));
	float b = graphHash12(i + vec2(1.0, 0.0));
	float c = graphHash12(i + vec2(0.0, 1.0));
	float d = graphHash12(i + vec2(1.0, 1.0));
	return mix(mix(a, b, u.x), mix(c, d, u.x), u.y);
}

float graphFBM2D(vec2 p, float detail, float roughness) {
	float layers = clamp(floor(detail), 1.0, 8.0);
	float amplitude = 0.5;
	float frequency = 1.0;
	float sum = 0.0;
	float weight = 0.0;
	for (int i = 0; i < 8; ++i) {
		if (float(i) >= layers) {
			break;
		}
		sum += graphNoise2D(p * frequency) * amplitude;
		weight += amplitude;
		frequency *= 2.0;
		amplitude *= clamp(roughness, 0.0, 1.0);
	}
	return clamp(sum / max(weight, 0.0001), 0.0, 1.0);
}

vec3 graphVoronoi2D(vec2 p, float jitter) {
	vec2 baseCell = floor(p);
	vec2 local = fract(p);
	float best = 8.0;
	float second = 8.0;
	float cellValue = 0.0;
	for (int y = -1; y <= 1; ++y) {
		for (int x = -1; x <= 1; ++x) {
			vec2 cell = vec2(float(x), float(y));
			vec2 offset = mix(vec2(0.5), graphHash22(baseCell + cell), clamp(jitter, 0.0, 1.0));
			vec2 delta = cell + offset - local;
			float dist = dot(delta, delta);
			if (dist < best) {
				second = best;
				best = dist;
				cellValue = graphHash12(baseCell + cell);
			} else if (dist < second) {
				second = dist;
			}
		}
	}
	float nearest = sqrt(best);
	float edge = max(sqrt(second) - nearest, 0.0);
	return vec3(nearest, cellValue, edge);
}

float graphCheckerMask(vec2 uv, float scale) {
	vec2 cell = floor(uv * max(abs(scale), 0.0001));
	return mod(cell.x + cell.y, 2.0);
}

float graphGradientFactor(vec2 uv, float angle, float radialMode) {
	vec2 centered = uv - vec2(0.5);
	vec2 direction = vec2(cos(angle), sin(angle));
	float linear = dot(centered, direction) + 0.5;
	float radial = length(centered) * 2.0;
	return clamp(mix(linear, radial, step(0.5, radialMode)), 0.0, 1.0);
}

float graphRemap(float value, float inMin, float inMax, float outMin, float outMax) {
	float denom = inMax - inMin;
	if (abs(denom) <= 0.00000001) {
		return outMin;
	}
	float t = (value - inMin) / denom;
	return mix(outMin, outMax, t);
}

float graphPosterize(float value, float steps) {
	float levels = max(floor(steps), 2.0);
	float v = clamp(value, 0.0, 1.0);
	return floor(v * (levels - 1.0) + 0.5) / (levels - 1.0);
}

vec4 graphPosterizeColor(vec4 color, float steps) {
	return vec4(
		graphPosterize(color.r, steps),
		graphPosterize(color.g, steps),
		graphPosterize(color.b, steps),
		graphPosterize(color.a, steps)
	);
}

float graphFresnel(vec3 normal, vec3 viewDir, float power, float bias, float scale) {
	vec3 n = safeNormalize(normal, vec3(0.0, 1.0, 0.0));
	vec3 v = safeNormalize(viewDir, vec3(0.0, 0.0, 1.0));
	float facing = max(dot(n, v), 0.0);
	return clamp(bias + scale * pow(1.0 - facing, max(power, 0.0001)), 0.0, 1.0);
}

float graphRimFactor(vec3 normal, vec3 viewDir, float power, float intensity) {
	return clamp(graphFresnel(normal, viewDir, power, 0.0, 1.0) * max(intensity, 0.0), 0.0, 1.0);
}

vec3 pbrNormal(vec3 geometricNormal) {
	vec3 normalSample = texture(textures[1], fragTexCoords).rgb;
	vec3 tangentNormal = normalSample * 2.0 - 1.0;
	bool whiteFallback = all(greaterThanEqual(normalSample, vec3(0.999)));
	if (whiteFallback || dot(tangentNormal, tangentNormal) <= 0.0001) {
		tangentNormal = vec3(0.0, 0.0, 1.0);
	}
	mat3 tbn = cotangentFrame(geometricNormal, fragPos, fragTexCoords);
	return normalize(tbn * normalize(tangentNormal));
}

float distanceAttenuation(LightInfo light, float dist) {
	float denom = light.constant + light.linear * dist + light.quadratic * dist * dist;
	return max(light.intensity, 0.0) / max(denom, 0.0001);
}

float lightVisibility(int lightType, int lightIdx, vec3 n, vec3 l, vec4 lightSpace, LightInfo light) {
	#ifdef SHADOW_SAMPLERS
		if (lightType == 0) {
			return 1.0 - directShadowCalculation(n, l, lightIdx, light.farPlane);
		}
		if (lightType == 1) {
			return 1.0 - pointShadowCalculation(fragPos, light.position, light.farPlane, lightIdx, n);
		}
		if (lightType == 2) {
			return 1.0 - spotShadowCalculation(lightSpace, n, l, light.nearPlane, light.farPlane, lightIdx);
		}
	#endif
	return 1.0;
}

void main() {
	vec4 baseSample = texture(textures[0], fragTexCoords);
	vec4 graphBaseColor = %s;
	vec3 albedo = srgbToLinear(baseSample.rgb) * max(graphBaseColor.rgb, vec3(0.0));
	float alpha = %s;

	vec4 mrSample = texture(textures[2], fragTexCoords);
	float metallic = %s;
	float roughness = %s;
	float occlusion = %s;
	vec3 emission = %s;

	vec3 geometricNormal = safeNormalize(fragNormal, vec3(0.0, 1.0, 0.0));
	vec3 N = %s;
	vec3 V = safeNormalize(cameraPosition.xyz - fragPos, geometricNormal);
	float NdotV = max(dot(N, V), 0.0);

	processGBuffer(N);

	vec3 F0 = mix(vec3(0.04 * %s), albedo, metallic);
	vec3 Lo = vec3(0.0);
	vec3 ambient = vec3(DEFAULT_AMBIENT_STRENGTH) * albedo * occlusion;

	for (int i = 0; i < fragLightCount; ++i) {
		int lightIdx = fragLightIndexes[i];
		if (lightIdx < 0 || lightIdx >= MAX_LIGHTS) {
			continue;
		}
		LightInfo light = lightInfos[lightIdx];
		vec3 L = vec3(0.0);
		float attenuation = 0.0;
		if (light.type == 0) {
			L = safeNormalize(-light.direction, geometricNormal);
			attenuation = max(light.intensity, 0.0);
		} else if (light.type == 1) {
			vec3 toLight = light.position - fragPos;
			float dist = length(toLight);
			L = safeNormalize(toLight, geometricNormal);
			attenuation = distanceAttenuation(light, dist);
		} else if (light.type == 2) {
			vec3 toLight = light.position - fragPos;
			float dist = length(toLight);
			L = safeNormalize(toLight, geometricNormal);
			attenuation = distanceAttenuation(light, dist);
			vec3 lightToFrag = safeNormalize(fragPos - light.position, -L);
			float theta = dot(safeNormalize(light.direction, -L), lightToFrag);
			float epsilon = max(light.cutoff - light.outerCutoff, 0.0001);
			attenuation *= clamp((theta - light.outerCutoff) / epsilon, 0.0, 1.0);
		} else {
			continue;
		}

		float NdotL = max(dot(N, L), 0.0);
		if (attenuation <= 0.0 || NdotL <= 0.0) {
			continue;
		}

		vec3 H = safeNormalize(V + L, N);
		float NDF = distributionGGX(N, H, roughness);
		float G = geometrySmith(N, V, L, roughness);
		vec3 F = fresnelSchlick(max(dot(H, V), 0.0), F0);
		vec3 kD = (vec3(1.0) - F) * (1.0 - metallic);
		vec3 specular = (NDF * G * F) / max(4.0 * NdotV * NdotL, 0.001);
		vec3 radiance = max(light.diffuse, vec3(0.0)) * attenuation;
		float visibility = lightVisibility(light.type, lightIdx, N, L, fragPosLightSpace[i], light);
		Lo += (kD * albedo / PI + specular) * radiance * NdotL * visibility;
		ambient += max(light.ambient, vec3(0.0)) * albedo * occlusion;
	}

	vec3 color = ambient + Lo + emission;
	color = linearToSrgb(acesTonemap(color));
	processFinalColor(vec4(color, alpha));
}
`, samplerCount, surface.BaseColor, alphaExpr, metallicExpr, roughnessExpr,
		occlusionExpr, emissionExpr, normalExpr, specularExpr)
}
