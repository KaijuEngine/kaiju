//go:build !js && !wasm

/*****************************************************************************/
/* gles.nonjs.go                                                             */
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

package gl

/*
#include <stdlib.h>
#include "dist/glad.h"

GLuint cglCreateShader(GLenum type) {
	return glCreateShader(type);
}

void cglShaderSource(GLuint shader, GLsizei count, const GLchar **string, const GLint *length) {
	glShaderSource(shader, count, string, length);
}

void cglCompileShader(GLuint shader) {
	glCompileShader(shader);
}

void cglGetShaderiv(GLuint shader, GLenum pname, GLint *params) {
	glGetShaderiv(shader, pname, params);
}

void cglClearColor(GLfloat red, GLfloat green, GLfloat blue, GLfloat alpha) {
	glClearColor(red, green, blue, alpha);
}

void cglClear(GLbitfield mask) {
	glClear(mask);
}

void cglGetShaderInfoLog(GLuint shader, GLsizei bufSize, GLsizei *length, GLchar *infoLog) {
	glGetShaderInfoLog(shader, bufSize, length, infoLog);
}

GLuint cglCreateProgram() {
	return glCreateProgram();
}

void cglAttachShader(GLuint program, GLuint shader) {
	glAttachShader(program, shader);
}

void cglLinkProgram(GLuint program) {
	glLinkProgram(program);
}

void cglGetProgramiv(GLuint program, GLenum pname, GLint *params) {
	glGetProgramiv(program, pname, params);
}

void cglGetProgramInfoLog(GLuint program, GLsizei bufSize, GLsizei *length, GLchar *infoLog) {
	glGetProgramInfoLog(program, bufSize, length, infoLog);
}

void cglDeleteShader(GLuint shader) {
	glDeleteShader(shader);
}

void cglDeleteProgram(GLuint program) {
	glDeleteProgram(program);
}

void cglDeleteTextures(GLsizei n, GLuint *textures) {
	glDeleteTextures(n, textures);
}

void cglDeleteFramebuffers(GLsizei n, GLuint *framebuffers) {
	glDeleteFramebuffers(n, framebuffers);
}

void cglGenVertexArrays(GLsizei n, GLuint *arrays) {
	glGenVertexArrays(n, arrays);
}

void cglGenBuffers(GLsizei n, GLuint *buffers) {
	glGenBuffers(n, buffers);
}

void cglGenFramebuffers(GLsizei n, GLuint *framebuffers) {
	glGenFramebuffers(n, framebuffers);
}

void cglGenTextures(GLsizei n, GLuint *textures) {
	glGenTextures(n, textures);
}

void cglBindVertexArray(GLuint array) {
	glBindVertexArray(array);
}

void cglBindBuffer(GLenum target, GLuint buffer) {
	glBindBuffer(target, buffer);
}

void cglBindTexture(GLenum target, GLuint texture) {
	glBindTexture(target, texture);
}

void cglBindFramebuffer(GLenum target, GLuint framebuffer) {
	glBindFramebuffer(target, framebuffer);
}

void cglActiveTexture(GLenum texture) {
	glActiveTexture(texture);
}

void cglUniform1i(GLint location, GLint value) {
	glUniform1i(location, value);
}

void cglTexImage2D(GLenum target, GLint level, GLint internalFormat, GLsizei width, GLsizei height, GLint border, GLenum format, GLenum type, const void *pixels) {
	glTexImage2D(target, level, internalFormat, width, height, border, format, type, pixels);
}

void cglCompressedTexImage2D(GLenum target, GLint level, GLint internalFormat, GLsizei width, GLsizei height, GLint border, GLsizei imageSize, const void *data) {
	glCompressedTexImage2D(target, level, internalFormat, width, height, border, imageSize, data);
}

void cglGenerateMipmap(GLenum target) {
	glGenerateMipmap(target);
}

void cglGetTexImage(GLenum target, GLint level, GLenum format, GLenum type, void *pixels) {
	glGetTexImage(target, level, format, type, pixels);
}

void cglFrameBufferTexture2D(GLenum target, GLenum attachment, GLenum textarget, GLuint texture, GLint level) {
	glFramebufferTexture2D(target, attachment, textarget, texture, level);
}

void cglTexParameteri(GLenum target, GLenum pname, GLint param) {
	glTexParameteri(target, pname, param);
}

void cglBufferData(GLenum target, GLsizeiptr size, const void *data, GLenum usage) {
	glBufferData(target, size, data, usage);
}

void cglVertexAttribPointer(GLuint index, GLint size, GLenum type, GLboolean normalized, GLsizei stride, const void *pointer) {
	glVertexAttribPointer(index, size, type, normalized, stride, pointer);
}

void cglEnableVertexAttribArray(GLuint index) {
	glEnableVertexAttribArray(index);
}

void cglUseProgram(GLuint program) {
	glUseProgram(program);
}

void cglDrawArrays(GLenum mode, GLint first, GLsizei count) {
	glDrawArrays(mode, first, count);
}

void cglDrawElementsInstanced(GLenum mode, GLsizei count, GLenum type, const void *indices, GLsizei instanceCount) {
	glDrawElementsInstanced(mode, count, type, indices, instanceCount);
}

GLint cglGetUniformLocation(GLuint program, const GLchar *name) {
	return glGetUniformLocation(program, name);
}

void cglUniformMatrix4fv(GLint location, GLsizei count, GLboolean transpose, const GLfloat *value) {
	glUniformMatrix4fv(location, count, transpose, value);
}

void cglUniform3fv(GLint location, GLsizei count, const GLfloat *value) {
	glUniform3fv(location, count, value);
}

void cglUniform1f(GLint location, GLfloat value) {
	glUniform1f(location, value);
}

void cglEnable(GLenum capability) {
	glEnable(capability);
}

void cglDisable(GLenum capability) {
	glDisable(capability);
}

void cglDepthMask(GLboolean flag) {
	glDepthMask(flag);
}

void cglDepthFunc(GLenum fun) {
	glDepthFunc(fun);
}

void cglBlendFunc(GLenum sfactor, GLenum dfactor) {
	glBlendFunc(sfactor, dfactor);
}

void cglBlendEquation(GLenum mode) {
	glBlendEquation(mode);
}

void cglFrontFace(GLenum mode) {
	glFrontFace(mode);
}

GLenum cglCheckFramebufferStatus(GLenum target) {
	return glCheckFramebufferStatus(target);
}

void cglDrawBuffers(GLsizei n, const GLenum *bufs) {
	glDrawBuffers(n, bufs);
}

void cglClearBufferfv(GLenum buffer, GLint drawBuffer, const GLfloat *value) {
	glClearBufferfv(buffer, drawBuffer, value);
}

void cglViewport(GLint x, GLint y, GLsizei width, GLsizei height) {
	glViewport(x, y, width, height);
}
*/
import "C"
import (
	_ "kaiju/gl/dist"
	"kaiju/matrix"
	"unsafe"
)

type Handle uint32
type Texture = Handle

func (h Handle) IsValid() bool {
	return h != 0
}

func (h Handle) AsGL() C.GLuint {
	return C.GLuint(h)
}

type Result int32

func (r Result) IsOkay() bool {
	return r != 0
}

func (r Result) Equal(value int32) bool {
	return int32(r) == value
}

const (
	CompileStatus           = 0x8B81
	ShaderType              = 0x8B4F
	FragmentShader          = 0x8B30
	VertexShader            = 0x8B31
	GeometryShader          = 0x8DD9
	TessControlShader       = 0x8E88
	TessEvaluationShader    = 0x8E87
	InfoLogLength           = 0x8B84
	LinkStatus              = 0x8B82
	ArrayBuffer             = 0x8892
	ElementArrayBuffer      = 0x8893
	StaticDraw              = 0x88E4
	Float                   = 0x1406
	HalfFloat               = 0x140B
	Int                     = 0x1404
	UnsignedInt             = 0x1405
	UnsignedByte            = 0x1401
	Points                  = 0x0000
	Lines                   = 0x0001
	Triangles               = 0x0004
	Texture2D               = 0x0DE1
	FrameBuffer             = 0x8D40
	FrameBufferComplete     = 0x8CD5
	Red                     = 0x1903
	R32F                    = 0x822E
	RGBA32F                 = 0x8814
	RGB                     = 0x1907
	RGBA                    = 0x1908
	RGBA8                   = 0x8058
	RGB8                    = 0x8051
	RGBA16F                 = 0x881A
	TextureWrapS            = 0x2802
	TextureWrapT            = 0x2803
	TextureMinFilter        = 0x2801
	TextureMagFilter        = 0x2800
	ClampToEdge             = 0x812F
	ColorAttachment0        = 0x8CE0
	ColorAttachment1        = 0x8CE1
	DepthAttachment         = 0x8D00
	Nearest                 = 0x2600
	Texture0                = 0x84C0
	Texture1                = 0x84C1
	Repeat                  = 0x2901
	Linear                  = 0x2601
	LinearMipMapLinear      = 0x2703
	CompressedRgbaAstc4x4   = 0x93B0
	CompressedRgbaAstc5x4   = 0x93B1
	CompressedRgbaAstc5x5   = 0x93B2
	CompressedRgbaAstc6x5   = 0x93B3
	CompressedRgbaAstc6x6   = 0x93B4
	CompressedRgbaAstc8x5   = 0x93B5
	CompressedRgbaAstc8x6   = 0x93B6
	CompressedRgbaAstc8x8   = 0x93B7
	CompressedRgbaAstc10x5  = 0x93B8
	CompressedRgbaAstc10x6  = 0x93B9
	CompressedRgbaAstc10x8  = 0x93BA
	CompressedRgbaAstc10x10 = 0x93BB
	CompressedRgbaAstc12x10 = 0x93BC
	CompressedRgbaAstc12x12 = 0x93BD
	CullFace                = 0x0B44
	DepthTest               = 0x0B71
	Zero                    = 0x0000
	One                     = 0x0001
	Less                    = 0x0201
	LEqual                  = 0x0203
	FuncAdd                 = 0x8006
	StencilTest             = 0x0B90
	Blend                   = 0x0BE2
	SrcAlpha                = 0x0302
	Always                  = 0x0207
	OneMinusSrcAlpha        = 0x0303
	CCW                     = 0x0901
	CW                      = 0x0900
	DepthComponent          = 0x1902
	DepthComponent32F       = 0x8CAC
	Color                   = 0x1800
	OneMinusSrcColor        = 0x0301
	ColorBufferBit          = 0x00004000
	DepthBufferBit          = 0x00000100
)

func ClearColor(r, g, b, a float32) {
	C.cglClearColor(C.GLfloat(r), C.GLfloat(g), C.GLfloat(b), C.GLfloat(a))
}

func Clear(mask Handle) {
	C.cglClear(C.GLbitfield(mask))
}

func Viewport(x, y, width, height int32) {
	C.cglViewport(C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}

func CreateShader(shaderType Handle) Handle {
	res := C.cglCreateShader(C.GLenum(shaderType))
	return Handle(res)
}

func ShaderSource(shader Handle, src string) {
	csrc := C.CString(src)
	defer C.free(unsafe.Pointer(csrc))
	C.cglShaderSource(shader.AsGL(), 1, &csrc, nil)
}

func CompileShader(shader Handle) {
	C.cglCompileShader(shader.AsGL())
}

func GetShaderParameter(shader Handle, param Handle) Result {
	var res C.GLint
	C.cglGetShaderiv(shader.AsGL(), C.GLenum(param), &res)
	return Result(res)
}

func GetShaderInfoLog(shader Handle) string {
	var length C.GLint
	C.cglGetShaderiv(shader.AsGL(), C.GLenum(InfoLogLength), &length)
	if length == 0 {
		return ""
	}
	log := make([]byte, length)
	C.cglGetShaderInfoLog(shader.AsGL(), C.GLsizei(length), nil, (*C.GLchar)(unsafe.Pointer(&log[0])))
	return string(log)
}

func CreateProgram() Handle {
	return Handle(C.cglCreateProgram())
}

func AttachShader(program Handle, shader Handle) {
	C.cglAttachShader(program.AsGL(), C.GLuint(shader))
}

func LinkProgram(program Handle) {
	C.cglLinkProgram(program.AsGL())
}

func GetProgramParameter(program Handle, param Handle) Result {
	var res C.GLint
	C.cglGetProgramiv(program.AsGL(), C.GLenum(param), &res)
	return Result(res)
}

func GetProgramInfoLog(program Handle) string {
	var length C.GLint
	C.cglGetProgramiv(program.AsGL(), C.GLenum(InfoLogLength), &length)
	if length == 0 {
		return ""
	}
	log := make([]byte, length)
	C.cglGetProgramInfoLog(program.AsGL(), C.GLsizei(length), nil, (*C.GLchar)(unsafe.Pointer(&log[0])))
	return string(log)
}

func DeleteShader(shader Handle) {
	C.cglDeleteShader(shader.AsGL())
}

func DeleteProgram(program Handle) {
	C.cglDeleteProgram(program.AsGL())
}

func DeleteTextures(n int32, textures *Handle) {
	C.cglDeleteTextures(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(textures)))
}

func DeleteFrameBuffers(n int32, framebuffers *Handle) {
	C.cglDeleteFramebuffers(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(framebuffers)))
}

func GenVertexArrays(n int32, arrays *Handle) {
	C.cglGenVertexArrays(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(arrays)))
}

func GenBuffers(n int32, buffers *Handle) {
	C.cglGenBuffers(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(buffers)))
}

func GenTextures(n int32, textures *Handle) {
	C.cglGenTextures(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(textures)))
}

func GenFrameBuffers(n int32, framebuffers *Handle) {
	C.cglGenFramebuffers(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(framebuffers)))
}

func BindVertexArray(array Handle) {
	C.cglBindVertexArray(array.AsGL())
}

func BindBuffer(target Handle, buffer Handle) {
	C.cglBindBuffer(C.GLenum(target), buffer.AsGL())
}

func BindTexture(target Handle, texture Handle) {
	C.cglBindTexture(C.GLenum(target), texture.AsGL())
}

func BindFrameBuffer(target Handle, framebuffer Handle) {
	C.cglBindFramebuffer(C.GLenum(target), framebuffer.AsGL())
}

func Uniform1i(location Result, value int32) {
	C.cglUniform1i(C.GLint(location), C.GLint(value))
}

func TexImage2D(target Handle, level int32, internalFormat Handle, width int32, height int32, border int32, format Handle, typ Handle, pixels unsafe.Pointer) {
	C.cglTexImage2D(C.GLenum(target), C.GLint(level), C.GLint(internalFormat), C.GLsizei(width), C.GLsizei(height), C.GLint(border), C.GLenum(format), C.GLenum(typ), pixels)
}

func CompressedTexImage2D(target Handle, level int32, internalFormat Handle, width int32, height int32, border int32, imageSize int32, data unsafe.Pointer) {
	C.cglCompressedTexImage2D(C.GLenum(target), C.GLint(level), C.GLint(internalFormat), C.GLsizei(width), C.GLsizei(height), C.GLint(border), C.GLsizei(imageSize), data)
}

func GenerateMipmap(target Handle) {
	C.cglGenerateMipmap(C.GLenum(target))
}

func GetTexImage(target Handle, level int32, format Handle, typ Handle, pixels unsafe.Pointer) {
	C.cglGetTexImage(C.GLenum(target), C.GLint(level), C.GLenum(format), C.GLenum(typ), pixels)
}

func FrameBufferTexture2D(target Handle, attachment Handle, textarget Handle, texture Handle, level int32) {
	C.cglFrameBufferTexture2D(C.GLenum(target), C.GLenum(attachment), C.GLenum(textarget), texture.AsGL(), C.GLint(level))
}

func TexParameteri(target Handle, pname Handle, param Handle) {
	C.cglTexParameteri(C.GLenum(target), C.GLenum(pname), C.GLint(param))
}

func BufferData(target Handle, data unsafe.Pointer, dataSize uint, usage Handle) {
	C.cglBufferData(C.GLenum(target), C.GLsizeiptr(dataSize), data, C.GLenum(usage))
}

func VertexAttribPointer(index uint32, size int32, typ Handle, normalized bool, stride int32, offset int32) {
	var nml uint8
	if normalized {
		nml = 1
	}
	C.cglVertexAttribPointer(C.GLuint(index), C.GLint(size), C.GLenum(typ), C.GLboolean(nml), C.GLsizei(stride), unsafe.Pointer(uintptr(offset)))
}

func EnableVertexAttribArray(index uint32) {
	C.cglEnableVertexAttribArray(C.GLuint(index))
}

func UseProgram(program Handle) {
	C.cglUseProgram(program.AsGL())
}

func DrawArrays(mode Handle, first int32, count int32) {
	C.cglDrawArrays(C.GLenum(mode), C.GLint(first), C.GLsizei(count))
}

func DrawElementsInstanced(mode Handle, count int32, typ Handle, first int32, instanceCount int32) {
	C.cglDrawElementsInstanced(C.GLenum(mode), C.GLsizei(count), C.GLenum(typ), unsafe.Pointer(uintptr(first)), C.GLsizei(instanceCount))
}

func UnBindBuffer(target Handle) {
	C.cglBindBuffer(C.GLenum(target), 0)
}

func UnBindVertexArray() {
	C.cglBindVertexArray(0)
}

func UnBindTexture(target Handle) {
	C.cglBindTexture(C.GLenum(target), 0)
}

func UnBindFrameBuffer(target Handle) {
	C.cglBindFramebuffer(C.GLenum(target), 0)
}

func ActivateTexture(target Handle) {
	C.cglActiveTexture(C.GLenum(target))
}

func GetUniformLocation(program Handle, name string) Result {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	res := C.cglGetUniformLocation(program.AsGL(), cname)
	return Result(res)
}

func UniformMatrix4fv(location Result, transpose bool, matrices *matrix.Mat4) {
	var nml uint8
	if transpose {
		nml = 1
	}
	C.cglUniformMatrix4fv(C.GLint(location), C.GLsizei(1), C.GLboolean(nml), (*C.GLfloat)(unsafe.Pointer(&matrices[0])))
}

func Uniform3fv(location Result, values *matrix.Vec3) {
	C.cglUniform3fv(C.GLint(location), C.GLsizei(1), (*C.GLfloat)(unsafe.Pointer(&values[0])))
}

func Uniform1f(location Result, value float32) {
	C.cglUniform1f(C.GLint(location), C.GLfloat(value))
}

func Enable(capability Handle) {
	C.cglEnable(C.GLenum(capability))
}

func Disable(capability Handle) {
	C.cglDisable(C.GLenum(capability))
}

func DepthMask(flag bool) {
	var nml uint8
	if flag {
		nml = 1
	}
	C.cglDepthMask(C.GLboolean(nml))
}

func DepthFunc(fun Handle) {
	C.cglDepthFunc(C.GLenum(fun))
}

func BlendFunc(src, dst Handle) {
	C.cglBlendFunc(C.GLenum(src), C.GLenum(dst))
}

func BlendEquation(mode Handle) {
	C.cglBlendEquation(C.GLenum(mode))
}

func FrontFace(mode Handle) {
	C.cglFrontFace(C.GLenum(mode))
}

func CheckFrameBufferStatus(target Handle) Result {
	res := C.cglCheckFramebufferStatus(C.GLenum(target))
	return Result(res)
}

func DrawBuffers(buffers []Handle) {
	C.cglDrawBuffers(C.GLsizei(len(buffers)), (*C.GLenum)(unsafe.Pointer(&buffers[0])))
}

func ClearBufferfv(buffer Handle, drawBuffer int32, value matrix.Vec4) {
	C.cglClearBufferfv(C.GLenum(buffer), C.GLint(drawBuffer), (*C.GLfloat)(unsafe.Pointer(&value[0])))
}
