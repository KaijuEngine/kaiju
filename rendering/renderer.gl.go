package rendering

import (
	"kaiju/assets"
	"kaiju/gl"
	"log"
)

type GLRenderer struct {
}

func NewGLRenderer() GLRenderer {
	return GLRenderer{}
}

func createShaderObject(assetDatabase *assets.Database, shaderKey string, shaderType gl.Handle) gl.Handle {
	src, err := assetDatabase.ReadAsset(shaderKey)
	if err != nil {
		panic(err)
	}
	shaderObj := gl.CreateShader(shaderType)
	gl.ShaderSource(shaderObj, src)
	gl.CompileShader(shaderObj)
	res := gl.GetShaderParameter(shaderObj, gl.CompileStatus)
	if !res.IsOkay() {
		res = gl.GetShaderParameter(shaderObj, gl.ShaderType)
		var sType string
		if res.Equal(gl.FragmentShader) {
			sType = "Fragment"
		} else if res.Equal(gl.VertexShader) {
			sType = "Vertex"
		} else if res.Equal(gl.GeometryShader) {
			sType = "Geometry"
		} else if res.Equal(gl.TessControlShader) {
			sType = "Tessellation Control"
		} else if res.Equal(gl.TessEvaluationShader) {
			sType = "Tessellation Evaluation"
		} else {
			sType = "Unknown"
		}
		errLog := gl.GetShaderInfoLog(shaderObj)
		if len(errLog) > 0 {
			log.Fatalf("Error compiling shader %s: %s\n", sType, errLog)
		} else {
			log.Fatalf("Error compiling shader %s: There was an error compiling the shader and could not retrieve the error log for unknown reasons\n", sType)
		}
	}
	return shaderObj
}

func linkShader(vert, frag, geom, tesc, tese gl.Handle) gl.Handle {
	shader := gl.CreateProgram()
	gl.AttachShader(shader, vert)
	gl.AttachShader(shader, frag)
	if geom.IsValid() {
		gl.AttachShader(shader, geom)
	}
	if tesc.IsValid() {
		gl.AttachShader(shader, tesc)
	}
	if tese.IsValid() {
		gl.AttachShader(shader, tese)
	}
	gl.LinkProgram(shader)
	// Check for linker errors
	res := gl.GetProgramParameter(shader, gl.LinkStatus)
	if !res.IsOkay() {
		errLog := gl.GetProgramInfoLog(shader)
		log.Fatalf("Error linking shader: %s\n", errLog)
	}
	return shader
}

func (r GLRenderer) CreateShader(shader *Shader, assetDatabase *assets.Database) {
	vert := createShaderObject(assetDatabase, shader.VertPath, gl.VertexShader)
	frag := createShaderObject(assetDatabase, shader.FragPath, gl.FragmentShader)
	var geom, tesc, tese gl.Handle
	if len(shader.GeomPath) > 0 {
		geom = createShaderObject(assetDatabase, shader.GeomPath, gl.GeometryShader)
	}
	if len(shader.CtrlPath) > 0 {
		tesc = createShaderObject(assetDatabase, shader.CtrlPath, gl.TessControlShader)
	}
	if len(shader.EvalPath) > 0 {
		tese = createShaderObject(assetDatabase, shader.EvalPath, gl.TessEvaluationShader)
	}
	shader.RenderId = linkShader(vert, frag, geom, tesc, tese)
	gl.DeleteShader(vert)
	gl.DeleteShader(frag)
	if geom.IsValid() {
		gl.DeleteShader(geom)
	}
	if tesc.IsValid() {
		gl.DeleteShader(tesc)
	}
	if tese.IsValid() {
		gl.DeleteShader(tese)
	}
}

func (r GLRenderer) FreeShader(shader *Shader) {
	gl.DeleteProgram(shader.RenderId.(gl.Handle))
}
