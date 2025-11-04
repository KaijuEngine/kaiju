/******************************************************************************/
/* assets_config.go                                                           */
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

package assets

// Textures
const (
	TextureSquare      = "square.png"
	TextureBlankSquare = "blank_square.png"
	TextureTriangle    = "triangle.png"
)

// Material definitions
const (
	MaterialDefinitionGrid                = "grid"
	MaterialDefinitionUnlit               = "unlit"
	MaterialDefinitionUnlitTransparent    = "unlit_transparent"
	MaterialDefinitionBasic               = "basic"
	MaterialDefinitionBasicLit            = "basic_lit"
	MaterialDefinitionBasicLitStatic      = "basic_lit_static"
	MaterialDefinitionBasicLitDynamic     = "basic_lit_dynamic"
	MaterialDefinitionBasicLitTransparent = "basic_lit_transparent"
	MaterialDefinitionBasicTransparent    = "basic_transparent"
	MaterialDefinitionPBR                 = "pbr"
	MaterialDefinitionTerrain             = "terrain"
	MaterialDefinitionBasicSkinned        = "basic_skinned"
	MaterialDefinitionBasicColor          = "basic_color"
	MaterialDefinitionText3D              = "text3d"
	MaterialDefinitionText                = "text"
	MaterialDefinitionCombine             = "combine"
	MaterialDefinitionComposite           = "composite"
	MaterialDefinitionUI                  = "ui"
	MaterialDefinitionUITransparent       = "ui_transparent"
	MaterialDefinitionSprite              = "sprite"
	MaterialDefinitionSpriteTransparent   = "sprite_transparent"
	MaterialDefinitionOutline             = "outline"
	MaterialDefinitionLightDepth          = "light_depth"
	MaterialDefinitionLightCubeDepth      = "light_cube_depth"

	MaterialDefinitionEdTransformWire = "ed_transform_wire"
	MaterialDefinitionEdFrustumWire   = "ed_frustum_wire"
	MaterialDefinitionEdGizmo         = "ed_gizmo"
)
