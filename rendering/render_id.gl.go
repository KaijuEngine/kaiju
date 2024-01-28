//go:build OPENGL

package rendering

import "kaiju/gl"

type ShaderId gl.Handle
type TextureId gl.Handle
type MeshId gl.Handle

func (m MeshId) IsValid() bool {
	return m != 0
}

type DriverData struct {
	Defines []string
}

func (d *ShaderDriverData) setup(def ShaderDef, _ uint32) {
}

func NewDriverData() DriverData {
	return DriverData{
		Defines: make([]string, 0),
	}
}
