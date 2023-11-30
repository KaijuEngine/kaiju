package rendering

import (
	"kaiju/assets"
	"kaiju/gl"
	"log"
	"unsafe"
)

type MeshIdGL struct {
	VAO        gl.Handle
	VBO        gl.Handle
	EBO        gl.Handle
	indexCount int32
}

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

func (r GLRenderer) CreateMesh(mesh *Mesh, verts []Vertex, indices []uint32) {
	id := MeshIdGL{}
	stride := int32(unsafe.Sizeof(verts[0]))
	vertsSize := uint(stride) * uint(len(verts))
	indexSize := uint(unsafe.Sizeof(indices[0])) * uint(len(indices))
	id.indexCount = int32(len(indices))
	gl.GenVertexArrays(1, &id.VAO)
	gl.GenBuffers(1, &id.VBO)
	gl.GenBuffers(1, &id.EBO)
	gl.BindVertexArray(id.VAO)

	gl.BindBuffer(gl.ArrayBuffer, id.VBO)
	gl.BufferData(gl.ArrayBuffer, unsafe.Pointer(&verts[0]), vertsSize, gl.StaticDraw)
	gl.BindBuffer(gl.ElementArrayBuffer, id.EBO)
	gl.BufferData(gl.ElementArrayBuffer, unsafe.Pointer(&indices[0]), indexSize, gl.StaticDraw)
	pOffset := int32(0)
	// Vertex positions
	gl.VertexAttribPointer(0, 3, gl.Float, false, stride, pOffset)
	gl.EnableVertexAttribArray(0)
	pOffset += int32(unsafe.Sizeof(verts[0].Position))
	// Vertex normals
	gl.VertexAttribPointer(1, 3, gl.Float, false, stride, pOffset)
	gl.EnableVertexAttribArray(1)
	pOffset += int32(unsafe.Sizeof(verts[0].Normal))
	// Vertex tangent
	gl.VertexAttribPointer(2, 4, gl.Float, false, stride, pOffset)
	gl.EnableVertexAttribArray(2)
	pOffset += int32(unsafe.Sizeof(verts[0].Tangent))
	// Vertex texture coordinates
	gl.VertexAttribPointer(3, 2, gl.Float, false, stride, pOffset)
	gl.EnableVertexAttribArray(3)
	pOffset += int32(unsafe.Sizeof(verts[0].UV0))
	// Vertex color
	gl.VertexAttribPointer(4, 4, gl.Float, false, stride, pOffset)
	gl.EnableVertexAttribArray(4)
	pOffset += int32(unsafe.Sizeof(verts[0].Color))
	// Vertex joint ids
	gl.VertexAttribPointer(5, 4, gl.Int, false, stride, pOffset)
	gl.EnableVertexAttribArray(5)
	pOffset += int32(unsafe.Sizeof(verts[0].JointIds))
	// Vertex joint weights
	gl.VertexAttribPointer(6, 4, gl.Float, false, stride, pOffset)
	gl.EnableVertexAttribArray(6)
	pOffset += int32(unsafe.Sizeof(verts[0].JointWeights))
	// Vertex morph target
	gl.VertexAttribPointer(7, 3, gl.Float, false, stride, pOffset)
	gl.EnableVertexAttribArray(7)
	// Unbind
	gl.UnBindBuffer(gl.ArrayBuffer)
	gl.UnBindVertexArray()
	mesh.MeshId = id
}
