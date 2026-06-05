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
	"kaijuengine.com/rendering"
)

const (
	renderGraphOutputFloat = "float"
	renderGraphOutputVec2  = "vec2"
	renderGraphOutputVec3  = "vec3"
	renderGraphOutputColor = "color"
	renderGraphOutputTex2D = "texture2d"
)

type renderGraphCompiledOutput struct {
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
	nodes        map[string]RenderGraphNode
	specs        map[string]shaderGraphNodeSpec
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
	return renderGraphCompiledOutput{
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
		specs:        make(map[string]shaderGraphNodeSpec, len(document.Nodes)),
		incoming:     make(map[RenderGraphPortRef]RenderGraphPortRef, len(document.Connections)),
		cache:        map[RenderGraphPortRef]renderGraphOutputExpression{},
		visiting:     map[RenderGraphPortRef]bool{},
		textures:     renderGraphDefaultTextureSlots(),
		textureSlots: map[string]int{},
	}
	for i := range document.Nodes {
		node := document.Nodes[i]
		spec, _ := shaderGraphNodeCatalogSpec(node.Type)
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
		if shaderGraphPortTypeKey(outputPort.Type) != shaderGraphPortTypeKey(inputPort.Type) {
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
	wantType := shaderGraphPortTypeKey(spec.Outputs[ref.Port].Type)
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
			Value: fmt.Sprintf("vec4(%s, %s, %s, %s)", glslFloat(float64(value.R())), glslFloat(float64(value.G())), glslFloat(float64(value.B())), glslFloat(float64(value.A()))),
		}, nil
	case "vector":
		value, err := c.vectorField(node, "vector")
		return renderGraphOutputExpression{Type: renderGraphOutputVec3, Value: value}, err
	case "texture-2d":
		return c.emitTexture2D(node)
	case "sample-texture-2d":
		return c.emitSampleTexture2D(node, port)
	case "uv":
		return renderGraphOutputExpression{Type: renderGraphOutputVec2, Value: "fragTexCoords"}, nil
	case "uv-transform":
		return c.emitUVTransform(node)
	case "split-rgba":
		return c.emitSplitRGBA(node, port)
	case "channel-mask":
		return c.emitChannelMask(node)
	case "texel-size":
		return c.emitTexelSize(node, port)
	case "add", "subtract", "multiply", "divide", "minimum", "maximum", "power":
		return c.emitFloatBinary(node)
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
	uv, err := c.optionalInputExpression(node, 0, renderGraphOutputVec2, "fragTexCoords")
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

func (c *renderGraphOutputCompiler) fieldValue(node RenderGraphNode, id string) shaderGraphNodeFieldValue {
	if value, ok := node.Values[id]; ok {
		out := shaderGraphNodeFieldValue{
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
			return shaderGraphDefaultFieldValue(spec.Fields[i])
		}
	}
	return shaderGraphNodeFieldValue{}
}

func (c *renderGraphOutputCompiler) floatField(node RenderGraphNode, id string) (string, error) {
	return glslFloatFromText(c.fieldValue(node, id).Text)
}

func (c *renderGraphOutputCompiler) vectorField(node RenderGraphNode, id string) (string, error) {
	parts := c.fieldValue(node, id).Parts
	if len(parts) < 3 {
		parts = shaderGraphFieldParts(parts, 3)
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
		parts = shaderGraphFieldParts(parts, 2)
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
	if (maxLen <= 0.00000001) {
		return fallbackTBN(n);
	}
	float invMax = inversesqrt(maxLen);
	return mat3(t * invMax, b * invMax, n);
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
