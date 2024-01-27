//go:build OPENGL

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
