package rendering

import (
	"kaiju/matrix"
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
	setTransform(transform *matrix.Transform)
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

func NewShaderDataBase() ShaderDataBase {
	sdb := ShaderDataBase{}
	sdb.SetModel(matrix.Mat4Identity())
	return sdb
}

func (s *ShaderDataBase) Size() int {
	return int(unsafe.Sizeof(*s) - ShaderBaseDataStart)
}

func (s *ShaderDataBase) Destroy()          { s.destroyed = true }
func (s *ShaderDataBase) CancelDestroy()    { s.destroyed = false }
func (s *ShaderDataBase) IsDestroyed() bool { return s.destroyed }
func (s *ShaderDataBase) Activate()         { s.deactivated = false }
func (s *ShaderDataBase) Deactivate()       { s.deactivated = true }
func (s *ShaderDataBase) IsActive() bool    { return !s.deactivated }

func (s *ShaderDataBase) setTransform(transform *matrix.Transform) {
	s.transform = transform
}

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
	Mesh *Mesh
	InstanceDriverData
	Textures     []*Texture
	Instances    []DrawInstance
	instanceData []byte
	instanceSize int
	padding      int
	useBlending  bool
}

func NewDrawInstanceGroup(mesh *Mesh, dataSize int) DrawInstanceGroup {
	return DrawInstanceGroup{
		Mesh:         mesh,
		Instances:    make([]DrawInstance, 0),
		instanceData: make([]byte, 0),
		instanceSize: dataSize,
		padding:      dataSize % 16,
	}
}

func (d *DrawInstanceGroup) AlterPadding(blockSize int) {
	newPadding := blockSize - d.instanceSize%blockSize
	if d.padding != newPadding {
		d.padding = newPadding
		d.instanceData = make([]byte, d.TotalSize())
	}
}

func (d *DrawInstanceGroup) IsEmpty() bool {
	return len(d.Instances) == 0
}

func (d *DrawInstanceGroup) IsReady() bool {
	// TODO:  Check if textures are ready?
	return d.Mesh.IsReady() && !d.IsEmpty()
}

func (d *DrawInstanceGroup) TotalSize() int {
	return len(d.Instances) * (d.instanceSize + d.padding)
}

func (d *DrawInstanceGroup) AddInstance(instance DrawInstance, renderer Renderer, shader *Shader) {
	d.Instances = append(d.Instances, instance)
	d.instanceData = append(d.instanceData, make([]byte, d.instanceSize+d.padding)...)
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
	count := len(d.Instances)
	for i := 0; i < count; i++ {
		instance := d.Instances[i]
		instance.UpdateModel()
		if instance.IsDestroyed() {
			d.Instances[i] = d.Instances[count-1]
			i--
			count--
		} else if instance.IsActive() {
			to := unsafe.Pointer(uintptr(base) + offset)
			from := instance.DataPointer()
			copy(unsafe.Slice((*byte)(to), d.instanceSize),
				unsafe.Slice((*byte)(from), d.instanceSize))
			offset += uintptr(d.instanceSize + d.padding)
		}
	}
	if count < len(d.Instances) {
		d.Instances = d.Instances[:count]
	}
	d.bindInstanceDriverData()
}
