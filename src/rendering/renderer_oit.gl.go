//go:build OPENGL

/*****************************************************************************/
/* renderer_oit.gl.go                                                        */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
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
	"kaiju/matrix"
	"log"
)

func (r *GLRenderer) setupOITFrameBuffer(width, height int32) {
	gl.DeleteFrameBuffers(1, &r.opaqueFBO)
	gl.DeleteFrameBuffers(1, &r.transparentAccumFBO)
	gl.DeleteFrameBuffers(1, &r.transparentRevealFBO)
	gl.DeleteTextures(1, &r.opaqueTexture)
	gl.DeleteTextures(1, &r.accumTexture)
	gl.DeleteTextures(1, &r.revealAccumTexture)
	gl.DeleteTextures(1, &r.revealRevealTexture)
	gl.DeleteTextures(1, &r.revealTexture)
	gl.DeleteTextures(1, &r.depthTexture)

	gl.GenTextures(1, &r.opaqueTexture)
	gl.BindTexture(gl.Texture2D, r.opaqueTexture)
	gl.TexImage2D(gl.Texture2D, 0, gl.RGBA16F, width, height, 0, gl.RGBA, gl.HalfFloat, nil)
	gl.TexParameteri(gl.Texture2D, gl.TextureMinFilter, gl.Linear)
	gl.TexParameteri(gl.Texture2D, gl.TextureMagFilter, gl.Linear)
	gl.UnBindTexture(gl.Texture2D)

	gl.GenTextures(1, &r.depthTexture)
	gl.BindTexture(gl.Texture2D, r.depthTexture)
	gl.TexImage2D(gl.Texture2D, 0, gl.DepthComponent32F,
		width, height, 0, gl.DepthComponent, gl.Float, nil)
	gl.UnBindTexture(gl.Texture2D)

	gl.GenFrameBuffers(1, &r.opaqueFBO)
	gl.BindFrameBuffer(gl.FrameBuffer, r.opaqueFBO)
	gl.FrameBufferTexture2D(gl.FrameBuffer, gl.ColorAttachment0, gl.Texture2D, r.opaqueTexture, 0)
	gl.FrameBufferTexture2D(gl.FrameBuffer, gl.DepthAttachment, gl.Texture2D, r.depthTexture, 0)

	if !gl.CheckFrameBufferStatus(gl.FrameBuffer).Equal(gl.FrameBufferComplete) {
		log.Fatalf("%s\n", "FrameBuffer opaque FBO not complete")
	}

	gl.UnBindFrameBuffer(gl.FrameBuffer)

	gl.GenTextures(1, &r.accumTexture)
	gl.BindTexture(gl.Texture2D, r.accumTexture)
	gl.TexImage2D(gl.Texture2D, 0, gl.RGBA16F, width, height, 0, gl.RGBA, gl.HalfFloat, nil)
	gl.TexParameteri(gl.Texture2D, gl.TextureMinFilter, gl.Linear)
	gl.TexParameteri(gl.Texture2D, gl.TextureMagFilter, gl.Linear)
	gl.UnBindTexture(gl.Texture2D)

	gl.GenTextures(1, &r.revealRevealTexture)
	gl.BindTexture(gl.Texture2D, r.revealRevealTexture)
	gl.TexImage2D(gl.Texture2D, 0, gl.R32F, width, height, 0, gl.Red, gl.Float, nil)
	gl.TexParameteri(gl.Texture2D, gl.TextureMinFilter, gl.Linear)
	gl.TexParameteri(gl.Texture2D, gl.TextureMagFilter, gl.Linear)
	gl.UnBindTexture(gl.Texture2D)

	gl.GenFrameBuffers(1, &r.transparentAccumFBO)
	gl.BindFrameBuffer(gl.FrameBuffer, r.transparentAccumFBO)
	gl.FrameBufferTexture2D(gl.FrameBuffer, gl.ColorAttachment0, gl.Texture2D, r.accumTexture, 0)
	gl.FrameBufferTexture2D(gl.FrameBuffer, gl.ColorAttachment1, gl.Texture2D, r.revealRevealTexture, 0)
	gl.FrameBufferTexture2D(gl.FrameBuffer, gl.DepthAttachment, gl.Texture2D, r.depthTexture, 0)

	accumDrawBuffers := []gl.Handle{gl.ColorAttachment0, gl.ColorAttachment1}
	gl.DrawBuffers(accumDrawBuffers)

	if !gl.CheckFrameBufferStatus(gl.FrameBuffer).Equal(gl.FrameBufferComplete) {
		log.Fatalf("%s\n", "FrameBuffer transparent FBO not complete")
	}

	gl.UnBindFrameBuffer(gl.FrameBuffer)

	gl.GenTextures(1, &r.revealAccumTexture)
	gl.BindTexture(gl.Texture2D, r.revealAccumTexture)
	gl.TexImage2D(gl.Texture2D, 0, gl.RGBA16F, width, height, 0, gl.RGBA, gl.HalfFloat, nil)
	gl.TexParameteri(gl.Texture2D, gl.TextureMinFilter, gl.Linear)
	gl.TexParameteri(gl.Texture2D, gl.TextureMagFilter, gl.Linear)
	gl.UnBindTexture(gl.Texture2D)

	gl.GenTextures(1, &r.revealTexture)
	gl.BindTexture(gl.Texture2D, r.revealTexture)
	gl.TexImage2D(gl.Texture2D, 0, gl.R32F, width, height, 0, gl.Red, gl.Float, nil)
	gl.TexParameteri(gl.Texture2D, gl.TextureMinFilter, gl.Linear)
	gl.TexParameteri(gl.Texture2D, gl.TextureMagFilter, gl.Linear)
	gl.UnBindTexture(gl.Texture2D)

	gl.GenFrameBuffers(1, &r.transparentRevealFBO)
	gl.BindFrameBuffer(gl.FrameBuffer, r.transparentRevealFBO)
	gl.FrameBufferTexture2D(gl.FrameBuffer, gl.ColorAttachment0, gl.Texture2D, r.revealAccumTexture, 0)
	gl.FrameBufferTexture2D(gl.FrameBuffer, gl.ColorAttachment1, gl.Texture2D, r.revealTexture, 0)
	gl.FrameBufferTexture2D(gl.FrameBuffer, gl.DepthAttachment, gl.Texture2D, r.depthTexture, 0)

	revealDrawBuffers := []gl.Handle{gl.ColorAttachment0, gl.ColorAttachment1}
	gl.DrawBuffers(revealDrawBuffers)
	if !gl.CheckFrameBufferStatus(gl.FrameBuffer).Equal(gl.FrameBufferComplete) {
		log.Fatalf("%s\n", "FrameBuffer transparent reveal FBO not complete")
	}
	gl.UnBindFrameBuffer(gl.FrameBuffer)
}

func (r *GLRenderer) solidPass(drawings []ShaderDraw, clearColor matrix.Color) {
	gl.Enable(gl.DepthTest)
	gl.DepthFunc(gl.Less)
	gl.DepthMask(true)
	gl.Disable(gl.Blend)
	gl.ClearColor(clearColor.R(), clearColor.G(), clearColor.B(), clearColor.A())
	gl.BindFrameBuffer(gl.FrameBuffer, r.opaqueFBO)
	gl.Clear(gl.ColorBufferBit | gl.DepthBufferBit)
	r.draw(drawings)
}

func (r *GLRenderer) transparentPass(drawings []ShaderDraw) {
	gl.DepthMask(false)
	gl.Enable(gl.Blend)
	// TODO:  Figure this out, blend func doesn't take in an arg num to first arg
	gl.BlendFunc(gl.One, gl.One)
	gl.BlendEquation(gl.FuncAdd)
	gl.BindFrameBuffer(gl.FrameBuffer, r.transparentAccumFBO)
	gl.ClearBufferfv(gl.Color, 0, matrix.Vec4Zero())
	gl.ClearBufferfv(gl.Color, 1, matrix.Vec4One())
	r.draw(drawings)

	gl.BlendFunc(gl.Zero, gl.OneMinusSrcColor)
	gl.BindFrameBuffer(gl.FrameBuffer, r.transparentRevealFBO)
	gl.ClearBufferfv(gl.Color, 0, matrix.Vec4Zero())
	gl.ClearBufferfv(gl.Color, 1, matrix.Vec4One())
	r.draw(drawings)
}

func (r *GLRenderer) composePass() {
	id := r.compositeShader.RenderId.(gl.Handle)
	meshId := r.composeQuad.MeshId.(MeshIdGL)
	gl.DepthFunc(gl.Always)
	gl.Enable(gl.Blend)
	gl.BlendFunc(gl.SrcAlpha, gl.OneMinusSrcAlpha)
	gl.BindFrameBuffer(gl.FrameBuffer, r.opaqueFBO)
	gl.UseProgram(id)
	gl.ActivateTexture(gl.Texture0)
	gl.BindTexture(gl.Texture2D, r.accumTexture)
	gl.Uniform1i(gl.GetUniformLocation(id, "accum"), 0)
	gl.ActivateTexture(gl.Texture1)
	gl.BindTexture(gl.Texture2D, r.revealTexture)
	gl.Uniform1i(gl.GetUniformLocation(id, "reveal"), 1)
	gl.BindVertexArray(meshId.VAO)
	gl.BindBuffer(gl.ElementArrayBuffer, meshId.EBO)
	gl.DrawElementsInstanced(gl.Triangles, 6, gl.UnsignedInt, 0, 1)
	gl.UnBindBuffer(gl.ElementArrayBuffer)
	gl.UnBindTexture(gl.Texture2D)
	gl.UnBindVertexArray()
}
