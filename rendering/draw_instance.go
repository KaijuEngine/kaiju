package rendering

import (
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
	Shader       *Shader
	Mesh         *Mesh
	Instances    []DrawInstance
	instanceData []byte
	instanceSize int
	dataSize     int
}

func NewDrawInstanceGroup(shader *Shader, mesh *Mesh, dataSize int) DrawInstanceGroup {
	return DrawInstanceGroup{
		Shader:       shader,
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
}

func (d *DrawInstanceGroup) UpdateData() {
	base := unsafe.Pointer(&d.instanceData[0])
	for i, instance := range d.Instances {
		to := unsafe.Pointer(uintptr(base) + uintptr(i*d.instanceSize))
		from := instance.DataPointer()
		copy(unsafe.Slice((*byte)(to), d.instanceSize),
			unsafe.Slice((*byte)(from), d.instanceSize))
	}
}
