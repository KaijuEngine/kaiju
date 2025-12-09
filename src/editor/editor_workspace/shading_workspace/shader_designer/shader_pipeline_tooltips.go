/******************************************************************************/
/* shader_pipeline_tooltips.go                                                */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package shader_designer

var pipelineTooltips = map[string]string{
	// vkPipelineInputAssemblyStateCreateInfo
	"Topology":         "Specifies the primitive topology used for rendering, defining how vertices are assembled into primitives (e.g., points, lines, or triangles). Common options include TRIANGLE_LIST for rendering triangles or LINE_LIST for lines. This determines the structure of the geometry processed by the pipeline.",
	"PrimitiveRestart": "Enables or disables primitive restart, a feature that allows breaking a single primitive strip (e.g., triangle strip) into multiple segments. When enabled (set to TRUE), a special index value (e.g., 0xFFFFFFFF for 32-bit indices) in the index buffer restarts the primitive. Useful for rendering complex shapes efficiently in a single draw call.",

	// vkPipelineRasterizationStateCreateInfo
	"DepthClampEnable":        "Enables depth clamping to restrict depth values to the [0, 1] range instead of clipping fragments outside this range. When set to TRUE, fragments beyond the near or far planes are clamped rather than discarded, useful for effects like shadow mapping.",
	"RasterizerDiscardEnable": "Controls whether the rasterizer discards all primitives before processing. Set to TRUE to disable rasterization (e.g., for transform feedback-only pipelines); FALSE allows normal rendering to proceed.",
	"PolygonMode":             "Defines how polygons are rendered: FILL for solid shapes, LINE for wireframes, or POINT for points at vertices. Determines the visual style of rendered geometry.",
	"CullMode":                "Specifies which polygon faces are culled (discarded): NONE (no culling), FRONT_BIT (cull front-facing), BACK_BIT (cull back-facing), or FRONT_AND_BACK (cull both). Optimizes rendering by skipping unseen faces.",
	"FrontFace":               "Defines which polygon winding order is considered front-facing: COUNTER_CLOCKWISE for counterclockwise vertex order or CLOCKWISE for clockwise. Used with culling to determine visible faces.",
	"DepthBiasEnable":         "Enables depth bias (offset) adjustments to depth values. Set to TRUE to apply bias (controlled by other depth bias fields); useful for fixing z-fighting artifacts in shadow mapping or decals.",
	"DepthBiasConstantFactor": "A constant value added to the depth of each fragment when depth bias is enabled. Positive values push fragments away, negative values pull them closer. Helps resolve depth precision issues like z-fighting.",
	"DepthBiasClamp":          "Sets the maximum (or minimum) value for the depth bias adjustment. Limits the effect of bias to prevent extreme shifts, ensuring subtle corrections (e.g., for shadows or decals).",
	"DepthBiasSlopeFactor":    "A scaling factor applied to the fragment's slope (rate of depth change) when calculating depth bias. Adjusts bias dynamically based on polygon steepness, useful for consistent shadow rendering on sloped surfaces.",
	"LineWidth":               "Sets the width of lines in pixels when rendering in line mode (e.g., LINE). Typically 1.0 by default; values greater than 1.0 may require wide line support from the device.",

	// vkPipelineMultisampleStateCreateInfo
	"RasterizationSamples":  "Specifies the number of samples per pixel used for multisample anti-aliasing (MSAA), such as 1_BIT (no MSAA) or 4_BIT (4x MSAA). Higher counts improve edge smoothness but increase performance cost.",
	"SampleShadingEnable":   "Enables sample shading when set to TRUE. Forces the shader to run per-sample (instead of per-fragment) for finer anti-aliasing control, improving quality at a performance cost. Requires multisampling to be active.",
	"MinSampleShading":      "Sets the minimum fraction of samples shaded per fragment when sample shading is enabled (range 0.0 to 1.0). A value of 1.0 ensures all samples are shaded, enhancing anti-aliasing quality; lower values balance quality and performance.",
	"AlphaToCoverageEnable": "Enables alpha-to-coverage when set to TRUE. Converts the fragment's alpha value into a coverage mask for multisampling, useful for rendering transparent effects (e.g., foliage) with smoother edges.",
	"AlphaToOneEnable":      "Forces the alpha value of fragments to 1.0 when set to TRUE, while preserving RGB values. Useful for specific blending scenarios or when alpha-to-coverage needs full opacity, though rarely used in modern pipelines.",

	// vkPipelineColorBlendAttachmentState
	"BlendEnable":         "Enables blending for this attachment when set to TRUE. When enabled, the source and destination colors/alpha are combined using the specified blend factors and operations; otherwise, the source color overwrites the destination.",
	"SrcColorBlendFactor": "Defines the factor applied to the source RGB values during blending (e.g., ONE, SRC_ALPHA). Determines how much of the source color contributes to the final output.",
	"DstColorBlendFactor": "Defines the factor applied to the destination RGB values during blending (e.g., ZERO, ONE_MINUS_SRC_ALPHA). Controls the influence of the existing color in the framebuffer.",
	"ColorBlendOp":        "Specifies the operation used to combine source and destination RGB values (e.g., ADD, SUBTRACT). Defines how blending merges colors, such as addition for transparency or subtraction for effects.",
	"SrcAlphaBlendFactor": "Sets the factor applied to the source alpha value during blending (e.g., ONE, SRC_ALPHA). Controls the source alpha's contribution to the final alpha value.",
	"DstAlphaBlendFactor": "Sets the factor applied to the destination alpha value during blending (e.g., ZERO, ONE_MINUS_SRC_ALPHA). Determines how the existing alpha in the framebuffer affects the result.",
	"AlphaBlendOp":        "Specifies the operation for combining source and destination alpha values (e.g., ADD, MULTIPLY). Defines how alpha blending is computed, often matching the color operation for consistency.",
	"ColorWriteMask":      "Controls which color channels (R, G, B, A) are written to the framebuffer. Uses a bitmask (e.g., R_BIT | G_BIT) to enable/disable writing specific channels, useful for selective updates.",

	// vkPipelineColorBlendStateCreateInfo
	"LogicOpEnable":  "Enables logical operations for color blending when set to TRUE. When enabled, the specified logic operation (e.g., AND, OR) is applied to source and destination colors instead of standard blending. Ignored if blending is enabled on any attachment.",
	"LogicOp":        "Specifies the logical operation applied to source and destination colors when LogicOpEnable is TRUE (e.g., XOR, AND). Defines how pixel values are combined, useful for bitwise effects rather than typical alpha blending.",
	"BlendConstants": "A four-component array (R, G, B, A) of float values used as constant blend factors in blending equations. Referenced when blend factors like COLOR or ALPHA are set, providing a fixed value for blending calculations.",

	// vkPipelineDepthStencilStateCreateInfo
	"DepthTestEnable":                 "Enables depth testing when set to TRUE. When active, compares the fragment's depth value against the depth buffer to determine if it should be rendered or discarded, based on the DepthCompareOp.",
	"DepthWriteEnable":                "Controls whether depth values are written to the depth buffer. Set to TRUE to update the depth buffer with new fragment depths; FALSE prevents writes, useful for transparent objects or depth-only passes.",
	"DepthCompareOp":                  "Defines the comparison operation for depth testing (e.g., LESS, EQUAL). Determines if a fragment passes the depth test by comparing its depth against the depth buffer value.",
	"DepthBoundsTestEnable":           "Enables depth bounds testing when set to TRUE. Restricts fragment rendering to a specified range (MinDepthBounds to MaxDepthBounds), discarding fragments outside this range. Requires device support.",
	"StencilTestEnable":               "Enables stencil testing when set to TRUE. Activates stencil operations, comparing the fragment's stencil value against a reference value to control rendering, based on Front and Back settings.",
	"PipelineDepthStencilState.Front": "Defines stencil test and operation settings for front-facing polygons. Includes reference value, comparison operation (e.g., ALWAYS), and actions (e.g., KEEP) for pass/fail/depth-fail cases.",
	"PipelineDepthStencilState.Back":  "Defines stencil test and operation settings for back-facing polygons. Similar to Front, specifying reference value, comparison, and operations to manage stencil buffer updates for back faces.",
	"MinDepthBounds":                  "Sets the minimum depth value for the depth bounds test (range 0.0 to 1.0). Fragments with depth values less than this are discarded if DepthBoundsTestEnable is TRUE.",
	"MaxDepthBounds":                  "Sets the maximum depth value for the depth bounds test (range 0.0 to 1.0). Fragments with depth values greater than this are discarded if DepthBoundsTestEnable is TRUE.",
	"FailOp":                          "Specifies the operation applied to the stencil buffer when the stencil test fails for front-facing polygons (e.g., KEEP, INCREMENT). Defines how the stencil value changes if the comparison doesn't pass.",
	"PassOp":                          "Sets the operation applied to the stencil buffer when both the stencil and depth tests pass for front-facing polygons (e.g., REPLACE, ZERO). Controls the stencil update on successful rendering.",
	"DepthFailOp":                     "Defines the operation applied to the stencil buffer when the stencil test passes but the depth test fails for front-facing polygons (e.g., KEEP, DECREMENT). Handles cases where depth prevents rendering.",
	"CompareOp":                       "Specifies the comparison operation for the stencil test (e.g., LESS, EQUAL). Compares the Reference value (masked by CompareMask) against the stencil buffer value (masked similarly) to determine test outcome.",
	"CompareMask":                     "A bitmask that selects which bits of the stencil buffer value and Reference are used in the stencil comparison. Bits set to 1 are compared; bits set to 0 are ignored, allowing selective stencil testing.",
	"WriteMask":                       "A bitmask that determines which bits of the stencil buffer are updated by stencil operations (e.g., FailOp, PassOp). Bits set to 1 allow writes; bits set to 0 preserve existing values, enabling partial updates.",
	"Reference":                       "The reference value used in the stencil test comparison for front-facing polygons. Compared against the stencil buffer value (after applying CompareMask) using the CompareOp to decide if the test passes.",

	// vkPipelineTessellationStateCreateInfo
	"PatchControlPoints": "Specifies the number of control points per patch in tessellation. Defines how many vertices form each patch (e.g., 3 for a triangle, 4 for a quad), which the tessellation shader processes to generate additional geometry. Must match the shader's expectation.",
}
