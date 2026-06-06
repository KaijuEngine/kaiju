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
			ID:          "vector2",
			Name:        "Vector 2",
			Description: "Single vec2 value.",
			Tags:        []string{"vector", "vec2", "constant"},
			Spec: shaderGraphNodeSpec{
				Name:        "Vector 2",
				Description: "Single vec2 value.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:            "vector",
						Label:         "Vector",
						Type:          shaderGraphNodeFieldVector2,
						DefaultValues: []string{"0", "0"},
					},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Vector", Type: "vec2"},
				},
			},
		},
		{
			ID:          "vector4",
			Name:        "Vector 4",
			Description: "Single vec4 value.",
			Tags:        []string{"vector", "vec4", "constant", "color", "rgba"},
			Spec: shaderGraphNodeSpec{
				Name:        "Vector 4",
				Description: "Single vec4 value.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:            "vector",
						Label:         "Vector",
						Type:          shaderGraphNodeFieldVector4,
						DefaultValues: []string{"0", "0", "0", "0"},
					},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Vector", Type: "vec4"},
					{Name: "Color", Type: "color"},
				},
			},
		},
		shaderGraphCombineVectorNode("combine-vec2", "Combine Vec2", "Constructs a vec2 from scalar components.",
			[]string{"vector", "vec2", "compose", "combine", "construct"}, []string{"X", "Y"}, "vec2"),
		shaderGraphCombineVectorNode("combine-vec3", "Combine Vec3", "Constructs a vec3 from scalar components.",
			[]string{"vector", "vec3", "compose", "combine", "construct"}, []string{"X", "Y", "Z"}, "vec3"),
		shaderGraphCombineVectorNode("combine-vec4", "Combine Vec4", "Constructs a vec4 from scalar components.",
			[]string{"vector", "vec4", "compose", "combine", "construct", "rgba"}, []string{"X", "Y", "Z", "W"}, "vec4"),
		shaderGraphSplitVectorNode("split-vec2", "Split Vec2", "Splits a vec2 into scalar components.",
			[]string{"vector", "vec2", "split", "components", "xy"}, []string{"X", "Y"}, "vec2"),
		shaderGraphSplitVectorNode("split-vec3", "Split Vec3", "Splits a vec3 into scalar components.",
			[]string{"vector", "vec3", "split", "components", "xyz"}, []string{"X", "Y", "Z"}, "vec3"),
		shaderGraphSplitVectorNode("split-vec4", "Split Vec4", "Splits a vec4 into scalar components.",
			[]string{"vector", "vec4", "split", "components", "xyzw", "rgba"}, []string{"X", "Y", "Z", "W"}, "vec4"),
		shaderGraphSwizzleVectorNode("swizzle-vec2", "Swizzle Vec2", "Reorders vec2 components.",
			[]string{"vector", "vec2", "swizzle", "reorder", "xy"}, []string{"X", "Y"}, "vec2"),
		shaderGraphSwizzleVectorNode("swizzle-vec3", "Swizzle Vec3", "Reorders vec3 components.",
			[]string{"vector", "vec3", "swizzle", "reorder", "xyz"}, []string{"X", "Y", "Z"}, "vec3"),
		shaderGraphSwizzleVectorNode("swizzle-vec4", "Swizzle Vec4", "Reorders vec4/color components.",
			[]string{"vector", "vec4", "color", "swizzle", "reorder", "xyzw", "rgba"}, []string{"X", "Y", "Z", "W"}, "vec4"),
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
						Preview: true,
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
			ID:          "normal-map",
			Name:        "Normal Map",
			Description: "Unpacks a tangent-space normal map sample into a world-space normal.",
			Tags:        []string{"material", "texture", "normal", "map", "tangent", "pbr"},
			Spec: shaderGraphNodeSpec{
				Name:        "Normal Map",
				Description: "Unpacks a tangent-space normal map sample into a world-space normal.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "strength",
						Label:   "Strength",
						Type:    shaderGraphNodeFieldNumber,
						Default: "1.000",
					},
					{
						ID:      "y",
						Label:   "Y",
						Type:    shaderGraphNodeFieldSelect,
						Default: "opengl",
						Options: []shaderGraphNodeFieldOption{
							{Label: "OpenGL +Y", Value: "opengl"},
							{Label: "DirectX -Y", Value: "directx"},
						},
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "RGB", Type: "vec3"},
					{Name: "UV", Type: "vec2"},
					{Name: "Strength", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
					{Name: "Tangent", Type: "vec3"},
				},
			},
		},
		{
			ID:          "normal-strength",
			Name:        "Normal Strength",
			Description: "Adjusts a normal's influence relative to the geometric normal.",
			Tags:        []string{"material", "normal", "strength", "blend", "pbr"},
			Spec: shaderGraphNodeSpec{
				Name:        "Normal Strength",
				Description: "Adjusts a normal's influence relative to the geometric normal.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "strength",
						Label:   "Strength",
						Type:    shaderGraphNodeFieldNumber,
						Default: "1.000",
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
					{Name: "Strength", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
				},
			},
		},
		{
			ID:          "blend-normals",
			Name:        "Blend Normals",
			Description: "Layers a detail normal over a base normal.",
			Tags:        []string{"material", "normal", "blend", "detail", "pbr"},
			Spec: shaderGraphNodeSpec{
				Name:        "Blend Normals",
				Description: "Layers a detail normal over a base normal.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "strength",
						Label:   "Strength",
						Type:    shaderGraphNodeFieldNumber,
						Default: "1.000",
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Base", Type: "vec3"},
					{Name: "Detail", Type: "vec3"},
					{Name: "Strength", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
				},
			},
		},
		{
			ID:          "orm-mra-unpack",
			Name:        "ORM/MRA Unpack",
			Description: "Extracts occlusion, roughness, and metallic channels from a packed PBR map.",
			Tags:        []string{"material", "texture", "orm", "mra", "packed", "roughness", "metallic", "occlusion"},
			Spec: shaderGraphNodeSpec{
				Name:        "ORM/MRA Unpack",
				Description: "Extracts occlusion, roughness, and metallic channels from a packed PBR map.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "layout",
						Label:   "Layout",
						Type:    shaderGraphNodeFieldSelect,
						Default: "orm",
						Options: []shaderGraphNodeFieldOption{
							{Label: "ORM", Value: "orm"},
							{Label: "MRA", Value: "mra"},
						},
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Map", Type: "color"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Occlusion", Type: "float"},
					{Name: "Roughness", Type: "float"},
					{Name: "Metallic", Type: "float"},
				},
			},
		},
		{
			ID:          "height-bump",
			Name:        "Height/Bump",
			Description: "Derives a perturbed normal from a scalar height map.",
			Tags:        []string{"material", "texture", "height", "bump", "normal", "pbr"},
			Spec: shaderGraphNodeSpec{
				Name:        "Height/Bump",
				Description: "Derives a perturbed normal from a scalar height map.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "strength",
						Label:   "Strength",
						Type:    shaderGraphNodeFieldNumber,
						Default: "0.050",
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Height", Type: "float"},
					{Name: "Strength", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
				},
			},
		},
		{
			ID:          "parallax",
			Name:        "Parallax",
			Description: "Offsets UV coordinates using a height map and view direction.",
			Tags:        []string{"material", "texture", "height", "parallax", "uv", "pbr"},
			Spec: shaderGraphNodeSpec{
				Name:        "Parallax",
				Description: "Offsets UV coordinates using a height map and view direction.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "scale",
						Label:   "Scale",
						Type:    shaderGraphNodeFieldNumber,
						Default: "0.050",
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Height", Type: "float"},
					{Name: "Scale", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Offset", Type: "vec2"},
				},
			},
		},
		{
			ID:          "triplanar",
			Name:        "Triplanar",
			Description: "Samples a Texture 2D from world-space projection axes.",
			Tags:        []string{"material", "texture", "triplanar", "projection", "world", "pbr"},
			Spec: shaderGraphNodeSpec{
				Name:        "Triplanar",
				Description: "Samples a Texture 2D from world-space projection axes.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "scale",
						Label:   "Scale",
						Type:    shaderGraphNodeFieldNumber,
						Default: "1.000",
					},
					{
						ID:      "blend",
						Label:   "Blend",
						Type:    shaderGraphNodeFieldNumber,
						Default: "4.000",
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Texture", Type: "texture2D"},
					{Name: "Position", Type: "vec3"},
					{Name: "Normal", Type: "vec3"},
					{Name: "Scale", Type: "float"},
					{Name: "Blend", Type: "float"},
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
			ID:          "detail-texture",
			Name:        "Detail Texture",
			Description: "Blends a detail texture sample into a base color.",
			Tags:        []string{"material", "texture", "detail", "blend", "color", "pbr"},
			Spec: shaderGraphNodeSpec{
				Name:        "Detail Texture",
				Description: "Blends a detail texture sample into a base color.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "mode",
						Label:   "Mode",
						Type:    shaderGraphNodeFieldSelect,
						Default: "multiply",
						Options: []shaderGraphNodeFieldOption{
							{Label: "Multiply", Value: "multiply"},
							{Label: "Add", Value: "add"},
							{Label: "Overlay", Value: "overlay"},
						},
					},
					{
						ID:      "strength",
						Label:   "Strength",
						Type:    shaderGraphNodeFieldNumber,
						Default: "1.000",
					},
					{
						ID:          "clamp",
						Label:       "Clamp",
						Type:        shaderGraphNodeFieldBool,
						DefaultBool: true,
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Base", Type: "color"},
					{Name: "Detail", Type: "color"},
					{Name: "Mask", Type: "float"},
					{Name: "Strength", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
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
		{
			ID:          "noise",
			Name:        "Noise",
			Description: "Layered value noise for procedural masks and color variation.",
			Tags:        []string{"procedural", "noise", "fbm", "random", "mask", "texture"},
			Spec: shaderGraphNodeSpec{
				Name:        "Noise",
				Description: "Layered value noise for procedural masks and color variation.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "scale",
						Label:   "Scale",
						Type:    shaderGraphNodeFieldNumber,
						Default: "8.000",
					},
					{
						ID:      "detail",
						Label:   "Detail",
						Type:    shaderGraphNodeFieldNumber,
						Default: "4.000",
					},
					{
						ID:      "roughness",
						Label:   "Rough",
						Type:    shaderGraphNodeFieldNumber,
						Default: "0.500",
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Scale", Type: "float"},
					{Name: "Detail", Type: "float"},
					{Name: "Roughness", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "voronoi",
			Name:        "Voronoi",
			Description: "Cellular procedural pattern with distance, cell, and edge outputs.",
			Tags:        []string{"procedural", "voronoi", "cellular", "cells", "random", "mask"},
			Spec: shaderGraphNodeSpec{
				Name:        "Voronoi",
				Description: "Cellular procedural pattern with distance, cell, and edge outputs.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "scale",
						Label:   "Scale",
						Type:    shaderGraphNodeFieldNumber,
						Default: "8.000",
					},
					{
						ID:      "jitter",
						Label:   "Jitter",
						Type:    shaderGraphNodeFieldNumber,
						Default: "1.000",
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Scale", Type: "float"},
					{Name: "Jitter", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Distance", Type: "float"},
					{Name: "Cell", Type: "float"},
					{Name: "Edge", Type: "float"},
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "checker",
			Name:        "Checker",
			Description: "Procedural checkerboard pattern with mask and color outputs.",
			Tags:        []string{"procedural", "checker", "grid", "pattern", "mask", "texture"},
			Spec: shaderGraphNodeSpec{
				Name:        "Checker",
				Description: "Procedural checkerboard pattern with mask and color outputs.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "scale",
						Label:   "Scale",
						Type:    shaderGraphNodeFieldNumber,
						Default: "8.000",
					},
					{
						ID:           "color-a",
						Label:        "A",
						Type:         shaderGraphNodeFieldColor,
						DefaultColor: matrix.ColorWhite(),
					},
					{
						ID:           "color-b",
						Label:        "B",
						Type:         shaderGraphNodeFieldColor,
						DefaultColor: matrix.ColorBlack(),
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Scale", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
					{Name: "Mask", Type: "float"},
				},
			},
		},
		{
			ID:          "gradient",
			Name:        "Gradient",
			Description: "Linear or radial procedural gradient.",
			Tags:        []string{"procedural", "gradient", "ramp", "linear", "radial", "mask"},
			Spec: shaderGraphNodeSpec{
				Name:        "Gradient",
				Description: "Linear or radial procedural gradient.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "mode",
						Label:   "Mode",
						Type:    shaderGraphNodeFieldSelect,
						Default: "linear",
						Options: []shaderGraphNodeFieldOption{
							{Label: "Linear", Value: "linear"},
							{Label: "Radial", Value: "radial"},
						},
					},
					{
						ID:      "angle",
						Label:   "Angle",
						Type:    shaderGraphNodeFieldNumber,
						Default: "0.000",
					},
					{
						ID:           "color-a",
						Label:        "A",
						Type:         shaderGraphNodeFieldColor,
						DefaultColor: matrix.ColorBlack(),
					},
					{
						ID:           "color-b",
						Label:        "B",
						Type:         shaderGraphNodeFieldColor,
						DefaultColor: matrix.ColorWhite(),
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Angle", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
					{Name: "Factor", Type: "float"},
				},
			},
		},
		{
			ID:          "remap",
			Name:        "Remap",
			Description: "Maps a value from one range to another.",
			Tags:        []string{"procedural", "remap", "map", "range", "normalize", "mask"},
			Spec: shaderGraphNodeSpec{
				Name:        "Remap",
				Description: "Maps a value from one range to another.",
				Fields: []shaderGraphNodeFieldSpec{
					{ID: "in-min", Label: "In Min", Type: shaderGraphNodeFieldNumber, Default: "0.000"},
					{ID: "in-max", Label: "In Max", Type: shaderGraphNodeFieldNumber, Default: "1.000"},
					{ID: "out-min", Label: "Out Min", Type: shaderGraphNodeFieldNumber, Default: "0.000"},
					{ID: "out-max", Label: "Out Max", Type: shaderGraphNodeFieldNumber, Default: "1.000"},
					{ID: "clamp", Label: "Clamp", Type: shaderGraphNodeFieldBool, DefaultBool: false},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
					{Name: "In Min", Type: "float"},
					{Name: "In Max", Type: "float"},
					{Name: "Out Min", Type: "float"},
					{Name: "Out Max", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "posterize",
			Name:        "Posterize",
			Description: "Quantizes a scalar value into a limited number of steps.",
			Tags:        []string{"procedural", "posterize", "quantize", "steps", "toon", "mask"},
			Spec: shaderGraphNodeSpec{
				Name:        "Posterize",
				Description: "Quantizes a scalar value into a limited number of steps.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "steps",
						Label:   "Steps",
						Type:    shaderGraphNodeFieldNumber,
						Default: "4.000",
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
					{Name: "Steps", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "posterize-color",
			Name:        "Posterize Color",
			Description: "Quantizes each color channel into a limited number of steps.",
			Tags:        []string{"procedural", "posterize", "color", "quantize", "steps", "toon"},
			Spec: shaderGraphNodeSpec{
				Name:        "Posterize Color",
				Description: "Quantizes each color channel into a limited number of steps.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "steps",
						Label:   "Steps",
						Type:    shaderGraphNodeFieldNumber,
						Default: "4.000",
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
					{Name: "Steps", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "fresnel",
			Name:        "Fresnel",
			Description: "View-angle mask for edge highlights and falloff effects.",
			Tags:        []string{"procedural", "fresnel", "rim", "view", "normal", "falloff"},
			Spec: shaderGraphNodeSpec{
				Name:        "Fresnel",
				Description: "View-angle mask for edge highlights and falloff effects.",
				Fields: []shaderGraphNodeFieldSpec{
					{ID: "power", Label: "Power", Type: shaderGraphNodeFieldNumber, Default: "5.000"},
					{ID: "bias", Label: "Bias", Type: shaderGraphNodeFieldNumber, Default: "0.000"},
					{ID: "scale", Label: "Scale", Type: shaderGraphNodeFieldNumber, Default: "1.000"},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
					{Name: "View", Type: "vec3"},
					{Name: "Power", Type: "float"},
					{Name: "Bias", Type: "float"},
					{Name: "Scale", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Factor", Type: "float"},
				},
			},
		},
		{
			ID:          "rim-light",
			Name:        "Rim Light",
			Description: "Generates a colored rim-light mask from normal and view direction.",
			Tags:        []string{"procedural", "rim", "light", "fresnel", "edge", "view"},
			Spec: shaderGraphNodeSpec{
				Name:        "Rim Light",
				Description: "Generates a colored rim-light mask from normal and view direction.",
				Fields: []shaderGraphNodeFieldSpec{
					{ID: "power", Label: "Power", Type: shaderGraphNodeFieldNumber, Default: "3.000"},
					{ID: "intensity", Label: "Intens", Type: shaderGraphNodeFieldNumber, Default: "1.000"},
					{ID: "color", Label: "Color", Type: shaderGraphNodeFieldColor, DefaultColor: matrix.ColorWhite()},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
					{Name: "View", Type: "vec3"},
					{Name: "Power", Type: "float"},
					{Name: "Intensity", Type: "float"},
					{Name: "Color", Type: "color"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Factor", Type: "float"},
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "fwidth",
			Name:        "FWidth",
			Description: "Returns the approximate screen-space width of a scalar expression.",
			Tags:        []string{"procedural", "derivative", "fwidth", "antialias", "screen"},
			Spec: shaderGraphNodeSpec{
				Name:        "FWidth",
				Description: "Returns the approximate screen-space width of a scalar expression.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "ddx",
			Name:        "DDX",
			Description: "Returns the screen-space derivative of a scalar value along X.",
			Tags:        []string{"procedural", "derivative", "ddx", "dfdx", "screen"},
			Spec: shaderGraphNodeSpec{
				Name:        "DDX",
				Description: "Returns the screen-space derivative of a scalar value along X.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "ddy",
			Name:        "DDY",
			Description: "Returns the screen-space derivative of a scalar value along Y.",
			Tags:        []string{"procedural", "derivative", "ddy", "dfdy", "screen"},
			Spec: shaderGraphNodeSpec{
				Name:        "DDY",
				Description: "Returns the screen-space derivative of a scalar value along Y.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
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
		shaderGraphVectorBinaryNode("add-vec2", "Add Vec2", "Adds two vec2 values component-wise.",
			[]string{"math", "vector", "vec2", "add", "plus", "sum"}, "vec2"),
		shaderGraphVectorBinaryNode("subtract-vec2", "Subtract Vec2", "Subtracts B from A component-wise.",
			[]string{"math", "vector", "vec2", "subtract", "minus", "difference"}, "vec2"),
		shaderGraphVectorBinaryNode("multiply-vec2", "Multiply Vec2", "Multiplies two vec2 values component-wise.",
			[]string{"math", "vector", "vec2", "multiply", "product"}, "vec2"),
		shaderGraphVectorBinaryNode("divide-vec2", "Divide Vec2", "Divides A by B component-wise.",
			[]string{"math", "vector", "vec2", "divide", "quotient"}, "vec2"),
		shaderGraphVectorBinaryNode("add-vec3", "Add Vec3", "Adds two vec3 values component-wise.",
			[]string{"math", "vector", "vec3", "add", "plus", "sum"}, "vec3"),
		shaderGraphVectorBinaryNode("subtract-vec3", "Subtract Vec3", "Subtracts B from A component-wise.",
			[]string{"math", "vector", "vec3", "subtract", "minus", "difference"}, "vec3"),
		shaderGraphVectorBinaryNode("multiply-vec3", "Multiply Vec3", "Multiplies two vec3 values component-wise.",
			[]string{"math", "vector", "vec3", "multiply", "product"}, "vec3"),
		shaderGraphVectorBinaryNode("divide-vec3", "Divide Vec3", "Divides A by B component-wise.",
			[]string{"math", "vector", "vec3", "divide", "quotient"}, "vec3"),
		shaderGraphVectorBinaryNode("add-vec4", "Add Vec4", "Adds two vec4 values component-wise.",
			[]string{"math", "vector", "vec4", "add", "plus", "sum", "color"}, "vec4"),
		shaderGraphVectorBinaryNode("subtract-vec4", "Subtract Vec4", "Subtracts B from A component-wise.",
			[]string{"math", "vector", "vec4", "subtract", "minus", "difference", "color"}, "vec4"),
		shaderGraphVectorBinaryNode("multiply-vec4", "Multiply Vec4", "Multiplies two vec4 values component-wise.",
			[]string{"math", "vector", "vec4", "multiply", "product", "color"}, "vec4"),
		shaderGraphVectorBinaryNode("divide-vec4", "Divide Vec4", "Divides A by B component-wise.",
			[]string{"math", "vector", "vec4", "divide", "quotient", "color"}, "vec4"),
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
					{Name: "Displacement", Type: "float"},
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

func shaderGraphVectorBinaryNode(id, name, description string, tags []string, vectorType string) shaderGraphNodeCatalogEntry {
	outputs := []shaderGraphPortSpec{{Name: "Vector", Type: vectorType}}
	if vectorType == "vec4" {
		outputs = append(outputs, shaderGraphPortSpec{Name: "Color", Type: "color"})
	}
	return shaderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: shaderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs: []shaderGraphPortSpec{
				{Name: "A", Type: vectorType},
				{Name: "B", Type: vectorType},
			},
			Outputs: outputs,
		},
	}
}

func shaderGraphCombineVectorNode(id, name, description string, tags, components []string, outputType string) shaderGraphNodeCatalogEntry {
	inputs := make([]shaderGraphPortSpec, len(components))
	for i := range components {
		inputs[i] = shaderGraphPortSpec{Name: components[i], Type: "float"}
	}
	outputs := []shaderGraphPortSpec{{Name: "Vector", Type: outputType}}
	if outputType == "vec4" {
		outputs = append(outputs, shaderGraphPortSpec{Name: "Color", Type: "color"})
	}
	return shaderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: shaderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs:      inputs,
			Outputs:     outputs,
		},
	}
}

func shaderGraphSplitVectorNode(id, name, description string, tags, components []string, inputType string) shaderGraphNodeCatalogEntry {
	outputs := make([]shaderGraphPortSpec, len(components))
	for i := range components {
		outputs[i] = shaderGraphPortSpec{Name: components[i], Type: "float"}
	}
	return shaderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: shaderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs: []shaderGraphPortSpec{
				{Name: "Vector", Type: inputType},
			},
			Outputs: outputs,
		},
	}
}

func shaderGraphSwizzleVectorNode(id, name, description string, tags, components []string, vectorType string) shaderGraphNodeCatalogEntry {
	fields := make([]shaderGraphNodeFieldSpec, len(components))
	options := shaderGraphSwizzleFieldOptions(components)
	for i := range components {
		fields[i] = shaderGraphNodeFieldSpec{
			ID:      strings.ToLower(components[i]),
			Label:   components[i],
			Type:    shaderGraphNodeFieldSelect,
			Default: strings.ToLower(components[i]),
			Options: options,
		}
	}
	outputs := []shaderGraphPortSpec{{Name: "Vector", Type: vectorType}}
	if vectorType == "vec4" {
		outputs = append(outputs, shaderGraphPortSpec{Name: "Color", Type: "color"})
	}
	return shaderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: shaderGraphNodeSpec{
			Name:        name,
			Description: description,
			Fields:      fields,
			Inputs: []shaderGraphPortSpec{
				{Name: "Vector", Type: vectorType},
			},
			Outputs: outputs,
		},
	}
}

func shaderGraphSwizzleFieldOptions(components []string) []shaderGraphNodeFieldOption {
	options := make([]shaderGraphNodeFieldOption, 0, len(components)+2)
	for i := range components {
		component := strings.ToLower(components[i])
		label := components[i]
		if len(components) == 4 {
			label += " / " + []string{"R", "G", "B", "A"}[i]
		}
		options = append(options, shaderGraphNodeFieldOption{Label: label, Value: component})
	}
	options = append(options,
		shaderGraphNodeFieldOption{Label: "0", Value: "0"},
		shaderGraphNodeFieldOption{Label: "1", Value: "1"},
	)
	return options
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
