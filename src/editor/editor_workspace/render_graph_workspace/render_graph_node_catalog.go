/******************************************************************************/
/* render_graph_node_catalog.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"strings"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
)

type renderGraphNodeCatalogEntry struct {
	ID          string
	Name        string
	Description string
	Tags        []string
	Spec        renderGraphNodeSpec
}

type renderGraphNodeMenuData struct {
	ID          string
	Name        string
	Description string
	Search      string
}

type renderGraphNodePortCompatibility struct {
	Active       bool
	SourceOutput bool
	Type         string
}

func renderGraphNodeCatalog() []renderGraphNodeCatalogEntry {
	return []renderGraphNodeCatalogEntry{
		{
			ID:          "value",
			Name:        "Value",
			Description: "Single float value.",
			Tags:        []string{"float", "number", "constant"},
			Spec: renderGraphNodeSpec{
				Name:        "Value",
				Description: "Single float value.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "value",
						Label:   "Value",
						Type:    renderGraphNodeFieldNumber,
						Default: "1.000",
					},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "color",
			Name:        "Color",
			Description: "Single color value.",
			Tags:        []string{"color", "constant", "albedo"},
			Spec: renderGraphNodeSpec{
				Name:        "Color",
				Description: "Single color value.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:           "color",
						Label:        "Color",
						Type:         renderGraphNodeFieldColor,
						DefaultColor: matrix.ColorWhite(),
					},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "vector",
			Name:        "Vector",
			Description: "Single vec3 value.",
			Tags:        []string{"vector", "vec3", "constant"},
			Spec: renderGraphNodeSpec{
				Name:        "Vector",
				Description: "Single vec3 value.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:            "vector",
						Label:         "Vector",
						Type:          renderGraphNodeFieldVector3,
						DefaultValues: []string{"0", "0", "0"},
					},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
			},
		},
		{
			ID:          "vector2",
			Name:        "Vector 2",
			Description: "Single vec2 value.",
			Tags:        []string{"vector", "vec2", "constant"},
			Spec: renderGraphNodeSpec{
				Name:        "Vector 2",
				Description: "Single vec2 value.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:            "vector",
						Label:         "Vector",
						Type:          renderGraphNodeFieldVector2,
						DefaultValues: []string{"0", "0"},
					},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Vector", Type: "vec2"},
				},
			},
		},
		{
			ID:          "vector4",
			Name:        "Vector 4",
			Description: "Single vec4 value.",
			Tags:        []string{"vector", "vec4", "constant", "color", "rgba"},
			Spec: renderGraphNodeSpec{
				Name:        "Vector 4",
				Description: "Single vec4 value.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:            "vector",
						Label:         "Vector",
						Type:          renderGraphNodeFieldVector4,
						DefaultValues: []string{"0", "0", "0", "0"},
					},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Vector", Type: "vec4"},
					{Name: "Color", Type: "color"},
				},
			},
		},
		renderGraphCombineVectorNode("combine-vec2", "Combine Vec2", "Constructs a vec2 from scalar components.",
			[]string{"vector", "vec2", "compose", "combine", "construct"}, []string{"X", "Y"}, "vec2"),
		renderGraphCombineVectorNode("combine-vec3", "Combine Vec3", "Constructs a vec3 from scalar components.",
			[]string{"vector", "vec3", "compose", "combine", "construct"}, []string{"X", "Y", "Z"}, "vec3"),
		renderGraphCombineVectorNode("combine-vec4", "Combine Vec4", "Constructs a vec4 from scalar components.",
			[]string{"vector", "vec4", "compose", "combine", "construct", "rgba"}, []string{"X", "Y", "Z", "W"}, "vec4"),
		renderGraphSplitVectorNode("split-vec2", "Split Vec2", "Splits a vec2 into scalar components.",
			[]string{"vector", "vec2", "split", "components", "xy"}, []string{"X", "Y"}, "vec2"),
		renderGraphSplitVectorNode("split-vec3", "Split Vec3", "Splits a vec3 into scalar components.",
			[]string{"vector", "vec3", "split", "components", "xyz"}, []string{"X", "Y", "Z"}, "vec3"),
		renderGraphSplitVectorNode("split-vec4", "Split Vec4", "Splits a vec4 into scalar components.",
			[]string{"vector", "vec4", "split", "components", "xyzw", "rgba"}, []string{"X", "Y", "Z", "W"}, "vec4"),
		renderGraphSwizzleVectorNode("swizzle-vec2", "Swizzle Vec2", "Reorders vec2 components.",
			[]string{"vector", "vec2", "swizzle", "reorder", "xy"}, []string{"X", "Y"}, "vec2"),
		renderGraphSwizzleVectorNode("swizzle-vec3", "Swizzle Vec3", "Reorders vec3 components.",
			[]string{"vector", "vec3", "swizzle", "reorder", "xyz"}, []string{"X", "Y", "Z"}, "vec3"),
		renderGraphSwizzleVectorNode("swizzle-vec4", "Swizzle Vec4", "Reorders vec4/color components.",
			[]string{"vector", "vec4", "color", "swizzle", "reorder", "xyzw", "rgba"}, []string{"X", "Y", "Z", "W"}, "vec4"),
		{
			ID:          "texture-2d",
			Name:        "Texture 2D",
			Description: "Texture asset used by texture sample nodes.",
			Tags:        []string{"texture", "image", "sampler", "asset"},
			Spec: renderGraphNodeSpec{
				Name:        "Texture 2D",
				Description: "Texture asset used by texture sample nodes.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "texture",
						Label:   "Texture",
						Type:    renderGraphNodeFieldTexture,
						Default: assets.TextureSquare,
						Preview: true,
					},
					{
						ID:      "label",
						Label:   "Label",
						Type:    renderGraphNodeFieldText,
						Default: "Texture",
					},
					{
						ID:      "filter",
						Label:   "Filter",
						Type:    renderGraphNodeFieldSelect,
						Default: "Linear",
						Options: []renderGraphNodeFieldOption{
							{Label: "Linear", Value: "Linear"},
							{Label: "Nearest", Value: "Nearest"},
						},
					},
					{
						ID:      "color-space",
						Label:   "Space",
						Type:    renderGraphNodeFieldSelect,
						Default: "srgb",
						Options: []renderGraphNodeFieldOption{
							{Label: "sRGB", Value: "srgb"},
							{Label: "Linear", Value: "linear"},
						},
					},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Texture", Type: "texture2D"},
				},
			},
		},
		{
			ID:          "sample-texture-2d",
			Name:        "Sample Texture 2D",
			Description: "Samples a Texture 2D at UV coordinates.",
			Tags:        []string{"texture", "sample", "image", "sampler", "uv"},
			Spec: renderGraphNodeSpec{
				Name:        "Sample Texture 2D",
				Description: "Samples a Texture 2D at UV coordinates.",
				Inputs: []renderGraphPortSpec{
					{Name: "Texture", Type: "texture2D"},
					{Name: "UV", Type: "vec2"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "UV",
				Description: "Primary mesh UV coordinates.",
				Outputs: []renderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
				},
			},
		},
		{
			ID:          "uv-transform",
			Name:        "UV Transform",
			Description: "Applies tiling and offset to UV coordinates.",
			Tags:        []string{"texture", "uv", "tiling", "offset", "coordinates"},
			Spec: renderGraphNodeSpec{
				Name:        "UV Transform",
				Description: "Applies tiling and offset to UV coordinates.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:            "tiling",
						Label:         "Tiling",
						Type:          renderGraphNodeFieldVector2,
						DefaultValues: []string{"1", "1"},
					},
					{
						ID:            "offset",
						Label:         "Offset",
						Type:          renderGraphNodeFieldVector2,
						DefaultValues: []string{"0", "0"},
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
				},
			},
		},
		{
			ID:          "split-rgba",
			Name:        "Split RGBA",
			Description: "Splits a color into scalar channels.",
			Tags:        []string{"texture", "color", "channel", "rgba", "split"},
			Spec: renderGraphNodeSpec{
				Name:        "Split RGBA",
				Description: "Splits a color into scalar channels.",
				Inputs: []renderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Channel Mask",
				Description: "Extracts one scalar channel from a texture or color sample.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "channel",
						Label:   "Channel",
						Type:    renderGraphNodeFieldSelect,
						Default: "r",
						Options: []renderGraphNodeFieldOption{
							{Label: "R", Value: "r"},
							{Label: "G", Value: "g"},
							{Label: "B", Value: "b"},
							{Label: "A", Value: "a"},
							{Label: "Luma", Value: "luma"},
						},
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "texel-size",
			Name:        "Texel Size",
			Description: "Returns inverse texture dimensions for a Texture 2D.",
			Tags:        []string{"texture", "texel", "size", "dimensions", "pixel"},
			Spec: renderGraphNodeSpec{
				Name:        "Texel Size",
				Description: "Returns inverse texture dimensions for a Texture 2D.",
				Inputs: []renderGraphPortSpec{
					{Name: "Texture", Type: "texture2D"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Normal Map",
				Description: "Unpacks a tangent-space normal map sample into a world-space normal.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "strength",
						Label:   "Strength",
						Type:    renderGraphNodeFieldNumber,
						Default: "1.000",
					},
					{
						ID:      "y",
						Label:   "Y",
						Type:    renderGraphNodeFieldSelect,
						Default: "opengl",
						Options: []renderGraphNodeFieldOption{
							{Label: "OpenGL +Y", Value: "opengl"},
							{Label: "DirectX -Y", Value: "directx"},
						},
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "RGB", Type: "vec3"},
					{Name: "UV", Type: "vec2"},
					{Name: "Strength", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Normal Strength",
				Description: "Adjusts a normal's influence relative to the geometric normal.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "strength",
						Label:   "Strength",
						Type:    renderGraphNodeFieldNumber,
						Default: "1.000",
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
					{Name: "Strength", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
				},
			},
		},
		{
			ID:          "blend-normals",
			Name:        "Blend Normals",
			Description: "Layers a detail normal over a base normal.",
			Tags:        []string{"material", "normal", "blend", "detail", "pbr"},
			Spec: renderGraphNodeSpec{
				Name:        "Blend Normals",
				Description: "Layers a detail normal over a base normal.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "strength",
						Label:   "Strength",
						Type:    renderGraphNodeFieldNumber,
						Default: "1.000",
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Base", Type: "vec3"},
					{Name: "Detail", Type: "vec3"},
					{Name: "Strength", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
				},
			},
		},
		{
			ID:          "orm-mra-unpack",
			Name:        "ORM/MRA Unpack",
			Description: "Extracts occlusion, roughness, and metallic channels from a packed PBR map.",
			Tags:        []string{"material", "texture", "orm", "mra", "packed", "roughness", "metallic", "occlusion"},
			Spec: renderGraphNodeSpec{
				Name:        "ORM/MRA Unpack",
				Description: "Extracts occlusion, roughness, and metallic channels from a packed PBR map.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "layout",
						Label:   "Layout",
						Type:    renderGraphNodeFieldSelect,
						Default: "orm",
						Options: []renderGraphNodeFieldOption{
							{Label: "ORM", Value: "orm"},
							{Label: "MRA", Value: "mra"},
						},
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Map", Type: "color"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Height/Bump",
				Description: "Derives a perturbed normal from a scalar height map.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "strength",
						Label:   "Strength",
						Type:    renderGraphNodeFieldNumber,
						Default: "0.050",
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Height", Type: "float"},
					{Name: "Strength", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
				},
			},
		},
		{
			ID:          "parallax",
			Name:        "Parallax",
			Description: "Offsets UV coordinates using a height map and view direction.",
			Tags:        []string{"material", "texture", "height", "parallax", "uv", "pbr"},
			Spec: renderGraphNodeSpec{
				Name:        "Parallax",
				Description: "Offsets UV coordinates using a height map and view direction.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "scale",
						Label:   "Scale",
						Type:    renderGraphNodeFieldNumber,
						Default: "0.050",
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Height", Type: "float"},
					{Name: "Scale", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Triplanar",
				Description: "Samples a Texture 2D from world-space projection axes.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "scale",
						Label:   "Scale",
						Type:    renderGraphNodeFieldNumber,
						Default: "1.000",
					},
					{
						ID:      "blend",
						Label:   "Blend",
						Type:    renderGraphNodeFieldNumber,
						Default: "4.000",
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Texture", Type: "texture2D"},
					{Name: "Position", Type: "vec3"},
					{Name: "Normal", Type: "vec3"},
					{Name: "Scale", Type: "float"},
					{Name: "Blend", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Detail Texture",
				Description: "Blends a detail texture sample into a base color.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "mode",
						Label:   "Mode",
						Type:    renderGraphNodeFieldSelect,
						Default: "multiply",
						Options: []renderGraphNodeFieldOption{
							{Label: "Multiply", Value: "multiply"},
							{Label: "Add", Value: "add"},
							{Label: "Overlay", Value: "overlay"},
						},
					},
					{
						ID:      "strength",
						Label:   "Strength",
						Type:    renderGraphNodeFieldNumber,
						Default: "1.000",
					},
					{
						ID:          "clamp",
						Label:       "Clamp",
						Type:        renderGraphNodeFieldBool,
						DefaultBool: true,
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Base", Type: "color"},
					{Name: "Detail", Type: "color"},
					{Name: "Mask", Type: "float"},
					{Name: "Strength", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "time",
			Name:        "Time",
			Description: "Runtime shader time values.",
			Tags:        []string{"context", "time", "animation", "runtime", "clock"},
			Spec: renderGraphNodeSpec{
				Name:        "Time",
				Description: "Runtime shader time values.",
				Outputs: []renderGraphPortSpec{
					{Name: "Time", Type: "float"},
				},
			},
		},
		{
			ID:          "world-position",
			Name:        "World Position",
			Description: "Fragment position in world space.",
			Tags:        []string{"context", "position", "world", "fragment", "space"},
			Spec: renderGraphNodeSpec{
				Name:        "World Position",
				Description: "Fragment position in world space.",
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Normal Vector",
				Description: "World-space geometric normal.",
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Tangent Vector",
				Description: "Derived world-space tangent vector.",
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Bitangent Vector",
				Description: "Derived world-space bitangent vector.",
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "View Direction",
				Description: "Direction from the fragment toward the camera.",
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Camera Position",
				Description: "Camera position in world space.",
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Screen Position",
				Description: "Fragment coordinates on the screen.",
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Vertex Color",
				Description: "Interpolated vertex and instance color.",
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Noise",
				Description: "Layered value noise for procedural masks and color variation.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "scale",
						Label:   "Scale",
						Type:    renderGraphNodeFieldNumber,
						Default: "8.000",
					},
					{
						ID:      "detail",
						Label:   "Detail",
						Type:    renderGraphNodeFieldNumber,
						Default: "4.000",
					},
					{
						ID:      "roughness",
						Label:   "Rough",
						Type:    renderGraphNodeFieldNumber,
						Default: "0.500",
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Scale", Type: "float"},
					{Name: "Detail", Type: "float"},
					{Name: "Roughness", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Voronoi",
				Description: "Cellular procedural pattern with distance, cell, and edge outputs.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "scale",
						Label:   "Scale",
						Type:    renderGraphNodeFieldNumber,
						Default: "8.000",
					},
					{
						ID:      "jitter",
						Label:   "Jitter",
						Type:    renderGraphNodeFieldNumber,
						Default: "1.000",
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Scale", Type: "float"},
					{Name: "Jitter", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Checker",
				Description: "Procedural checkerboard pattern with mask and color outputs.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "scale",
						Label:   "Scale",
						Type:    renderGraphNodeFieldNumber,
						Default: "8.000",
					},
					{
						ID:           "color-a",
						Label:        "A",
						Type:         renderGraphNodeFieldColor,
						DefaultColor: matrix.ColorWhite(),
					},
					{
						ID:           "color-b",
						Label:        "B",
						Type:         renderGraphNodeFieldColor,
						DefaultColor: matrix.ColorBlack(),
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Scale", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Gradient",
				Description: "Linear or radial procedural gradient.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "mode",
						Label:   "Mode",
						Type:    renderGraphNodeFieldSelect,
						Default: "linear",
						Options: []renderGraphNodeFieldOption{
							{Label: "Linear", Value: "linear"},
							{Label: "Radial", Value: "radial"},
						},
					},
					{
						ID:      "angle",
						Label:   "Angle",
						Type:    renderGraphNodeFieldNumber,
						Default: "0.000",
					},
					{
						ID:           "color-a",
						Label:        "A",
						Type:         renderGraphNodeFieldColor,
						DefaultColor: matrix.ColorBlack(),
					},
					{
						ID:           "color-b",
						Label:        "B",
						Type:         renderGraphNodeFieldColor,
						DefaultColor: matrix.ColorWhite(),
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "UV", Type: "vec2"},
					{Name: "Angle", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "Remap",
				Description: "Maps a value from one range to another.",
				Fields: []renderGraphNodeFieldSpec{
					{ID: "in-min", Label: "In Min", Type: renderGraphNodeFieldNumber, Default: "0.000"},
					{ID: "in-max", Label: "In Max", Type: renderGraphNodeFieldNumber, Default: "1.000"},
					{ID: "out-min", Label: "Out Min", Type: renderGraphNodeFieldNumber, Default: "0.000"},
					{ID: "out-max", Label: "Out Max", Type: renderGraphNodeFieldNumber, Default: "1.000"},
					{ID: "clamp", Label: "Clamp", Type: renderGraphNodeFieldBool, DefaultBool: false},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
					{Name: "In Min", Type: "float"},
					{Name: "In Max", Type: "float"},
					{Name: "Out Min", Type: "float"},
					{Name: "Out Max", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "posterize",
			Name:        "Posterize",
			Description: "Quantizes a scalar value into a limited number of steps.",
			Tags:        []string{"procedural", "posterize", "quantize", "steps", "toon", "mask"},
			Spec: renderGraphNodeSpec{
				Name:        "Posterize",
				Description: "Quantizes a scalar value into a limited number of steps.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "steps",
						Label:   "Steps",
						Type:    renderGraphNodeFieldNumber,
						Default: "4.000",
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
					{Name: "Steps", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "posterize-color",
			Name:        "Posterize Color",
			Description: "Quantizes each color channel into a limited number of steps.",
			Tags:        []string{"procedural", "posterize", "color", "quantize", "steps", "toon"},
			Spec: renderGraphNodeSpec{
				Name:        "Posterize Color",
				Description: "Quantizes each color channel into a limited number of steps.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:      "steps",
						Label:   "Steps",
						Type:    renderGraphNodeFieldNumber,
						Default: "4.000",
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Color", Type: "color"},
					{Name: "Steps", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "fresnel",
			Name:        "Fresnel",
			Description: "View-angle mask for edge highlights and falloff effects.",
			Tags:        []string{"procedural", "fresnel", "rim", "view", "normal", "falloff"},
			Spec: renderGraphNodeSpec{
				Name:        "Fresnel",
				Description: "View-angle mask for edge highlights and falloff effects.",
				Fields: []renderGraphNodeFieldSpec{
					{ID: "power", Label: "Power", Type: renderGraphNodeFieldNumber, Default: "5.000"},
					{ID: "bias", Label: "Bias", Type: renderGraphNodeFieldNumber, Default: "0.000"},
					{ID: "scale", Label: "Scale", Type: renderGraphNodeFieldNumber, Default: "1.000"},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
					{Name: "View", Type: "vec3"},
					{Name: "Power", Type: "float"},
					{Name: "Bias", Type: "float"},
					{Name: "Scale", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Factor", Type: "float"},
				},
			},
		},
		{
			ID:          "rim-light",
			Name:        "Rim Light",
			Description: "Generates a colored rim-light mask from normal and view direction.",
			Tags:        []string{"procedural", "rim", "light", "fresnel", "edge", "view"},
			Spec: renderGraphNodeSpec{
				Name:        "Rim Light",
				Description: "Generates a colored rim-light mask from normal and view direction.",
				Fields: []renderGraphNodeFieldSpec{
					{ID: "power", Label: "Power", Type: renderGraphNodeFieldNumber, Default: "3.000"},
					{ID: "intensity", Label: "Intens", Type: renderGraphNodeFieldNumber, Default: "1.000"},
					{ID: "color", Label: "Color", Type: renderGraphNodeFieldColor, DefaultColor: matrix.ColorWhite()},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Normal", Type: "vec3"},
					{Name: "View", Type: "vec3"},
					{Name: "Power", Type: "float"},
					{Name: "Intensity", Type: "float"},
					{Name: "Color", Type: "color"},
				},
				Outputs: []renderGraphPortSpec{
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
			Spec: renderGraphNodeSpec{
				Name:        "FWidth",
				Description: "Returns the approximate screen-space width of a scalar expression.",
				Inputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "ddx",
			Name:        "DDX",
			Description: "Returns the screen-space derivative of a scalar value along X.",
			Tags:        []string{"procedural", "derivative", "ddx", "dfdx", "screen"},
			Spec: renderGraphNodeSpec{
				Name:        "DDX",
				Description: "Returns the screen-space derivative of a scalar value along X.",
				Inputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "ddy",
			Name:        "DDY",
			Description: "Returns the screen-space derivative of a scalar value along Y.",
			Tags:        []string{"procedural", "derivative", "ddy", "dfdy", "screen"},
			Spec: renderGraphNodeSpec{
				Name:        "DDY",
				Description: "Returns the screen-space derivative of a scalar value along Y.",
				Inputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		renderGraphFloatBinaryNode("add", "Add", "Adds two float values.",
			[]string{"math", "float", "plus", "sum"}, "A", "B", "Value"),
		renderGraphFloatBinaryNode("subtract", "Subtract", "Subtracts B from A.",
			[]string{"math", "float", "minus", "difference"}, "A", "B", "Value"),
		renderGraphFloatBinaryNode("multiply", "Multiply", "Multiplies two float values.",
			[]string{"math", "float", "product"}, "A", "B", "Value"),
		renderGraphFloatBinaryNode("divide", "Divide", "Divides A by B.",
			[]string{"math", "float", "quotient"}, "A", "B", "Value"),
		renderGraphVectorBinaryNode("add-vec2", "Add Vec2", "Adds two vec2 values component-wise.",
			[]string{"math", "vector", "vec2", "add", "plus", "sum"}, "vec2"),
		renderGraphVectorBinaryNode("subtract-vec2", "Subtract Vec2", "Subtracts B from A component-wise.",
			[]string{"math", "vector", "vec2", "subtract", "minus", "difference"}, "vec2"),
		renderGraphVectorBinaryNode("multiply-vec2", "Multiply Vec2", "Multiplies two vec2 values component-wise.",
			[]string{"math", "vector", "vec2", "multiply", "product"}, "vec2"),
		renderGraphVectorBinaryNode("divide-vec2", "Divide Vec2", "Divides A by B component-wise.",
			[]string{"math", "vector", "vec2", "divide", "quotient"}, "vec2"),
		renderGraphVectorBinaryNode("add-vec3", "Add Vec3", "Adds two vec3 values component-wise.",
			[]string{"math", "vector", "vec3", "add", "plus", "sum"}, "vec3"),
		renderGraphVectorBinaryNode("subtract-vec3", "Subtract Vec3", "Subtracts B from A component-wise.",
			[]string{"math", "vector", "vec3", "subtract", "minus", "difference"}, "vec3"),
		renderGraphVectorBinaryNode("multiply-vec3", "Multiply Vec3", "Multiplies two vec3 values component-wise.",
			[]string{"math", "vector", "vec3", "multiply", "product"}, "vec3"),
		renderGraphVectorBinaryNode("divide-vec3", "Divide Vec3", "Divides A by B component-wise.",
			[]string{"math", "vector", "vec3", "divide", "quotient"}, "vec3"),
		renderGraphVectorBinaryNode("add-vec4", "Add Vec4", "Adds two vec4 values component-wise.",
			[]string{"math", "vector", "vec4", "add", "plus", "sum", "color"}, "vec4"),
		renderGraphVectorBinaryNode("subtract-vec4", "Subtract Vec4", "Subtracts B from A component-wise.",
			[]string{"math", "vector", "vec4", "subtract", "minus", "difference", "color"}, "vec4"),
		renderGraphVectorBinaryNode("multiply-vec4", "Multiply Vec4", "Multiplies two vec4 values component-wise.",
			[]string{"math", "vector", "vec4", "multiply", "product", "color"}, "vec4"),
		renderGraphVectorBinaryNode("divide-vec4", "Divide Vec4", "Divides A by B component-wise.",
			[]string{"math", "vector", "vec4", "divide", "quotient", "color"}, "vec4"),
		renderGraphFloatBinaryNode("minimum", "Minimum", "Returns the smaller of two float values.",
			[]string{"math", "float", "min", "minimum"}, "A", "B", "Value"),
		renderGraphFloatBinaryNode("maximum", "Maximum", "Returns the larger of two float values.",
			[]string{"math", "float", "max", "maximum"}, "A", "B", "Value"),
		renderGraphFloatBinaryNode("power", "Power", "Raises Base to the Exponent.",
			[]string{"math", "float", "pow", "exponent"}, "Base", "Exponent", "Value"),
		renderGraphFloatUnaryNode("absolute", "Absolute", "Returns the absolute value.",
			[]string{"math", "float", "abs"}, "Value"),
		renderGraphFloatUnaryNode("one-minus", "One Minus", "Returns one minus the input value.",
			[]string{"math", "float", "invert", "complement"}, "Value"),
		renderGraphFloatUnaryNode("floor", "Floor", "Rounds a float down to the nearest integer.",
			[]string{"math", "float", "round"}, "Value"),
		renderGraphFloatUnaryNode("ceiling", "Ceiling", "Rounds a float up to the nearest integer.",
			[]string{"math", "float", "ceil", "round"}, "Value"),
		renderGraphFloatUnaryNode("fraction", "Fraction", "Returns the fractional part of a float.",
			[]string{"math", "float", "frac", "fract"}, "Value"),
		renderGraphFloatUnaryNode("sine", "Sine", "Returns the sine of the input angle.",
			[]string{"math", "float", "sin", "trig"}, "Angle"),
		renderGraphFloatUnaryNode("cosine", "Cosine", "Returns the cosine of the input angle.",
			[]string{"math", "float", "cos", "trig"}, "Angle"),
		renderGraphFloatUnaryNode("tangent", "Tangent", "Returns the tangent of the input angle.",
			[]string{"math", "float", "tan", "trig"}, "Angle"),
		renderGraphFloatUnaryNode("square-root", "Square Root", "Returns the square root of a float.",
			[]string{"math", "float", "sqrt"}, "Value"),
		renderGraphFloatTernaryNode("clamp", "Clamp", "Clamps a float between Min and Max.",
			[]string{"math", "float", "saturate", "limit"}, "Value", "Min", "Max", "Value"),
		renderGraphFloatTernaryNode("lerp", "Lerp", "Linearly interpolates between A and B by T.",
			[]string{"math", "float", "mix", "interpolate"}, "A", "B", "T", "Value"),
		{
			ID:          "step",
			Name:        "Step",
			Description: "Returns 0 or 1 by comparing Value against Edge.",
			Tags:        []string{"math", "float", "threshold", "compare"},
			Spec: renderGraphNodeSpec{
				Name:        "Step",
				Description: "Returns 0 or 1 by comparing Value against Edge.",
				Inputs: []renderGraphPortSpec{
					{Name: "Edge", Type: "float"},
					{Name: "Value", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Result", Type: "float"},
				},
			},
		},
		{
			ID:          "smoothstep",
			Name:        "Smoothstep",
			Description: "Smoothly interpolates from 0 to 1 between two edges.",
			Tags:        []string{"math", "float", "smooth", "threshold"},
			Spec: renderGraphNodeSpec{
				Name:        "Smoothstep",
				Description: "Smoothly interpolates from 0 to 1 between two edges.",
				Inputs: []renderGraphPortSpec{
					{Name: "Edge Min", Type: "float"},
					{Name: "Edge Max", Type: "float"},
					{Name: "Value", Type: "float"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Result", Type: "float"},
				},
			},
		},
		{
			ID:          "dot-product",
			Name:        "Dot Product",
			Description: "Returns the scalar dot product of two vectors.",
			Tags:        []string{"math", "vector", "vec3", "dot"},
			Spec: renderGraphNodeSpec{
				Name:        "Dot Product",
				Description: "Returns the scalar dot product of two vectors.",
				Inputs: []renderGraphPortSpec{
					{Name: "A", Type: "vec3"},
					{Name: "B", Type: "vec3"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "cross-product",
			Name:        "Cross Product",
			Description: "Returns the perpendicular cross product of two vectors.",
			Tags:        []string{"math", "vector", "vec3", "cross"},
			Spec: renderGraphNodeSpec{
				Name:        "Cross Product",
				Description: "Returns the perpendicular cross product of two vectors.",
				Inputs: []renderGraphPortSpec{
					{Name: "A", Type: "vec3"},
					{Name: "B", Type: "vec3"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
			},
		},
		{
			ID:          "normalize",
			Name:        "Normalize",
			Description: "Returns a vector with the same direction and unit length.",
			Tags:        []string{"math", "vector", "vec3", "normal"},
			Spec: renderGraphNodeSpec{
				Name:        "Normalize",
				Description: "Returns a vector with the same direction and unit length.",
				Inputs: []renderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
			},
		},
		{
			ID:          "length",
			Name:        "Length",
			Description: "Returns the length of a vector.",
			Tags:        []string{"math", "vector", "vec3", "magnitude"},
			Spec: renderGraphNodeSpec{
				Name:        "Length",
				Description: "Returns the length of a vector.",
				Inputs: []renderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "mix-color",
			Name:        "Mix Color",
			Description: "Blends two colors with a factor.",
			Tags:        []string{"mix", "blend", "color", "factor"},
			Spec: renderGraphNodeSpec{
				Name:        "Mix Color",
				Description: "Blends two colors with a factor.",
				Fields: []renderGraphNodeFieldSpec{
					{
						ID:          "clamp",
						Label:       "Clamp",
						Type:        renderGraphNodeFieldBool,
						DefaultBool: true,
					},
					{
						ID:      "mode",
						Label:   "Mode",
						Type:    renderGraphNodeFieldSelect,
						Default: "mix",
						Options: []renderGraphNodeFieldOption{
							{Label: "Mix", Value: "mix"},
							{Label: "Add", Value: "add"},
							{Label: "Multiply", Value: "multiply"},
						},
					},
				},
				Inputs: []renderGraphPortSpec{
					{Name: "Factor", Type: "float"},
					{Name: "A", Type: "color"},
					{Name: "B", Type: "color"},
				},
				Outputs: []renderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "principled-bsdf",
			Name:        "Principled BSDF",
			Description: "Surface shader with common material inputs.",
			Tags:        []string{"bsdf", "surface", "material", "shader"},
			Spec: renderGraphNodeSpec{
				Name:        "Principled BSDF",
				Description: "Surface shader with common material inputs.",
				Inputs: []renderGraphPortSpec{
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
				Outputs: []renderGraphPortSpec{
					{Name: "BSDF", Type: "surface"},
				},
			},
		},
		{
			ID:          "material-output",
			Name:        "Material Output",
			Description: "Terminal output for the material shader.",
			Tags:        []string{"output", "surface", "volume", "material"},
			Spec: renderGraphNodeSpec{
				Name:        "Material Output",
				Description: "Terminal output for the material shader.",
				Inputs: []renderGraphPortSpec{
					{Name: "Surface", Type: "surface"},
					{Name: "Displacement", Type: "float"},
				},
			},
		},
	}
}

func renderGraphFloatUnaryNode(id, name, description string, tags []string, input string) renderGraphNodeCatalogEntry {
	return renderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: renderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs: []renderGraphPortSpec{
				{Name: input, Type: "float"},
			},
			Outputs: []renderGraphPortSpec{
				{Name: "Value", Type: "float"},
			},
		},
	}
}

func renderGraphFloatBinaryNode(id, name, description string, tags []string, a, b, output string) renderGraphNodeCatalogEntry {
	return renderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: renderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs: []renderGraphPortSpec{
				{Name: a, Type: "float"},
				{Name: b, Type: "float"},
			},
			Outputs: []renderGraphPortSpec{
				{Name: output, Type: "float"},
			},
		},
	}
}

func renderGraphFloatTernaryNode(id, name, description string, tags []string, a, b, c, output string) renderGraphNodeCatalogEntry {
	return renderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: renderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs: []renderGraphPortSpec{
				{Name: a, Type: "float"},
				{Name: b, Type: "float"},
				{Name: c, Type: "float"},
			},
			Outputs: []renderGraphPortSpec{
				{Name: output, Type: "float"},
			},
		},
	}
}

func renderGraphVectorBinaryNode(id, name, description string, tags []string, vectorType string) renderGraphNodeCatalogEntry {
	outputs := []renderGraphPortSpec{{Name: "Vector", Type: vectorType}}
	if vectorType == "vec4" {
		outputs = append(outputs, renderGraphPortSpec{Name: "Color", Type: "color"})
	}
	return renderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: renderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs: []renderGraphPortSpec{
				{Name: "A", Type: vectorType},
				{Name: "B", Type: vectorType},
			},
			Outputs: outputs,
		},
	}
}

func renderGraphCombineVectorNode(id, name, description string, tags, components []string, outputType string) renderGraphNodeCatalogEntry {
	inputs := make([]renderGraphPortSpec, len(components))
	for i := range components {
		inputs[i] = renderGraphPortSpec{Name: components[i], Type: "float"}
	}
	outputs := []renderGraphPortSpec{{Name: "Vector", Type: outputType}}
	if outputType == "vec4" {
		outputs = append(outputs, renderGraphPortSpec{Name: "Color", Type: "color"})
	}
	return renderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: renderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs:      inputs,
			Outputs:     outputs,
		},
	}
}

func renderGraphSplitVectorNode(id, name, description string, tags, components []string, inputType string) renderGraphNodeCatalogEntry {
	outputs := make([]renderGraphPortSpec, len(components))
	for i := range components {
		outputs[i] = renderGraphPortSpec{Name: components[i], Type: "float"}
	}
	return renderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: renderGraphNodeSpec{
			Name:        name,
			Description: description,
			Inputs: []renderGraphPortSpec{
				{Name: "Vector", Type: inputType},
			},
			Outputs: outputs,
		},
	}
}

func renderGraphSwizzleVectorNode(id, name, description string, tags, components []string, vectorType string) renderGraphNodeCatalogEntry {
	fields := make([]renderGraphNodeFieldSpec, len(components))
	options := renderGraphSwizzleFieldOptions(components)
	for i := range components {
		fields[i] = renderGraphNodeFieldSpec{
			ID:      strings.ToLower(components[i]),
			Label:   components[i],
			Type:    renderGraphNodeFieldSelect,
			Default: strings.ToLower(components[i]),
			Options: options,
		}
	}
	outputs := []renderGraphPortSpec{{Name: "Vector", Type: vectorType}}
	if vectorType == "vec4" {
		outputs = append(outputs, renderGraphPortSpec{Name: "Color", Type: "color"})
	}
	return renderGraphNodeCatalogEntry{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Spec: renderGraphNodeSpec{
			Name:        name,
			Description: description,
			Fields:      fields,
			Inputs: []renderGraphPortSpec{
				{Name: "Vector", Type: vectorType},
			},
			Outputs: outputs,
		},
	}
}

func renderGraphSwizzleFieldOptions(components []string) []renderGraphNodeFieldOption {
	options := make([]renderGraphNodeFieldOption, 0, len(components)+2)
	for i := range components {
		component := strings.ToLower(components[i])
		label := components[i]
		if len(components) == 4 {
			label += " / " + []string{"R", "G", "B", "A"}[i]
		}
		options = append(options, renderGraphNodeFieldOption{Label: label, Value: component})
	}
	options = append(options,
		renderGraphNodeFieldOption{Label: "0", Value: "0"},
		renderGraphNodeFieldOption{Label: "1", Value: "1"},
	)
	return options
}

func renderGraphNodeCatalogMenuData() []renderGraphNodeMenuData {
	catalog := renderGraphNodeCatalog()
	data := make([]renderGraphNodeMenuData, 0, len(catalog))
	for i := range catalog {
		entry := catalog[i]
		search := strings.Join(append([]string{entry.ID, entry.Name, entry.Description}, entry.Tags...), " ")
		data = append(data, renderGraphNodeMenuData{
			ID:          entry.ID,
			Name:        entry.Name,
			Description: entry.Description,
			Search:      strings.ToLower(search),
		})
	}
	return data
}

func renderGraphNodeCatalogEntryCompatible(entry renderGraphNodeCatalogEntry, compatibility renderGraphNodePortCompatibility) bool {
	if !compatibility.Active {
		return true
	}
	_, ok := renderGraphNodeSpecCompatiblePortIndex(entry.Spec, compatibility.SourceOutput, compatibility.Type)
	return ok
}

func renderGraphNodeSpecCompatiblePortIndex(spec renderGraphNodeSpec, sourceOutput bool, sourceType string) (int, bool) {
	ports := spec.Outputs
	if sourceOutput {
		ports = spec.Inputs
	}
	sourceType = renderGraphPortTypeKey(sourceType)
	for i := range ports {
		if renderGraphPortTypeKey(ports[i].Type) == sourceType {
			return i, true
		}
	}
	return -1, false
}

func renderGraphNodeCatalogCompatibleIDs(sourceOutput bool, sourceType string) []string {
	catalog := renderGraphNodeCatalog()
	out := make([]string, 0, len(catalog))
	compatibility := renderGraphNodePortCompatibility{
		Active:       true,
		SourceOutput: sourceOutput,
		Type:         sourceType,
	}
	for i := range catalog {
		if renderGraphNodeCatalogEntryCompatible(catalog[i], compatibility) {
			out = append(out, catalog[i].ID)
		}
	}
	return out
}

func renderGraphNodeCatalogSpec(id string) (renderGraphNodeSpec, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, entry := range renderGraphNodeCatalog() {
		if entry.ID == id {
			return entry.Spec, true
		}
	}
	return renderGraphNodeSpec{}, false
}
