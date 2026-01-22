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

package assets

// Textures
const (
	TextureSquare      = "square.png"
	TextureCube        = "cube.png"
	TextureBlankSquare = "blank_square.png"
)

// Material definitions
const (
	MaterialDefinitionGrid                = "grid.material"
	MaterialDefinitionUnlit               = "unlit.material"
	MaterialDefinitionUnlitTransparent    = "unlit_transparent.material"
	MaterialDefinitionBasic               = "basic.material"
	MaterialDefinitionBasicTransparent    = "basic_transparent.material"
	MaterialDefinitionPBR                 = "pbr.material"
	MaterialDefinitionTerrain             = "terrain.material"
	MaterialDefinitionBasicSkinned        = "basic_skinned.material"
	MaterialDefinitionText3D              = "text3d.material"
	MaterialDefinitionText3DTransparent   = "text3d_transparent.material"
	MaterialDefinitionText                = "text.material"
	MaterialDefinitionTextTransparent     = "text_transparent.material"
	MaterialDefinitionCombine             = "combine.material"
	MaterialDefinitionComposite           = "composite.material"
	MaterialDefinitionUI                  = "ui.material"
	MaterialDefinitionUITransparent       = "ui_transparent.material"
	MaterialDefinitionSprite              = "sprite.material"
	MaterialDefinitionSpriteTransparent   = "sprite_transparent.material"
	MaterialDefinitionLightDepth          = "light_depth.material"
	MaterialDefinitionLightDepthCSM1      = "light_depth_csm1.material"
	MaterialDefinitionLightDepthCSM2      = "light_depth_csm2.material"
	MaterialDefinitionLightCubeDepth      = "light_cube_depth.material"
	MaterialDefinitionParticle            = "particle.material"
	MaterialDefinitionParticleTransparent = "particle_transparent.material"

	MaterialDefinitionEdTransformWire = "ed_transform_wire.material"
	MaterialDefinitionEdFrustumWire   = "ed_frustum_wire.material"
	MaterialDefinitionEdGizmo         = "ed_gizmo.material"
)
