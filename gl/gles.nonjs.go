//go:build !js && !wasm

package gl

/*
#include <stdlib.h>
#include "glad.h"

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

void cglClearScreen() {
	glClearColor(0.392f, 0.584f, 0.929f, 1.0f);
	glClear(GL_COLOR_BUFFER_BIT);
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

void cglGenVertexArrays(GLsizei n, GLuint *arrays) {
	glGenVertexArrays(n, arrays);
}

void cglGenBuffers(GLsizei n, GLuint *buffers) {
	glGenBuffers(n, buffers);
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

void cglActiveTexture(GLenum texture) {
	glActiveTexture(texture);
}

void cglUniform1i(GLint location, GLint value) {
	glUniform1i(location, value);
}

void cglTexImage2D(GLenum target, GLint level, GLint internalFormat, GLsizei width, GLsizei height, GLint border, GLenum format, GLenum type, const void *pixels) {
	glTexImage2D(target, level, internalFormat, width, height, border, format, type, pixels);
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
*/
import "C"
import (
	"kaiju/matrix"
	"unsafe"
)

type Handle uint32

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
	CompileStatus        = 0x8B81
	ShaderType           = 0x8B4F
	FragmentShader       = 0x8B30
	VertexShader         = 0x8B31
	GeometryShader       = 0x8DD9
	TessControlShader    = 0x8E88
	TessEvaluationShader = 0x8E87
	InfoLogLength        = 0x8B84
	LinkStatus           = 0x8B82
	ArrayBuffer          = 0x8892
	ElementArrayBuffer   = 0x8893
	StaticDraw           = 0x88E4
	Float                = 0x1406
	Int                  = 0x1404
	UnsignedInt          = 0x1405
	Triangles            = 0x0004
	Texture2D            = 0x0DE1
	RGBA32F              = 0x8814
	RGBA                 = 0x1908
	TextureWrapS         = 0x2802
	TextureWrapT         = 0x2803
	TextureMinFilter     = 0x2801
	TextureMagFilter     = 0x2800
	ClampToEdge          = 0x812F
	Nearest              = 0x2600
	Texture0             = 0x84C0
)

func ClearScreen() {
	C.cglClearScreen()
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

func GenVertexArrays(n int32, arrays *Handle) {
	C.cglGenVertexArrays(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(arrays)))
}

func GenBuffers(n int32, buffers *Handle) {
	C.cglGenBuffers(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(buffers)))
}

func GenTextures(n int32, textures *Handle) {
	C.cglGenTextures(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(textures)))
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

func Uniform1i(location Result, value int32) {
	C.cglUniform1i(C.GLint(location), C.GLint(value))
}

func TexImage2D(target Handle, level int32, internalFormat Handle, width int32, height int32, border int32, format Handle, typ Handle, pixels unsafe.Pointer) {
	C.cglTexImage2D(C.GLenum(target), C.GLint(level), C.GLint(internalFormat), C.GLsizei(width), C.GLsizei(height), C.GLint(border), C.GLenum(format), C.GLenum(typ), pixels)
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
