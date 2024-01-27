//go:build OPENGL

package rendering

import (
	"kaiju/assets"
	"kaiju/cameras"
	"kaiju/gl"
	"kaiju/matrix"
	"log"
	"strings"
	"unsafe"
)

type MeshIdGL struct {
	VAO        gl.Handle
	VBO        gl.Handle
	EBO        gl.Handle
	indexCount int32
}

func (m MeshIdGL) IsValid() bool {
	return m.VAO.IsValid()
}

type GLRenderer struct {
	globalShaderData     GlobalShaderData
	opaqueFBO            gl.Handle
	transparentAccumFBO  gl.Handle
	transparentRevealFBO gl.Handle
	opaqueTexture        gl.Texture
	depthTexture         gl.Texture
	accumTexture         gl.Texture
	revealTexture        gl.Texture
	revealAccumTexture   gl.Texture
	revealRevealTexture  gl.Texture
	colorBuffer          gl.Handle
	compositeShader      *Shader
	hdrShader            *Shader
	composeQuad          *Mesh
	hdr                  int
	exposure             float32
}

func NewGLRenderer() *GLRenderer {
	r := &GLRenderer{}
	gl.Enable(gl.CullFace)
	gl.Enable(gl.DepthTest)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LEqual)
	gl.Disable(gl.StencilTest)
	gl.Enable(gl.Blend)
	gl.BlendFunc(gl.SrcAlpha, gl.OneMinusSrcAlpha)
	gl.FrontFace(gl.CCW)
	// TODO:  For when doing WebGL stuff
	//gl.GetExtension("EXT_color_buffer_half_float")
	//gl.GetExtension("EXT_float_blend")
	//gl.GetExtension("EXT_color_buffer_float")
	//gl.GetExtension("OES_texture_float_linear")
	return r
}

func (r *GLRenderer) Initialize(caches RenderCaches, width, height int32) error {
	r.setupOITFrameBuffer(width, height)
	r.composeQuad = NewMeshUnitQuad(caches.MeshCache())
	r.compositeShader = caches.ShaderCache().Shader(
		assets.ShaderOitCompositeVert, assets.ShaderOitCompositeFrag, "", "", "")
	r.hdrShader = caches.ShaderCache().Shader(
		assets.ShaderHdrVert, assets.ShaderHdrFrag, "", "", "")
	return nil
}

func createShaderObject(assetDatabase *assets.Database, shaderKey string, shaderType gl.Handle, defines []string) (gl.Handle, string) {
	src, err := assetDatabase.ReadText(shaderKey)
	if err != nil {
		panic(err)
	}
	// TODO:  Setup this so it supports other versions
	const vulkanVersion = "#version 460"
	const glVersion = "#version 300 es\nprecision mediump float;"
	if strings.HasPrefix(src, vulkanVersion) {
		src = strings.Replace(src, vulkanVersion, glVersion, 1)
	}
	if len(defines) > 0 {
		defineStr := glVersion + "\n#define " + strings.Join(defines, "\n#define ")
		src = strings.Replace(src, glVersion, defineStr, 1)
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
	return shaderObj, src
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

func (r *GLRenderer) CreateShader(shader *Shader, assetDatabase *assets.Database) {
	noDef := []string{}
	vert, _ := createShaderObject(assetDatabase, shader.VertPath, gl.VertexShader, noDef)
	frag, fragSrc := createShaderObject(assetDatabase, shader.FragPath, gl.FragmentShader, shader.DriverData.Defines)
	var geom, tesc, tese gl.Handle
	if len(shader.GeomPath) > 0 {
		geom, _ = createShaderObject(assetDatabase, shader.GeomPath, gl.GeometryShader, noDef)
	}
	if len(shader.CtrlPath) > 0 {
		tesc, _ = createShaderObject(assetDatabase, shader.CtrlPath, gl.TessControlShader, noDef)
	}
	if len(shader.EvalPath) > 0 {
		tese, _ = createShaderObject(assetDatabase, shader.EvalPath, gl.TessEvaluationShader, noDef)
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
	const def = "#define OIT"
	const ifdef = "#ifdef OIT"
	if !strings.Contains(fragSrc, def) && strings.Contains(fragSrc, ifdef) {
		shader.SubShader = NewShader(shader.VertPath, shader.FragPath,
			shader.GeomPath, shader.CtrlPath, shader.EvalPath, r)
		shader.SubShader.DriverData.Defines = append(
			shader.SubShader.DriverData.Defines, "OIT")
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

func (r *GLRenderer) CreateTexture(texture *Texture, textureData *TextureData) {
	var id gl.Handle
	gl.GenTextures(1, &id)
	texture.RenderId = id
	gl.BindTexture(gl.Texture2D, id)
	gl.TexParameteri(gl.Texture2D, gl.TextureWrapS, gl.Repeat)
	gl.TexParameteri(gl.Texture2D, gl.TextureWrapT, gl.Repeat)
	gl.TexParameteri(gl.Texture2D, gl.TextureMinFilter, gl.LinearMipMapLinear)
	gl.TexParameteri(gl.Texture2D, gl.TextureMagFilter, gl.Linear)
	if texture.pendingData.InputType == TextureFileFormatAstc {
		gl.CompressedTexImage2D(gl.Texture2D, 0,
			toGLInternalFormat(texture.pendingData.InternalFormat),
			int32(texture.pendingData.Width), int32(texture.pendingData.Height), 0,
			int32(len(texture.pendingData.Mem)), unsafe.Pointer(&texture.pendingData.Mem[0]))
	} else {
		gl.TexImage2D(gl.Texture2D, 0, toGLInternalFormat(texture.pendingData.InternalFormat),
			int32(texture.pendingData.Width), int32(texture.pendingData.Height), 0,
			toGLFormat(texture.pendingData.Format),
			toGLType(texture.pendingData.Type), unsafe.Pointer(&texture.pendingData.Mem[0]))
	}
	gl.GenerateMipmap(gl.Texture2D)
	gl.UnBindTexture(gl.Texture2D)
}

func (w *GLRenderer) TextureReadPixel(texture *Texture, x, y int) matrix.Color {
	if texture.TexturePixelCache == nil {
		texture.TexturePixelCache = make([]uint8, texture.Width*texture.Height*bytesInPixel)
	}
	if texture.CacheInvalid {
		gl.GetTexImage(gl.Texture2D, 0, gl.RGBA, gl.UnsignedByte, unsafe.Pointer(&texture.TexturePixelCache[0]))
	}
	offset := (y*texture.Width + x) * bytesInPixel
	return matrix.Color{
		float32(texture.TexturePixelCache[offset]),
		float32(texture.TexturePixelCache[offset+1]),
		float32(texture.TexturePixelCache[offset+2]),
		float32(texture.TexturePixelCache[offset+3]),
	}
}

func (w *GLRenderer) TextureWritePixels(texture *Texture, x, y, width, height int, pixels []byte) {
	panic("TextureWritePixels not implemented")
}

func (r *GLRenderer) ReadyFrame(camera *cameras.StandardCamera, uiCamera *cameras.StandardCamera, runtime float32) bool {
	r.globalShaderData.View = camera.View()
	r.globalShaderData.Projection = camera.Projection()
	r.globalShaderData.UIView = uiCamera.View()
	r.globalShaderData.UIProjection = uiCamera.Projection()
	r.globalShaderData.CameraPosition = camera.Position()
	r.globalShaderData.UICameraPosition = uiCamera.Position()
	r.globalShaderData.Time = runtime
	return true
}

func (r GLRenderer) setGlobalUniforms(shader *Shader) {
	sid := shader.RenderId.(gl.Handle)
	viewLoc := gl.GetUniformLocation(sid, "globalData.view")
	projectionLoc := gl.GetUniformLocation(sid, "globalData.projection")
	uiViewLoc := gl.GetUniformLocation(sid, "globalData.uiView")
	uiProjectionLoc := gl.GetUniformLocation(sid, "globalData.uiProjection")
	cameraPositionLoc := gl.GetUniformLocation(sid, "globalData.cameraPosition")
	uiCameraPositionLoc := gl.GetUniformLocation(sid, "globalData.uiCameraPosition")
	timeLoc := gl.GetUniformLocation(sid, "globalData.time")
	gl.UniformMatrix4fv(viewLoc, false, &r.globalShaderData.View)
	gl.UniformMatrix4fv(projectionLoc, false, &r.globalShaderData.Projection)
	gl.UniformMatrix4fv(uiViewLoc, false, &r.globalShaderData.UIView)
	gl.UniformMatrix4fv(uiProjectionLoc, false, &r.globalShaderData.UIProjection)
	gl.Uniform3fv(cameraPositionLoc, &r.globalShaderData.CameraPosition)
	gl.Uniform3fv(uiCameraPositionLoc, &r.globalShaderData.UICameraPosition)
	gl.Uniform1f(timeLoc, r.globalShaderData.Time)
}

func (r *GLRenderer) draw(drawings []ShaderDraw) {
	for _, sd := range drawings {
		shaderId := sd.shader.RenderId.(gl.Handle)
		gl.UseProgram(shaderId)
		r.setGlobalUniforms(sd.shader)
		for _, draw := range sd.instanceGroups {
			if draw.IsEmpty() || !draw.Mesh.IsReady() {
				continue
			}
			draw.UpdateData()
			meshId := draw.Mesh.MeshId.(MeshIdGL)
			gl.BindVertexArray(meshId.VAO)
			gl.ActivateTexture(gl.Texture0)
			gl.BindTexture(gl.Texture2D, draw.InstanceDriverData)
			gl.Uniform1i(gl.GetUniformLocation(shaderId, "instanceSampler"), 0)
			for i, texture := range draw.Textures {
				gl.ActivateTexture(gl.Handle(int(gl.Texture1) + i))
				gl.BindTexture(gl.Texture2D, texture.RenderId.(gl.Handle))
				// TODO:  Set/get the uniform location as part of draw textures
				gl.Uniform1i(gl.GetUniformLocation(shaderId, "texSampler"), int32(i+1))
			}
			gl.BindBuffer(gl.ElementArrayBuffer, meshId.EBO)
			gl.DrawElementsInstanced(gl.Triangles, meshId.indexCount,
				gl.UnsignedInt, 0, int32(len(draw.Instances)))
			gl.UnBindBuffer(gl.ElementArrayBuffer)
			gl.UnBindTexture(gl.Texture2D)
			gl.UnBindVertexArray()
		}
	}
}

func (r GLRenderer) Draw(drawings []ShaderDraw) {
	solids := make([]ShaderDraw, 0)
	transparents := make([]ShaderDraw, 0)
	for _, sd := range drawings {
		ss := NewShaderDraw(sd.shader)
		var st ShaderDraw
		if sd.shader.SubShader != nil {
			st = NewShaderDraw(sd.shader.SubShader)
		} else {
			st = NewShaderDraw(sd.shader)
		}
		for _, dg := range sd.instanceGroups {
			if dg.useBlending {
				st.AddInstanceGroup(&dg)
			} else {
				ss.AddInstanceGroup(&dg)
			}
		}
		if len(ss.instanceGroups) > 0 {
			solids = append(solids, ss)
		}
		if len(st.instanceGroups) > 0 {
			transparents = append(transparents, st)
		}
	}
	r.solidPass(solids, matrix.ColorCornflowerBlue())
	r.transparentPass(transparents)
}

func (r *GLRenderer) SwapFrame(width, height int32) bool {
	r.composePass()
	gl.Disable(gl.DepthTest)
	gl.DepthMask(true)
	gl.Disable(gl.Blend)
	gl.UnBindFrameBuffer(gl.FrameBuffer)
	gl.Viewport(0, 0, width, height)
	gl.ClearColor(1, 0, 1, 1)
	gl.Clear(gl.ColorBufferBit | gl.DepthBufferBit | gl.StencilTest)
	id := r.hdrShader.RenderId.(gl.Handle)
	gl.UseProgram(id)
	meshId := r.composeQuad.MeshId.(MeshIdGL)
	gl.BindVertexArray(meshId.VAO)
	gl.Uniform1i(gl.GetUniformLocation(id, "hdr"), int32(r.hdr))
	gl.Uniform1f(gl.GetUniformLocation(id, "exposure"), r.exposure)
	gl.ActivateTexture(gl.Texture0)
	gl.BindTexture(gl.Texture2D, r.opaqueTexture)
	gl.BindBuffer(gl.ElementArrayBuffer, meshId.EBO)
	gl.DrawElementsInstanced(gl.Triangles, 6, gl.UnsignedInt, 0, 1)
	gl.UnBindBuffer(gl.ElementArrayBuffer)
	gl.UnBindTexture(gl.Texture2D)
	gl.UnBindVertexArray()
	return true
}

func (r *GLRenderer) Resize(width, height int) {
	gl.Viewport(0, 0, width, height)
}
