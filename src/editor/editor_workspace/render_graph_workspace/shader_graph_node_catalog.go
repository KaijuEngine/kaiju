/******************************************************************************/
/* shader_graph_node_catalog.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"strings"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
)

type shaderGraphNodeCatalogEntry struct {
	ID          string
	Name        string
	Description string
	Tags        []string
	Spec        shaderGraphNodeSpec
}

type shaderGraphNodeMenuData struct {
	ID          string
	Name        string
	Description string
	Search      string
}

type shaderGraphNodePortCompatibility struct {
	Active       bool
	SourceOutput bool
	Type         string
}

func shaderGraphNodeCatalog() []shaderGraphNodeCatalogEntry {
	return []shaderGraphNodeCatalogEntry{
		{
			ID:          "value",
			Name:        "Value",
			Description: "Single float value.",
			Tags:        []string{"float", "number", "constant"},
			Spec: shaderGraphNodeSpec{
				Name:        "Value",
				Description: "Single float value.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "value",
						Label:   "Value",
						Type:    shaderGraphNodeFieldNumber,
						Default: "1.000",
					},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "color",
			Name:        "Color",
			Description: "Single color value.",
			Tags:        []string{"color", "constant", "albedo"},
			Spec: shaderGraphNodeSpec{
				Name:        "Color",
				Description: "Single color value.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:           "color",
						Label:        "Color",
						Type:         shaderGraphNodeFieldColor,
						DefaultColor: matrix.ColorWhite(),
					},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "vector",
			Name:        "Vector",
			Description: "Single vec3 value.",
			Tags:        []string{"vector", "vec3", "constant"},
			Spec: shaderGraphNodeSpec{
				Name:        "Vector",
				Description: "Single vec3 value.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:            "vector",
						Label:         "Vector",
						Type:          shaderGraphNodeFieldVector3,
						DefaultValues: []string{"0", "0", "0"},
					},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
			},
		},
		{
			ID:          "texture-2d",
			Name:        "Texture 2D",
			Description: "Texture asset used by texture sample nodes.",
			Tags:        []string{"texture", "image", "sampler", "asset"},
			Spec: shaderGraphNodeSpec{
				Name:        "Texture 2D",
				Description: "Texture asset used by texture sample nodes.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "texture",
						Label:   "Texture",
						Type:    shaderGraphNodeFieldTexture,
						Default: assets.TextureSquare,
					},
					{
						ID:      "label",
						Label:   "Label",
						Type:    shaderGraphNodeFieldText,
						Default: "Texture",
					},
					{
						ID:      "filter",
						Label:   "Filter",
						Type:    shaderGraphNodeFieldSelect,
						Default: "Linear",
						Options: []shaderGraphNodeFieldOption{
							{Label: "Linear", Value: "Linear"},
							{Label: "Nearest", Value: "Nearest"},
						},
					},
					{
						ID:      "color-space",
						Label:   "Space",
						Type:    shaderGraphNodeFieldSelect,
						Default: "srgb",
						Options: []shaderGraphNodeFieldOption{
							{Label: "sRGB", Value: "srgb"},
							{Label: "Linear", Value: "linear"},
						},
					},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Texture", Type: "texture2D"},
				},
			},
		},
		{
			ID:          "sample-texture-2d",
			Name:        "Sample Texture 2D",
			Description: "Samples a Texture 2D at UV coordinates.",
			Tags:        []string{"texture", "sample", "image", "sampler", "uv"},
			Spec: shaderGraphNodeSpec{
				Name:        "Sample Texture 2D",
				Description: "Samples a Texture 2D at UV coordinates.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Texture", Type: "texture2D"},
					{Name: "UV", Type: "vec2"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
					{Name: "RGB", Type: "vec3"},
					{Name: "R", Type: "float"},
					{Name: "G", Type: "float"},
					{Name: "B", Type: "float"},
					{Name: "A", Type: "float"},
				},
			},
		},
		{
			ID:          "uv",
			Name:        "UV",
			Description: "Primary mesh UV coordinates.",
			Tags:        []string{"texture", "uv", "coordinates", "texcoord"},
			Spec: shaderGraphNodeSpec{
				Name:        "UV",
				Description: "Primary mesh UV coordinates.",
				Outputs: []shaderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
				},
			},
		},
		{
			ID:          "uv-transform",
			Name:        "UV Transform",
			Description: "Applies tiling and offset to UV coordinates.",
			Tags:        []string{"texture", "uv", "tiling", "offset", "coordinates"},
			Spec: shaderGraphNodeSpec{
				Name:        "UV Transform",
				Description: "Applies tiling and offset to UV coordinates.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:            "tiling",
						Label:         "Tiling",
						Type:          shaderGraphNodeFieldVector2,
						DefaultValues: []string{"1", "1"},
					},
					{
						ID:            "offset",
						Label:         "Offset",
						Type:          shaderGraphNodeFieldVector2,
						DefaultValues: []string{"0", "0"},
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
				},
			},
		},
		{
			ID:          "split-rgba",
			Name:        "Split RGBA",
			Description: "Splits a color into scalar channels.",
			Tags:        []string{"texture", "color", "channel", "rgba", "split"},
			Spec: shaderGraphNodeSpec{
				Name:        "Split RGBA",
				Description: "Splits a color into scalar channels.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "R", Type: "float"},
					{Name: "G", Type: "float"},
					{Name: "B", Type: "float"},
					{Name: "A", Type: "float"},
				},
			},
		},
		{
			ID:          "channel-mask",
			Name:        "Channel Mask",
			Description: "Extracts one scalar channel from a texture or color sample.",
			Tags:        []string{"texture", "color", "channel", "mask", "rgba", "luminance"},
			Spec: shaderGraphNodeSpec{
				Name:        "Channel Mask",
				Description: "Extracts one scalar channel from a texture or color sample.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "channel",
						Label:   "Channel",
						Type:    shaderGraphNodeFieldSelect,
						Default: "r",
						Options: []shaderGraphNodeFieldOption{
							{Label: "R", Value: "r"},
							{Label: "G", Value: "g"},
							{Label: "B", Value: "b"},
							{Label: "A", Value: "a"},
							{Label: "Luma", Value: "luma"},
						},
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "texel-size",
			Name:        "Texel Size",
			Description: "Returns inverse texture dimensions for a Texture 2D.",
			Tags:        []string{"texture", "texel", "size", "dimensions", "pixel"},
			Spec: shaderGraphNodeSpec{
				Name:        "Texel Size",
				Description: "Returns inverse texture dimensions for a Texture 2D.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Texture", Type: "texture2D"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Size", Type: "vec2"},
					{Name: "Width", Type: "float"},
					{Name: "Height", Type: "float"},
				},
			},
		},
		{
			ID:          "time",
			Name:        "Time",
			Description: "Runtime shader time values.",
			Tags:        []string{"context", "time", "animation", "runtime", "clock"},
			Spec: shaderGraphNodeSpec{
				Name:        "Time",
				Description: "Runtime shader time values.",
				Outputs: []shaderGraphPortSpec{
					{Name: "Time", Type: "float"},
					{Name: "Sine", Type: "float"},
					{Name: "Cosine", Type: "float"},
				},
			},
		},
		{
			ID:          "world-position",
			Name:        "World Position",
			Description: "Fragment position in world space.",
			Tags:        []string{"context", "position", "world", "fragment", "space"},
			Spec: shaderGraphNodeSpec{
				Name:        "World Position",
				Description: "Fragment position in world space.",
				Outputs: []shaderGraphPortSpec{
					{Name: "Position", Type: "vec3"},
					{Name: "X", Type: "float"},
					{Name: "Y", Type: "float"},
					{Name: "Z", Type: "float"},
				},
			},
		},
		{
			ID:          "normal-vector",
			Name:        "Normal Vector",
			Description: "World-space geometric normal.",
			Tags:        []string{"context", "normal", "vector", "world", "surface"},
			Spec: shaderGraphNodeSpec{
				Name:        "Normal Vector",
				Description: "World-space geometric normal.",
				Outputs: []shaderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
					{Name: "X", Type: "float"},
					{Name: "Y", Type: "float"},
					{Name: "Z", Type: "float"},
				},
			},
		},
		{
			ID:          "tangent-vector",
			Name:        "Tangent Vector",
			Description: "Derived world-space tangent vector.",
			Tags:        []string{"context", "tangent", "vector", "world", "surface"},
			Spec: shaderGraphNodeSpec{
				Name:        "Tangent Vector",
				Description: "Derived world-space tangent vector.",
				Outputs: []shaderGraphPortSpec{
					{Name: "Tangent", Type: "vec3"},
					{Name: "X", Type: "float"},
					{Name: "Y", Type: "float"},
					{Name: "Z", Type: "float"},
				},
			},
		},
		{
			ID:          "bitangent-vector",
			Name:        "Bitangent Vector",
			Description: "Derived world-space bitangent vector.",
			Tags:        []string{"context", "bitangent", "binormal", "vector", "world", "surface"},
			Spec: shaderGraphNodeSpec{
				Name:        "Bitangent Vector",
				Description: "Derived world-space bitangent vector.",
				Outputs: []shaderGraphPortSpec{
					{Name: "Bitangent", Type: "vec3"},
					{Name: "X", Type: "float"},
					{Name: "Y", Type: "float"},
					{Name: "Z", Type: "float"},
				},
			},
		},
		{
			ID:          "view-direction",
			Name:        "View Direction",
			Description: "Direction from the fragment toward the camera.",
			Tags:        []string{"context", "view", "camera", "direction", "vector"},
			Spec: shaderGraphNodeSpec{
				Name:        "View Direction",
				Description: "Direction from the fragment toward the camera.",
				Outputs: []shaderGraphPortSpec{
					{Name: "Direction", Type: "vec3"},
					{Name: "X", Type: "float"},
					{Name: "Y", Type: "float"},
					{Name: "Z", Type: "float"},
				},
			},
		},
		{
			ID:          "camera-position",
			Name:        "Camera Position",
			Description: "Camera position in world space.",
			Tags:        []string{"context", "camera", "position", "world", "view"},
			Spec: shaderGraphNodeSpec{
				Name:        "Camera Position",
				Description: "Camera position in world space.",
				Outputs: []shaderGraphPortSpec{
					{Name: "Position", Type: "vec3"},
					{Name: "X", Type: "float"},
					{Name: "Y", Type: "float"},
					{Name: "Z", Type: "float"},
				},
			},
		},
		{
			ID:          "screen-position",
			Name:        "Screen Position",
			Description: "Fragment coordinates on the screen.",
			Tags:        []string{"context", "screen", "position", "fragment", "pixel", "depth"},
			Spec: shaderGraphNodeSpec{
				Name:        "Screen Position",
				Description: "Fragment coordinates on the screen.",
				Outputs: []shaderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Pixel", Type: "vec2"},
					{Name: "X", Type: "float"},
					{Name: "Y", Type: "float"},
					{Name: "Depth", Type: "float"},
				},
			},
		},
		{
			ID:          "vertex-color",
			Name:        "Vertex Color",
			Description: "Interpolated vertex and instance color.",
			Tags:        []string{"context", "vertex", "color", "instance", "tint"},
			Spec: shaderGraphNodeSpec{
				Name:        "Vertex Color",
				Description: "Interpolated vertex and instance color.",
				Outputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
					{Name: "RGB", Type: "vec3"},
					{Name: "Alpha", Type: "float"},
				},
			},
		},
		shaderGraphFloatBinaryNode("add", "Add", "Adds two float values.",
			[]string{"math", "float", "plus", "sum"}, "A", "B", "Value"),
		shaderGraphFloatBinaryNode("subtract", "Subtract", "Subtracts B from A.",
			[]string{"math", "float", "minus", "difference"}, "A", "B", "Value"),
		shaderGraphFloatBinaryNode("multiply", "Multiply", "Multiplies two float values.",
			[]string{"math", "float", "product"}, "A", "B", "Value"),
		shaderGraphFloatBinaryNode("divide", "Divide", "Divides A by B.",
			[]string{"math", "float", "quotient"}, "A", "B", "Value"),
		shaderGraphFloatBinaryNode("minimum", "Minimum", "Returns the smaller of two float values.",
			[]string{"math", "float", "min", "minimum"}, "A", "B", "Value"),
		shaderGraphFloatBinaryNode("maximum", "Maximum", "Returns the larger of two float values.",
			[]string{"math", "float", "max", "maximum"}, "A", "B", "Value"),
		shaderGraphFloatBinaryNode("power", "Power", "Raises Base to the Exponent.",
			[]string{"math", "float", "pow", "exponent"}, "Base", "Exponent", "Value"),
		shaderGraphFloatUnaryNode("absolute", "Absolute", "Returns the absolute value.",
			[]string{"math", "float", "abs"}, "Value"),
		shaderGraphFloatUnaryNode("one-minus", "One Minus", "Returns one minus the input value.",
			[]string{"math", "float", "invert", "complement"}, "Value"),
		shaderGraphFloatUnaryNode("floor", "Floor", "Rounds a float down to the nearest integer.",
			[]string{"math", "float", "round"}, "Value"),
		shaderGraphFloatUnaryNode("ceiling", "Ceiling", "Rounds a float up to the nearest integer.",
			[]string{"math", "float", "ceil", "round"}, "Value"),
		shaderGraphFloatUnaryNode("fraction", "Fraction", "Returns the fractional part of a float.",
			[]string{"math", "float", "frac", "fract"}, "Value"),
		shaderGraphFloatUnaryNode("sine", "Sine", "Returns the sine of the input angle.",
			[]string{"math", "float", "sin", "trig"}, "Angle"),
		shaderGraphFloatUnaryNode("cosine", "Cosine", "Returns the cosine of the input angle.",
			[]string{"math", "float", "cos", "trig"}, "Angle"),
		shaderGraphFloatUnaryNode("tangent", "Tangent", "Returns the tangent of the input angle.",
			[]string{"math", "float", "tan", "trig"}, "Angle"),
		shaderGraphFloatUnaryNode("square-root", "Square Root", "Returns the square root of a float.",
			[]string{"math", "float", "sqrt"}, "Value"),
		shaderGraphFloatTernaryNode("clamp", "Clamp", "Clamps a float between Min and Max.",
			[]string{"math", "float", "saturate", "limit"}, "Value", "Min", "Max", "Value"),
		shaderGraphFloatTernaryNode("lerp", "Lerp", "Linearly interpolates between A and B by T.",
			[]string{"math", "float", "mix", "interpolate"}, "A", "B", "T", "Value"),
		{
			ID:          "step",
			Name:        "Step",
			Description: "Returns 0 or 1 by comparing Value against Edge.",
			Tags:        []string{"math", "float", "threshold", "compare"},
			Spec: shaderGraphNodeSpec{
				Name:        "Step",
				Description: "Returns 0 or 1 by comparing Value against Edge.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Edge", Type: "float"},
					{Name: "Value", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Result", Type: "float"},
				},
			},
		},
		{
			ID:          "smoothstep",
			Name:        "Smoothstep",
			Description: "Smoothly interpolates from 0 to 1 between two edges.",
			Tags:        []string{"math", "float", "smooth", "threshold"},
			Spec: shaderGraphNodeSpec{
				Name:        "Smoothstep",
				Description: "Smoothly interpolates from 0 to 1 between two edges.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Edge Min", Type: "float"},
					{Name: "Edge Max", Type: "float"},
					{Name: "Value", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Result", Type: "float"},
				},
			},
		},
		{
			ID:          "dot-product",
			Name:        "Dot Product",
			Description: "Returns the scalar dot product of two vectors.",
			Tags:        []string{"math", "vector", "vec3", "dot"},
			Spec: shaderGraphNodeSpec{
				Name:        "Dot Product",
				Description: "Returns the scalar dot product of two vectors.",
				Inputs: []shaderGraphPortSpec{
					{Name: "A", Type: "vec3"},
					{Name: "B", Type: "vec3"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "cross-product",
			Name:        "Cross Product",
			Description: "Returns the perpendicular cross product of two vectors.",
			Tags:        []string{"math", "vector", "vec3", "cross"},
			Spec: shaderGraphNodeSpec{
				Name:        "Cross Product",
				Description: "Returns the perpendicular cross product of two vectors.",
				Inputs: []shaderGraphPortSpec{
					{Name: "A", Type: "vec3"},
					{Name: "B", Type: "vec3"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
			},
		},
		{
			ID:          "normalize",
			Name:        "Normalize",
			Description: "Returns a vector with the same direction and unit length.",
			Tags:        []string{"math", "vector", "vec3", "normal"},
			Spec: shaderGraphNodeSpec{
				Name:        "Normalize",
				Description: "Returns a vector with the same direction and unit length.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
			},
		},
		{
			ID:          "length",
			Name:        "Length",
			Description: "Returns the length of a vector.",
			Tags:        []string{"math", "vector", "vec3", "magnitude"},
			Spec: shaderGraphNodeSpec{
				Name:        "Length",
				Description: "Returns the length of a vector.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "mix-color",
			Name:        "Mix Color",
			Description: "Blends two colors with a factor.",
			Tags:        []string{"mix", "blend", "color", "factor"},
			Spec: shaderGraphNodeSpec{
				Name:        "Mix Color",
				Description: "Blends two colors with a factor.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:          "clamp",
						Label:       "Clamp",
						Type:        shaderGraphNodeFieldBool,
						DefaultBool: true,
					},
					{
						ID:      "mode",
						Label:   "Mode",
						Type:    shaderGraphNodeFieldSelect,
						Default: "mix",
						Options: []shaderGraphNodeFieldOption{
							{Label: "Mix", Value: "mix"},
							{Label: "Add", Value: "add"},
							{Label: "Multiply", Value: "multiply"},
						},
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Factor", Type: "float"},
					{Name: "A", Type: "color"},
					{Name: "B", Type: "color"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "principled-bsdf",
			Name:        "Principled BSDF",
			Description: "Surface shader with common material inputs.",
			Tags:        []string{"bsdf", "surface", "material", "shader"},
			Spec: shaderGraphNodeSpec{
				Name:        "Principled BSDF",
				Description: "Surface shader with common material inputs.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Base Color", Type: "color"},
					{Name: "Roughness", Type: "float"},
					{Name: "Normal", Type: "vec3"},
					{Name: "Metallic", Type: "float"},
					{Name: "Occlusion", Type: "float"},
					{Name: "Emission Color", Type: "color"},
					{Name: "Emission Strength", Type: "float"},
					{Name: "Alpha", Type: "float"},
					{Name: "Specular", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "BSDF", Type: "surface"},
				},
			},
		},
		{
			ID:          "material-output",
			Name:        "Material Output",
			Description: "Terminal output for the material shader.",
			Tags:        []string{"output", "surface", "volume", "material"},
			Spec: shaderGraphNodeSpec{
				Name:        "Material Output",
				Description: "Terminal output for the material shader.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Surface", Type: "surface"},
				},
			},
		},
	}
}

func shaderGraphFloatUnaryNode(id, name, description string, tags []string, input string) shaderGraphNodeCatalogEntry {
	return shaderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: shaderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs: []shaderGraphPortSpec{
				{Name: input, Type: "float"},
			},
			Outputs: []shaderGraphPortSpec{
				{Name: "Value", Type: "float"},
			},
		},
	}
}

func shaderGraphFloatBinaryNode(id, name, description string, tags []string, a, b, output string) shaderGraphNodeCatalogEntry {
	return shaderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: shaderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs: []shaderGraphPortSpec{
				{Name: a, Type: "float"},
				{Name: b, Type: "float"},
			},
			Outputs: []shaderGraphPortSpec{
				{Name: output, Type: "float"},
			},
		},
	}
}

func shaderGraphFloatTernaryNode(id, name, description string, tags []string, a, b, c, output string) shaderGraphNodeCatalogEntry {
	return shaderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: shaderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs: []shaderGraphPortSpec{
				{Name: a, Type: "float"},
				{Name: b, Type: "float"},
				{Name: c, Type: "float"},
			},
			Outputs: []shaderGraphPortSpec{
				{Name: output, Type: "float"},
			},
		},
	}
}

func shaderGraphNodeCatalogMenuData() []shaderGraphNodeMenuData {
	catalog := shaderGraphNodeCatalog()
	data := make([]shaderGraphNodeMenuData, 0, len(catalog))
	for i := range catalog {
		entry := catalog[i]
		search := strings.Join(append([]string{entry.ID, entry.Name, entry.Description}, entry.Tags...), " ")
		data = append(data, shaderGraphNodeMenuData{
			ID:          entry.ID,
			Name:        entry.Name,
			Description: entry.Description,
			Search:      strings.ToLower(search),
		})
	}
	return data
}

func shaderGraphNodeCatalogEntryCompatible(entry shaderGraphNodeCatalogEntry, compatibility shaderGraphNodePortCompatibility) bool {
	if !compatibility.Active {
		return true
	}
	_, ok := shaderGraphNodeSpecCompatiblePortIndex(entry.Spec, compatibility.SourceOutput, compatibility.Type)
	return ok
}

func shaderGraphNodeSpecCompatiblePortIndex(spec shaderGraphNodeSpec, sourceOutput bool, sourceType string) (int, bool) {
	ports := spec.Outputs
	if sourceOutput {
		ports = spec.Inputs
	}
	sourceType = shaderGraphPortTypeKey(sourceType)
	for i := range ports {
		if shaderGraphPortTypeKey(ports[i].Type) == sourceType {
			return i, true
		}
	}
	return -1, false
}

func shaderGraphNodeCatalogCompatibleIDs(sourceOutput bool, sourceType string) []string {
	catalog := shaderGraphNodeCatalog()
	out := make([]string, 0, len(catalog))
	compatibility := shaderGraphNodePortCompatibility{
		Active:       true,
		SourceOutput: sourceOutput,
		Type:         sourceType,
	}
	for i := range catalog {
		if shaderGraphNodeCatalogEntryCompatible(catalog[i], compatibility) {
			out = append(out, catalog[i].ID)
		}
	}
	return out
}

func shaderGraphNodeCatalogSpec(id string) (shaderGraphNodeSpec, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, entry := range shaderGraphNodeCatalog() {
		if entry.ID == id {
			return entry.Spec, true
		}
	}
	return shaderGraphNodeSpec{}, false
}
