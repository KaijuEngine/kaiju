/******************************************************************************/
/* draw_instance.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"kaiju/engine/collision"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"reflect"
	"unsafe"
)

type ViewCuller interface {
	IsInView(box collision.AABB) bool
	ViewChanged() bool
}

type DrawInstance interface {
	Base() *ShaderDataBase
	SkinningHeader() *SkinnedShaderDataHeader
	Destroy()
	IsDestroyed() bool
	Activate()
	Deactivate()
	IsInView() bool
	Size() int
	SetModel(model matrix.Mat4)
	UpdateModel(viewCuller ViewCuller, container collision.AABB)
	DataPointer() unsafe.Pointer
	// Returns true if it should write the data, otherwise false
	UpdateBoundData() bool
	BoundDataPointer() unsafe.Pointer
	InstanceBoundDataSize() int
	setTransform(transform *matrix.Transform)
	SelectLights(lights LightsForRender)
	setShadow(shadow DrawInstance)
	renderBounds() collision.AABB
}

func ReflectDuplicateDrawInstance(target DrawInstance) DrawInstance {
	val := reflect.ValueOf(target)
	if !val.IsValid() {
		return nil
	}
	newVal := reflect.New(val.Elem().Type()).Elem()
	newVal.Set(val.Elem())
	return newVal.Addr().Interface().(DrawInstance)
}

const ShaderBaseDataStart = unsafe.Offsetof(ShaderDataBase{}.model)

type ShaderDataBase struct {
	aabb        collision.AABB
	destroyed   bool
	deactivated bool
	viewCulled  bool
	_           [1]byte // Byte alignment
	shadow      DrawInstance
	transform   *matrix.Transform
	InitModel   matrix.Mat4
	model       matrix.Mat4
}

type ShaderDataCombine struct {
	ShaderDataBase
	Color matrix.Color
}

func (t ShaderDataCombine) Size() int {
	return int(unsafe.Sizeof(ShaderDataCombine{}) - ShaderBaseDataStart)
}

func NewShaderDataBase() ShaderDataBase {
	sdb := ShaderDataBase{}
	sdb.Setup()
	return sdb
}

func (s *ShaderDataBase) Setup() {
	s.SetModel(matrix.Mat4Identity())
}

func (s *ShaderDataBase) SelectLights(lights LightsForRender) {}
func (s *ShaderDataBase) Transform() *matrix.Transform        { return s.transform }

func (s *ShaderDataBase) Base() *ShaderDataBase                    { return s }
func (s *ShaderDataBase) SkinningHeader() *SkinnedShaderDataHeader { return nil }
func (s *ShaderDataBase) Destroy()                                 { s.destroyed = true }
func (s *ShaderDataBase) CancelDestroy()                           { s.destroyed = false }
func (s *ShaderDataBase) IsDestroyed() bool                        { return s.destroyed }
func (s *ShaderDataBase) IsInView() bool                           { return !s.deactivated && !s.viewCulled }
func (s *ShaderDataBase) Model() matrix.Mat4                       { return s.model }
func (s *ShaderDataBase) ModelPtr() *matrix.Mat4                   { return &s.model }

func (s *ShaderDataBase) Activate() {
	s.deactivated = false
	if s.shadow != nil {
		s.shadow.Activate()
	}
}

func (s *ShaderDataBase) Deactivate() {
	s.deactivated = true
	if s.shadow != nil {
		s.shadow.Deactivate()
	}
}

func (s *ShaderDataBase) setTransform(transform *matrix.Transform) {
	s.transform = transform
	s.forceUpdateTransformModel()
	if s.transform != nil {
		s.transform.SetDirty()
	}
}

func (s *ShaderDataBase) setShadow(shadow DrawInstance) {
	s.shadow = shadow
	if s.deactivated {
		s.shadow.Deactivate()
	}
}

func (s *ShaderDataBase) SetModel(model matrix.Mat4) {
	s.InitModel = model
	if s.transform == nil {
		s.model = model
	}
}

func (s *ShaderDataBase) forceUpdateTransformModel() {
	if s.transform == nil {
		return
	}
	s.model = matrix.Mat4Multiply(s.InitModel, s.transform.WorldMatrix())
}

func (s *ShaderDataBase) UpdateModel(viewCuller ViewCuller, container collision.AABB) {
	recalcCulling := false
	if viewCuller != nil {
		recalcCulling = viewCuller.ViewChanged()
	}
	if s.transform != nil && s.transform.IsDirty() {
		s.forceUpdateTransformModel()
		a := s.model.TransformPoint(container.Min())
		b := s.model.TransformPoint(container.Max())
		s.aabb = collision.AABBFromMinMax(a, b)
		recalcCulling = true
	} else if s.transform == nil {
		s.aabb = container
	}
	if recalcCulling {
		s.viewCulled = !viewCuller.IsInView(s.aabb)
	}
}

func (s *ShaderDataBase) renderBounds() collision.AABB { return s.aabb }

func (s *ShaderDataBase) DataPointer() unsafe.Pointer {
	return unsafe.Pointer(&s.model[0])
}

func (s *ShaderDataBase) UpdateBoundData() bool { return false }

func (s *ShaderDataBase) BoundDataPointer() unsafe.Pointer { return nil }

func (s *ShaderDataBase) InstanceBoundDataSize() int { return 0 }

type InstanceCopyData struct {
	byteMapping [maxFramesInFlight]unsafe.Pointer
	padding     int
	length      int
}

func InstanceCopyDataNew(padding int) InstanceCopyData {
	return InstanceCopyData{
		padding: padding,
	}
}

type DrawInstanceGroup struct {
	Mesh *Mesh
	InstanceDriverData
	MaterialInstance  *Material
	viewCuller        ViewCuller
	Instances         []DrawInstance
	rawData           InstanceCopyData
	boundInstanceData []InstanceCopyData
	instanceSize      int
	visibleCount      int
	sort              int
	destroyed         bool
}

func NewDrawInstanceGroup(mesh *Mesh, dataSize int, viewCuller ViewCuller) DrawInstanceGroup {
	return DrawInstanceGroup{
		Mesh:              mesh,
		Instances:         make([]DrawInstance, 0),
		rawData:           InstanceCopyDataNew(dataSize % 16),
		boundInstanceData: make([]InstanceCopyData, 0),
		instanceSize:      dataSize,
		destroyed:         false,
		viewCuller:        viewCuller,
	}
}

func (d *DrawInstanceGroup) AlterPadding(blockSize int) {
	newPadding := 0
	remainder := d.instanceSize % blockSize
	if remainder != 0 {
		newPadding = blockSize - remainder
	}
	if d.rawData.padding != newPadding {
		d.rawData.padding = newPadding
		d.rawData.length = d.TotalSize()
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
	return len(d.Instances) * (d.instanceSize + d.rawData.padding)
}

func (d *DrawInstanceGroup) AddInstance(instance DrawInstance) {
	d.Instances = append(d.Instances, instance)
	d.rawData.length = d.instanceSize + d.rawData.padding
	for i := range d.MaterialInstance.shaderInfo.LayoutGroups {
		g := &d.MaterialInstance.shaderInfo.LayoutGroups[i]
		for j := range g.Layouts {
			if g.Layouts[j].IsBuffer() {
				b := &g.Layouts[j]
				if len(d.boundInstanceData) <= b.Binding {
					grow := (b.Binding + 1) - len(d.boundInstanceData)
					d.boundInstanceData = klib.SliceSetLen(d.boundInstanceData, grow)
				}
				s := &d.boundInstanceData[b.Binding]
				if s.length < b.Capacity() {
					s.length = instance.InstanceBoundDataSize() + s.padding
				}
			}
		}
	}
}

func (d *DrawInstanceGroup) AnyVisible() bool  { return d.visibleCount > 0 }
func (d *DrawInstanceGroup) VisibleCount() int { return d.visibleCount }

func (d *DrawInstanceGroup) VisibleSize() int {
	return d.visibleCount * (d.instanceSize + d.rawData.padding)
}

func (d *DrawInstanceGroup) updateBoundData(index, bindingId int, instance DrawInstance, frame int) {
	if !instance.UpdateBoundData() {
		return
	}
	if ptr := instance.BoundDataPointer(); ptr != nil {
		nb := d.boundBuffers[bindingId]
		data := d.boundInstanceData[bindingId]
		offset := uintptr((nb.stride) * index)
		base := data.byteMapping[frame]
		to := unsafe.Pointer(uintptr(base) + offset)
		klib.Memcpy(to, ptr, uint64(nb.stride))
	}
}

func (d *DrawInstanceGroup) UpdateData(renderer Renderer, frame int, lights LightsForRender) {
	defer tracing.NewRegion("DrawInstanceGroup.UpdateData").End()
	base := d.rawData.byteMapping[frame]
	offset := uintptr(0)
	count := len(d.Instances)
	d.visibleCount = 0
	instanceIndex := 0
	for i := 0; i < count; i++ {
		instance := d.Instances[i]
		// This gives me a tiny fraction of extra perf for some reason, don't judge me
		instanceBase := instance.Base()
		if instanceBase.IsDestroyed() {
			d.Instances[i] = d.Instances[count-1]
			i--
			count--
			continue
		}
		instanceBase.UpdateModel(d.viewCuller, d.Mesh.Bounds())
		if instanceBase.IsInView() {
			if d.MaterialInstance.IsLit {
				instance.SelectLights(lights)
			}
			if d.generatedSets && len(d.boundInstanceData) > 0 {
				for j := range d.boundInstanceData {
					d.updateBoundData(instanceIndex, j, instance, frame)
				}
			}
			to := unsafe.Pointer(uintptr(base) + offset)
			klib.Memcpy(to, instanceBase.DataPointer(), uint64(d.instanceSize))
			offset += uintptr(d.instanceSize + d.rawData.padding)
			d.visibleCount++
			instanceIndex++
		}
	}
	if count < len(d.Instances) {
		d.rawData.length = count * (d.instanceSize + d.rawData.padding)
		d.Instances = d.Instances[:count]
	}
	d.bindInstanceDriverData()
	if len(d.Instances) == 0 {
		renderer.DestroyGroup(d)
		d.destroyed = true
	}
}

func (d *DrawInstanceGroup) Clear(renderer Renderer) {
	if d.destroyed {
		return
	}
	for i := range d.Instances {
		d.Instances[i].Destroy()
	}
}

func (d *DrawInstanceGroup) Destroy(renderer Renderer) {
	if d.destroyed {
		return
	}
	d.Clear(renderer)
	d.Instances = klib.WipeSlice(d.Instances)
	d.Mesh = nil
	d.MaterialInstance = nil
	renderer.DestroyGroup(d)
}
