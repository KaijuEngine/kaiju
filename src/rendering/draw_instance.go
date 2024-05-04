/******************************************************************************/
/* draw_instance.go                                                           */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"encoding/gob"
	"kaiju/klib"
	"kaiju/matrix"
	"unsafe"
)

func init() {
	gob.Register(&ShaderDataBasic{})
}

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
	InitModel   matrix.Mat4
	model       matrix.Mat4
}

type ShaderDataBasic struct {
	ShaderDataBase
	Color matrix.Color
}

func (t ShaderDataBasic) Size() int {
	return int(unsafe.Sizeof(ShaderDataBasic{}) - ShaderBaseDataStart)
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
	s.InitModel = model
	if s.transform == nil {
		s.model = model
	}
}

func (s *ShaderDataBase) UpdateModel() {
	if s.transform != nil && s.transform.IsDirty() {
		s.model = matrix.Mat4Multiply(s.InitModel, s.transform.WorldMatrix())
	}
}

func (s *ShaderDataBase) DataPointer() unsafe.Pointer {
	return unsafe.Pointer(&s.model[0])
}

type DrawInstanceGroup struct {
	Mesh *Mesh
	InstanceDriverData
	Textures          []*Texture
	Instances         []DrawInstance
	instanceData      []byte
	namedInstanceData map[string][]byte
	instanceSize      int
	visibleCount      int
	padding           int
	useBlending       bool
	destroyed         bool
}

func NewDrawInstanceGroup(mesh *Mesh, dataSize int) DrawInstanceGroup {
	return DrawInstanceGroup{
		Mesh:              mesh,
		Instances:         make([]DrawInstance, 0),
		instanceData:      make([]byte, 0),
		namedInstanceData: make(map[string][]byte),
		instanceSize:      dataSize,
		padding:           dataSize % 16,
		destroyed:         false,
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

func (d *DrawInstanceGroup) AnyVisible() bool  { return d.visibleCount > 0 }
func (d *DrawInstanceGroup) VisibleCount() int { return d.visibleCount }

func (d *DrawInstanceGroup) VisibleSize() int {
	return d.visibleCount * (d.instanceSize + d.padding)
}

func (d *DrawInstanceGroup) UpdateData(renderer Renderer) {
	base := unsafe.Pointer(&d.instanceData[0])
	offset := uintptr(0)
	count := len(d.Instances)
	d.visibleCount = 0
	for i := 0; i < count; i++ {
		instance := d.Instances[i]
		instance.UpdateModel()
		if instance.IsDestroyed() {
			d.Instances[i] = d.Instances[count-1]
			i--
			count--
		} else if instance.IsActive() {
			to := unsafe.Pointer(uintptr(base) + offset)
			klib.Memcpy(to, instance.DataPointer(), uint64(d.instanceSize))
			offset += uintptr(d.instanceSize + d.padding)
			d.visibleCount++
		}
	}
	if count < len(d.Instances) {
		newMemLen := count * (d.instanceSize + d.padding)
		d.Instances = d.Instances[:count]
		d.instanceData = d.instanceData[:newMemLen]
	}
	d.bindInstanceDriverData()
	if len(d.Instances) == 0 {
		renderer.DestroyGroup(d)
		d.destroyed = true
	}
}

func (d *DrawInstanceGroup) Destroy(renderer Renderer) {
	if d.destroyed {
		return
	}
	for i := range d.Instances {
		d.Instances[i].Destroy()
	}
	d.Instances = d.Instances[:0]
	renderer.DestroyGroup(d)
}
