/******************************************************************************/
/* shader_tooltips.go                                                         */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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

var shaderTooltips = map[string]string{
	"Vertex":                 "This field specifies the file path or source code for the vertex shader, which is executed for each vertex in your geometry. It transforms vertex positions, normals, and other per-vertex data from object space to screen space (via world, view, and projection transformations). In Vulkan, this shader is a required stage in the graphics pipeline, defining how raw vertex data from buffers is processed before primitive assembly.",
	"Fragment":               "This field holds the file path or source code for the fragment shader, which runs for each pixel (fragment) generated after rasterization. It determines the final color and depth of a fragment, often handling lighting calculations, texturing, and material properties. In Vulkan, this is another mandatory stage, critical for defining the visual appearance of surfaces in your rendered scene.",
	"Geometry":               "This optional field points to the file path or source code for the geometry shader, which operates on entire primitives (e.g., triangles, lines) after vertex processing. It can modify, discard, or generate new geometry on the fly, making it useful for effects like particle systems or silhouette enhancement. In Vulkan, this stage is optional and only included if advanced primitive manipulation is needed.",
	"TessellationControl":    "This field specifies the file path or source code for the tessellation control shader, part of Vulkan’s tessellation pipeline. It runs per vertex in a patch (a higher-order primitive) and defines how much a surface should be subdivided into smaller triangles. It sets the tessellation levels, controlling detail and smoothness, and is optional unless your renderer uses tessellation for enhanced geometry detail.",
	"TessellationEvaluation": "This field contains the file path or source code for the tessellation evaluation shader, which executes after the tessellation control stage. It processes the newly generated vertices from tessellation, positioning them to form the final subdivided surface (e.g., applying displacement mapping). In Vulkan, this optional stage completes the tessellation process, refining the geometry before it reaches the rasterizer.",
	"CompileFlags":           "Compile flags allows you to specify arguments to the glslc compiler. This is especially helpful for if your shaders use things like #define to enable/disable blocks of code and you want to set that definition value.",
}
