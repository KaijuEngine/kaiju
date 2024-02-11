//go:build OPENGL

/*****************************************************************************/
/* draw_instance.gl.go                                                       */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package rendering

import (
	"kaiju/gl"
	"unsafe"
)

type InstanceDriverData gl.Handle

func (d *DrawInstanceGroup) generateInstanceDriverData(renderer Renderer, shader *Shader) {
	gl.DeleteTextures(1, &d.InstanceDriverData)
	gl.GenTextures(1, &d.InstanceDriverData)
	gl.BindTexture(gl.Texture2D, d.InstanceDriverData)
	gl.TexParameteri(gl.Texture2D, gl.TextureWrapS, gl.ClampToEdge)
	gl.TexParameteri(gl.Texture2D, gl.TextureWrapT, gl.ClampToEdge)
	gl.TexParameteri(gl.Texture2D, gl.TextureMinFilter, gl.Nearest)
	gl.TexParameteri(gl.Texture2D, gl.TextureMagFilter, gl.Nearest)
	gl.UnBindTexture(gl.Texture2D)
}

func (d *DrawInstanceGroup) bindInstanceDriverData() {
	gl.BindTexture(gl.Texture2D, d.InstanceDriverData)
	w, h := d.texSize()
	gl.TexImage2D(gl.Texture2D, 0, gl.RGBA32F, w, h, 0,
		gl.RGBA, gl.Float, unsafe.Pointer(&d.instanceData[0]))
	gl.UnBindTexture(gl.Texture2D)
}
