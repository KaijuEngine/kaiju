/******************************************************************************/
/* draw_instance.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"reflect"
	"unsafe"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type ViewCuller interface {
	IsInView(box graviton.AABB) bool
	ViewChanged() bool
}

type renderViewFrustumCamera interface {
	Frustum() graviton.Frustum
	IsDirty() bool
}

type renderViewCameraCuller struct {
	camera renderViewFrustumCamera
}

func (c renderViewCameraCuller) IsInView(box graviton.AABB) bool {
	return box.IntersectsFrustum(c.camera.Frustum())
}

func (c renderViewCameraCuller) ViewChanged() bool {
	return c.camera.IsDirty()
}

func renderViewCuller(view *RenderView, fallback ViewCuller) ViewCuller {
	if view == nil {
		return fallback
	}
	camera := view.Camera()
	if camera == nil {
		return fallback
	}
	if culler, ok := camera.(ViewCuller); ok {
		return culler
	}
	if camera, ok := camera.(renderViewFrustumCamera); ok {
		return renderViewCameraCuller{camera: camera}
	}
	return fallback
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
	UpdateModel(viewCuller ViewCuller, container graviton.AABB)
	DataPointer() unsafe.Pointer
	// Returns true if it should write the data, otherwise false
	UpdateBoundData() bool
	BoundDataPointer() unsafe.Pointer
	InstanceBoundDataSize() int
	setTransform(transform *matrix.Transform)
	SelectLights(lights LightsForRender)
	addShadow(shadow DrawInstance)
	renderBounds() graviton.AABB
}

func ReflectDuplicateDrawInstance(target DrawInstance) DrawInstance {
	val := reflect.ValueOf(target)
	if !val.IsValid() {
		return nil
	}
	newVal := reflect.New(val.Elem().Type()).Elem()
	newVal.Set(val.Elem())
	dupe := newVal.Addr().Interface().(DrawInstance)
	dupe.Base().viewCullStates = nil
	return dupe
}

const ShaderBaseDataStart = unsafe.Offsetof(ShaderDataBase{}.model)

type ShaderDataBase struct {
	aabb           graviton.AABB
	destroyed      bool
	deactivated    bool
	viewCulled     bool
	_              [1]byte // Byte alignment
	viewCullStates map[*RenderView]bool
	shadows        []DrawInstance
	transform      *matrix.Transform
	InitModel      matrix.Mat4
	model          matrix.Mat4
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
func (s *ShaderDataBase) CancelDestroy()                           { s.destroyed = false }
func (s *ShaderDataBase) IsDestroyed() bool                        { return s.destroyed }
func (s *ShaderDataBase) IsInView() bool                           { return !s.deactivated && !s.viewCulled }
func (s *ShaderDataBase) IsInViewForView(view *RenderView) bool {
	if s.viewCullStates != nil {
		if culled, ok := s.viewCullStates[view]; ok {
			return !s.deactivated && !culled
		}
	}
	return s.IsInView()
}
func (s *ShaderDataBase) Model() matrix.Mat4     { return s.model }
func (s *ShaderDataBase) ModelPtr() *matrix.Mat4 { return &s.model }

func (s *ShaderDataBase) Destroy() {
	s.destroyed = true
	for i := range s.shadows {
		s.shadows[i].Destroy()
	}
}

func (s *ShaderDataBase) Activate() {
	s.deactivated = false
	for i := range s.shadows {
		s.shadows[i].Activate()
	}
}

func (s *ShaderDataBase) Deactivate() {
	s.deactivated = true
	for i := range s.shadows {
		s.shadows[i].Deactivate()
	}
}

func (s *ShaderDataBase) setTransform(transform *matrix.Transform) {
	s.transform = transform
	s.forceUpdateTransformModel()
	if s.transform != nil {
		s.transform.SetDirty()
	}
}

func (s *ShaderDataBase) addShadow(shadow DrawInstance) {
	s.shadows = append(s.shadows, shadow)
	if s.deactivated {
		shadow.Deactivate()
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

func (s *ShaderDataBase) UpdateModel(viewCuller ViewCuller, container graviton.AABB) {
	s.UpdateModelForView(nil, viewCuller, container)
}

func (s *ShaderDataBase) UpdateModelForView(view *RenderView, viewCuller ViewCuller, container graviton.AABB) bool {
	recalcCulling := false
	if viewCuller != nil {
		recalcCulling = viewCuller.ViewChanged()
	}
	if s.transform != nil && s.transform.IsDirty() {
		s.forceUpdateTransformModel()
		a := s.model.TransformPoint(container.Min())
		b := s.model.TransformPoint(container.Max())
		s.aabb = graviton.AABBFromMinMax(a, b)
		recalcCulling = true
	} else if s.transform == nil {
		s.aabb = container
	}
	if s.viewCullStates == nil {
		s.viewCullStates = make(map[*RenderView]bool)
	}
	culled, known := s.viewCullStates[view]
	if viewCuller == nil {
		culled = false
	} else if recalcCulling || !known {
		culled = !viewCuller.IsInView(s.aabb)
	}
	s.viewCullStates[view] = culled
	if view == nil {
		s.viewCulled = culled
	}
	return !s.deactivated && !culled
}

func (s *ShaderDataBase) renderBounds() graviton.AABB { return s.aabb }

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

type DrawInstanceViewState struct {
	InstanceDriverData
	rawData           InstanceCopyData
	boundInstanceData []InstanceCopyData
	visibleCount      int
}

func NewDrawInstanceViewState(dataSize int) DrawInstanceViewState {
	return DrawInstanceViewState{
		rawData:           InstanceCopyDataNew(dataSize % 16),
		boundInstanceData: make([]InstanceCopyData, 0),
	}
}

type DrawInstanceGroup struct {
	Mesh *Mesh
	InstanceDriverData
	MaterialInstance  *Material
	Layer             RenderLayerMask
	viewCuller        ViewCuller
	Instances         []DrawInstance
	rawData           InstanceCopyData
	boundInstanceData []InstanceCopyData
	viewStates        map[*RenderView]*DrawInstanceViewState
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
		viewStates:        make(map[*RenderView]*DrawInstanceViewState),
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

func (d *DrawInstanceGroup) EffectiveLayer() RenderLayerMask {
	return normalizeRenderLayerMask(d.Layer)
}

func (d *DrawInstanceGroup) MatchesLayer(mask RenderLayerMask) bool {
	return d.EffectiveLayer()&mask != 0
}

func (d *DrawInstanceGroup) IsReady() bool {
	// TODO:  Check if textures are ready?
	return d.Mesh.IsReady() && !d.IsEmpty()
}

func (d *DrawInstanceGroup) TotalSize() int {
	return len(d.Instances) * (d.instanceSize + d.rawData.padding)
}

func (d *DrawInstanceGroup) viewStateForView(view *RenderView) *DrawInstanceViewState {
	if d.viewStates == nil {
		d.viewStates = make(map[*RenderView]*DrawInstanceViewState)
	}
	if state, ok := d.viewStates[view]; ok {
		return state
	}
	state := NewDrawInstanceViewState(d.instanceSize)
	state.rawData.length = d.rawData.length
	state.rawData.padding = d.rawData.padding
	state.boundInstanceData = make([]InstanceCopyData, len(d.boundInstanceData))
	for i := range d.boundInstanceData {
		state.boundInstanceData[i].padding = d.boundInstanceData[i].padding
		state.boundInstanceData[i].length = d.boundInstanceData[i].length
	}
	d.viewStates[view] = &state
	return &state
}

func (d *DrawInstanceGroup) syncViewStateTemplates() {
	for _, state := range d.viewStates {
		state.rawData.padding = d.rawData.padding
		state.rawData.length = d.rawData.length
		if len(state.boundInstanceData) < len(d.boundInstanceData) {
			grow := len(d.boundInstanceData) - len(state.boundInstanceData)
			state.boundInstanceData = klib.SliceSetLen(state.boundInstanceData, grow)
		}
		for i := range d.boundInstanceData {
			state.boundInstanceData[i].padding = d.boundInstanceData[i].padding
			state.boundInstanceData[i].length = d.boundInstanceData[i].length
		}
	}
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
	d.syncViewStateTemplates()
}

func (d *DrawInstanceGroup) AnyVisible() bool  { return d.visibleCount > 0 }
func (d *DrawInstanceGroup) VisibleCount() int { return d.visibleCount }
func (d *DrawInstanceGroup) AnyVisibleForView(view *RenderView) bool {
	return d.viewStateForView(view).visibleCount > 0
}
func (d *DrawInstanceGroup) VisibleCountForView(view *RenderView) int {
	return d.viewStateForView(view).visibleCount
}

func (d *DrawInstanceGroup) VisibleSize() int {
	return d.visibleCount * (d.instanceSize + d.rawData.padding)
}

func (d *DrawInstanceGroup) updateBoundData(state *DrawInstanceViewState, index, bindingId int, instance DrawInstance, frame int) {
	if !instance.UpdateBoundData() {
		return
	}
	if ptr := instance.BoundDataPointer(); ptr != nil {
		nb := state.boundBuffers[bindingId]
		data := state.boundInstanceData[bindingId]
		offset := uintptr((nb.stride) * index)
		base := data.byteMapping[frame]
		to := unsafe.Pointer(uintptr(base) + offset)
		klib.Memcpy(to, ptr, uint64(nb.stride))
	}
}

func (d *DrawInstanceGroup) UpdateData(device *GPUDevice, frame int, lights LightsForRender) {
	d.UpdateDataForView(device, frame, lights, nil)
}

func (d *DrawInstanceGroup) UpdateDataForView(device *GPUDevice, frame int, lights LightsForRender, view *RenderView) {
	defer tracing.NewRegion("DrawInstanceGroup.UpdateData").End()
	state := d.viewStateForView(view)
	base := state.rawData.byteMapping[frame]
	offset := uintptr(0)
	count := len(d.Instances)
	state.visibleCount = 0
	instanceIndex := 0
	viewCuller := renderViewCuller(view, d.viewCuller)
	if d.EffectiveLayer() == RenderLayerUI {
		viewCuller = d.viewCuller
	}
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
		if instanceBase.UpdateModelForView(view, viewCuller, d.Mesh.Bounds()) {
			if d.MaterialInstance.IsLit {
				instance.SelectLights(lights)
			}
			if state.generatedSets && len(state.boundInstanceData) > 0 {
				for j := range state.boundInstanceData {
					d.updateBoundData(state, instanceIndex, j, instance, frame)
				}
			}
			to := unsafe.Pointer(uintptr(base) + offset)
			klib.Memcpy(to, instanceBase.DataPointer(), uint64(d.instanceSize))
			offset += uintptr(d.instanceSize + state.rawData.padding)
			state.visibleCount++
			instanceIndex++
		}
	}
	if count < len(d.Instances) {
		d.rawData.length = count * (d.instanceSize + d.rawData.padding)
		d.Instances = d.Instances[:count]
		d.syncViewStateTemplates()
	}
	d.visibleCount = state.visibleCount
	d.bindInstanceDriverData(state)
	if len(d.Instances) == 0 {
		device.LogicalDevice.DestroyGroup(d)
		d.destroyed = true
	}
}

func (d *DrawInstanceGroup) Clear() {
	if d.destroyed {
		return
	}
	for i := range d.Instances {
		d.Instances[i].Destroy()
	}
}

func (d *DrawInstanceGroup) Destroy(device *GPUDevice) {
	if d.destroyed {
		return
	}
	d.Clear()
	d.Instances = klib.WipeSlice(d.Instances)
	d.Mesh = nil
	d.MaterialInstance = nil
	device.LogicalDevice.DestroyGroup(d)
	d.viewStates = nil
}
