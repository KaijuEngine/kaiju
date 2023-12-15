package rendering

import (
	"kaiju/gl"
	"kaiju/matrix"
	"unsafe"
)

type DrawInstance interface {
	DataPointer() unsafe.Pointer
}

type ShaderDataBase struct {
	Model matrix.Mat4
}

func (s *ShaderDataBase) DataPointer() unsafe.Pointer {
	return unsafe.Pointer(&s.Model[0])
}

type DrawInstanceGroup struct {
	Mesh         *Mesh
	TextureData  gl.Handle
	Textures     []*Texture
	Instances    []DrawInstance
	instanceData []byte
	instanceSize int
	dataSize     int
}

func NewDrawInstanceGroup(mesh *Mesh, dataSize int) DrawInstanceGroup {
	return DrawInstanceGroup{
		Mesh:         mesh,
		Instances:    make([]DrawInstance, 0),
		instanceData: make([]byte, 0),
		instanceSize: dataSize,
		dataSize:     0,
	}
}

func (d *DrawInstanceGroup) IsEmpty() bool {
	return len(d.Instances) == 0
}

func (d *DrawInstanceGroup) AddInstance(instance DrawInstance) {
	d.Instances = append(d.Instances, instance)
	d.dataSize += d.instanceSize
	d.instanceData = append(d.instanceData, make([]byte, d.instanceSize)...)
	d.generateTexture()
}

func (d *DrawInstanceGroup) generateTexture() {
	gl.DeleteTextures(1, &d.TextureData)
	gl.GenTextures(1, &d.TextureData)
	gl.BindTexture(gl.Texture2D, d.TextureData)
	gl.TexParameteri(gl.Texture2D, gl.TextureWrapS, gl.ClampToEdge)
	gl.TexParameteri(gl.Texture2D, gl.TextureWrapT, gl.ClampToEdge)
	gl.TexParameteri(gl.Texture2D, gl.TextureMinFilter, gl.Nearest)
	gl.TexParameteri(gl.Texture2D, gl.TextureMagFilter, gl.Nearest)
	gl.UnBindTexture(gl.Texture2D)
}

func (d *DrawInstanceGroup) UpdateData() {
	base := unsafe.Pointer(&d.instanceData[0])
	for i, instance := range d.Instances {
		to := unsafe.Pointer(uintptr(base) + uintptr(i*d.instanceSize))
		from := instance.DataPointer()
		copy(unsafe.Slice((*byte)(to), d.instanceSize),
			unsafe.Slice((*byte)(from), d.instanceSize))
	}
	gl.BindTexture(gl.Texture2D, d.TextureData)
	gl.TexImage2D(gl.Texture2D, 0, gl.RGBA32F,
		int32(len(d.instanceData))/4/4, 1, 0,
		gl.RGBA, gl.Float, unsafe.Pointer(&d.instanceData[0]))
	gl.UnBindTexture(gl.Texture2D)
}
