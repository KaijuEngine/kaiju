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

void cglGenVertexArrays(GLsizei n, GLuint *arrays) {
	glGenVertexArrays(n, arrays);
}

void cglGenBuffers(GLsizei n, GLuint *buffers) {
	glGenBuffers(n, buffers);
}

void cglBindVertexArray(GLuint array) {
	glBindVertexArray(array);
}

void cglBindBuffer(GLenum target, GLuint buffer) {
	glBindBuffer(target, buffer);
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
*/
import "C"
import "unsafe"

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
	Triangles            = 0x0004
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

func GenVertexArrays(n int32, arrays *Handle) {
	C.cglGenVertexArrays(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(arrays)))
}

func GenBuffers(n int32, buffers *Handle) {
	C.cglGenBuffers(C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(buffers)))
}

func BindVertexArray(array Handle) {
	C.cglBindVertexArray(array.AsGL())
}

func BindBuffer(target Handle, buffer Handle) {
	C.cglBindBuffer(C.GLenum(target), buffer.AsGL())
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

func UnBindBuffer(target Handle) {
	C.cglBindBuffer(C.GLenum(target), 0)
}

func UnBindVertexArray() {
	C.cglBindVertexArray(0)
}
