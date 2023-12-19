package rendering

import (
	"kaiju/gl"
	"kaiju/matrix"
	"slices"
	"unsafe"
)

type DrawInstance interface {
	Destroy()
	IsDestroyed() bool
	Activate()
	Deactivate()
	IsActive() bool
	Size() int
	SetModel(model matrix.Mat4)
	UpdateModel()
	DataPointer() unsafe.Pointer
}

const ShaderBaseDataStart = unsafe.Offsetof(ShaderDataBase{}.model)

type ShaderDataBase struct {
	destroyed   bool
	deactivated bool
	_           [2]byte
	transform   *matrix.Transform
	initModel   matrix.Mat4
	model       matrix.Mat4
}

func (s *ShaderDataBase) Destroy()          { s.destroyed = true }
func (s *ShaderDataBase) IsDestroyed() bool { return s.destroyed }
func (s *ShaderDataBase) Activate()         { s.deactivated = false }
func (s *ShaderDataBase) Deactivate()       { s.deactivated = true }
func (s *ShaderDataBase) IsActive() bool    { return !s.deactivated }

func (s *ShaderDataBase) SetModel(model matrix.Mat4) {
	s.initModel = model
	if s.transform == nil {
		s.model = model
	}
}

func (s *ShaderDataBase) UpdateModel() {
	if s.transform != nil && s.transform.IsDirty() {
		s.model = s.initModel.Multiply(s.transform.Matrix())
	}
}

func (s *ShaderDataBase) DataPointer() unsafe.Pointer {
	return unsafe.Pointer(&s.model[0])
}

type DrawInstanceGroup struct {
	Mesh         *Mesh
	TextureData  gl.Handle
	Textures     []*Texture
	Instances    []DrawInstance
	instanceData []byte
	instanceSize int
}

func NewDrawInstanceGroup(mesh *Mesh, dataSize int) DrawInstanceGroup {
	return DrawInstanceGroup{
		Mesh:         mesh,
		Instances:    make([]DrawInstance, 0),
		instanceData: make([]byte, 0),
		instanceSize: dataSize,
	}
}

func (d *DrawInstanceGroup) IsEmpty() bool {
	return len(d.Instances) == 0
}

func (d *DrawInstanceGroup) AddInstance(instance DrawInstance) {
	d.Instances = append(d.Instances, instance)
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

func (d *DrawInstanceGroup) texSize() (int32, int32) {
	// Low end devices have a max 2048 texture size
	pixelCount := int32(len(d.instanceData)) / 4 / 4
	width := min(pixelCount, 2048)
	height := int32(1)
	for pixelCount > 2048 {
		height++
		pixelCount -= 2048
	}
	if height > 2048 {
		// TODO:  Handle this case with multiple textures
		panic("Too many instances")
	}
	return width, height
}

func (d *DrawInstanceGroup) UpdateData() {
	base := unsafe.Pointer(&d.instanceData[0])
	offset := uintptr(0)
	for i := 0; i < len(d.Instances); i++ {
		instance := d.Instances[i]
		instance.UpdateModel()
		if instance.IsDestroyed() {
			d.Instances = slices.Delete(d.Instances, i, i+1)
			i--
		} else if instance.IsActive() {
			to := unsafe.Pointer(uintptr(base) + offset)
			from := instance.DataPointer()
			copy(unsafe.Slice((*byte)(to), d.instanceSize),
				unsafe.Slice((*byte)(from), d.instanceSize))
			offset += uintptr(d.instanceSize)
		}
	}
	gl.BindTexture(gl.Texture2D, d.TextureData)
	w, h := d.texSize()
	gl.TexImage2D(gl.Texture2D, 0, gl.RGBA32F, w, h, 0,
		gl.RGBA, gl.Float, unsafe.Pointer(&d.instanceData[0]))
	gl.UnBindTexture(gl.Texture2D)
}
