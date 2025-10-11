/******************************************************************************/
/* shader_data.go                                                             */
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

package ui

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

type ShaderData struct {
	rendering.ShaderDataBase
	UVs          matrix.Vec4
	FgColor      matrix.Color
	BgColor      matrix.Color
	Scissor      matrix.Vec4
	Size2D       matrix.Vec4
	BorderRadius matrix.Vec4
	BorderSize   matrix.Vec4
	BorderColor  [4]matrix.Color
	BorderLen    matrix.Vec2
}

func (s ShaderData) Size() int {
	const size = int(unsafe.Sizeof(ShaderData{}) - rendering.ShaderBaseDataStart)
	return size
}

func (s *ShaderData) setSize2d(ui *UI, texWidth, texHeight float32) {
	// TODO:  This is skipped for text
	ws := ui.Entity().Transform.WorldScale()
	s.Size2D[0] = ws.X()
	s.Size2D[1] = ws.Y()
	s.Size2D[2] = texWidth
	s.Size2D[3] = texHeight
}
