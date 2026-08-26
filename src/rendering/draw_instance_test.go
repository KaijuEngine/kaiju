/******************************************************************************/
/* draw_instance_test.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"
	"unsafe"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

type testDrawInstance struct {
	ShaderDataBase
	boundData [4]float32
}

func newTestDrawInstance() *testDrawInstance {
	return &testDrawInstance{ShaderDataBase: NewShaderDataBase()}
}

func (d *testDrawInstance) Size() int                        { return int(unsafe.Sizeof(d.boundData)) }
func (d *testDrawInstance) UpdateBoundData() bool            { return true }
func (d *testDrawInstance) BoundDataPointer() unsafe.Pointer { return unsafe.Pointer(&d.boundData[0]) }
func (d *testDrawInstance) InstanceBoundDataSize() int       { return int(unsafe.Sizeof(d.boundData)) }

type testViewCuller struct {
	inView      bool
	viewChanged bool
	seen        graviton.AABB
}

func (c *testViewCuller) IsInView(box graviton.AABB) bool {
	c.seen = box
	return c.inView
}

func (c *testViewCuller) ViewChanged() bool { return c.viewChanged }

type testMeshLodCamera struct {
	position     matrix.Vec3
	inView       bool
	orthographic bool
}

func (c *testMeshLodCamera) Position() matrix.Vec3 { return c.position }
func (c *testMeshLodCamera) IsOrthographic() bool  { return c.orthographic }
func (c *testMeshLodCamera) IsInView(graviton.AABB) bool {
	return c.inView
}
func (c *testMeshLodCamera) ViewChanged() bool { return true }

func testMeshLodChain() (*Mesh, *Mesh, *Mesh) {
	verts := []Vertex{
		{Position: matrix.Vec3{-1, -1, -1}},
		{Position: matrix.Vec3{1, 1, 1}},
	}
	base := NewMesh("lod_base", verts, []uint32{0, 1})
	lod1 := NewMesh("lod_1", verts, []uint32{0, 1})
	lod2 := NewMesh("lod_2", verts, []uint32{0, 1})
	base.lods = MeshLod{Levels: []MeshLODInstance{
		{Mesh: base, Ratio: 1},
		{Mesh: lod1, Ratio: 0.5},
		{Mesh: lod2, Ratio: 0.25},
	}}
	return base, lod1, lod2
}

func translatedTestDrawInstance(position matrix.Vec3) *testDrawInstance {
	instance := newTestDrawInstance()
	model := matrix.Mat4Identity()
	model.Translate(position)
	instance.SetModel(model)
	return instance
}

func TestReflectDuplicateDrawInstance(t *testing.T) {
	if ReflectDuplicateDrawInstance(nil) != nil {
		t.Fatalf("nil duplicate should be nil")
	}
	original := newTestDrawInstance()
	original.SetModel(matrix.Mat4Identity())
	dupe := ReflectDuplicateDrawInstance(original)
	if dupe == nil || dupe == original {
		t.Fatalf("duplicate = %v, original = %v", dupe, original)
	}
	if dupe.Base().Model() != original.Model() {
		t.Fatalf("duplicate did not copy model")
	}
}

func TestShaderDataBaseSetupAndModel(t *testing.T) {
	base := NewShaderDataBase()
	if base.Model() != matrix.Mat4Identity() || base.InitModel != matrix.Mat4Identity() {
		t.Fatalf("new base should start with identity model")
	}
	model := matrix.Mat4Identity()
	model.Translate(matrix.Vec3{1, 2, 3})
	base.SetModel(model)
	if base.Model() != model || *base.ModelPtr() != model {
		t.Fatalf("SetModel without transform did not update model")
	}
	if base.DataPointer() != unsafe.Pointer(&base.model[0]) {
		t.Fatalf("DataPointer should point at model data")
	}
	if base.BoundDataPointer() != nil || base.InstanceBoundDataSize() != 0 || base.UpdateBoundData() {
		t.Fatalf("base bound data defaults are wrong")
	}
}

func TestShaderDataBaseActivationAndDestroy(t *testing.T) {
	base := NewShaderDataBase()
	shadow := newTestDrawInstance()
	base.addShadow(shadow)
	base.Deactivate()
	if !base.deactivated || !shadow.deactivated || base.IsInView() || shadow.IsInView() {
		t.Fatalf("Deactivate did not propagate")
	}
	base.Activate()
	if base.deactivated || shadow.deactivated || !base.IsInView() || !shadow.IsInView() {
		t.Fatalf("Activate did not propagate")
	}
	base.Destroy()
	if !base.IsDestroyed() || !shadow.IsDestroyed() {
		t.Fatalf("Destroy did not propagate")
	}
	base.CancelDestroy()
	if base.IsDestroyed() {
		t.Fatalf("CancelDestroy did not clear destroyed flag")
	}
}

func TestShaderDataBaseTransformModelAndBounds(t *testing.T) {
	base := NewShaderDataBase()
	container := graviton.AABBFromMinMax(matrix.Vec3{-1, -1, -1}, matrix.Vec3{1, 1, 1})
	base.UpdateModel(nil, container)
	if base.renderBounds() != container {
		t.Fatalf("no-transform bounds = %+v, want %+v", base.renderBounds(), container)
	}

	var transform matrix.Transform
	transform.SetupRawTransform()
	transform.SetPosition(matrix.Vec3{5, 0, 0})
	base.setTransform(&transform)
	base.UpdateModel(&testViewCuller{inView: true}, container)
	if base.Transform() != &transform {
		t.Fatalf("transform was not stored")
	}
	if got := base.Model().TransformPoint(matrix.Vec3Zero()); got != (matrix.Vec3{5, 0, 0}) {
		t.Fatalf("model translation = %v", got)
	}
	if got := base.renderBounds().Center; got != (matrix.Vec3{5, 0, 0}) {
		t.Fatalf("transformed bounds center = %v", got)
	}
}

func TestShaderDataBaseTransformBoundsUsesAllCorners(t *testing.T) {
	base := NewShaderDataBase()
	container := graviton.AABBFromMinMax(matrix.Vec3{-1, -1, -1}, matrix.Vec3{1, 1, 1})

	var transform matrix.Transform
	transform.SetupRawTransform()
	transform.SetRotation(matrix.Vec3{0, 0, 45})
	base.setTransform(&transform)
	base.UpdateModel(&testViewCuller{inView: true}, container)

	want := container.Transform(base.Model())
	got := base.renderBounds()
	if !matrix.Vec3ApproxTo(got.Center, want.Center, 0.0001) ||
		!matrix.Vec3ApproxTo(got.Extent, want.Extent, 0.0001) {
		t.Fatalf("rotated bounds = %+v, want %+v", got, want)
	}
	if got.Extent.X() < 1.4 || got.Extent.Y() < 1.4 {
		t.Fatalf("rotated bounds collapsed on an axis: %+v", got)
	}
}

func TestShaderDataBaseCulling(t *testing.T) {
	base := NewShaderDataBase()
	culler := &testViewCuller{inView: false, viewChanged: true}
	base.UpdateModel(culler, graviton.AABBFromWidth(matrix.Vec3Zero(), 1))
	if !base.viewCulled || base.IsInView() {
		t.Fatalf("out-of-view culling was not applied")
	}
	culler.inView = true
	base.UpdateModel(culler, graviton.AABBFromWidth(matrix.Vec3Zero(), 1))
	if base.viewCulled || !base.IsInView() {
		t.Fatalf("in-view culling was not applied")
	}
}

func TestShaderDataBaseCullingIsPerRenderView(t *testing.T) {
	base := NewShaderDataBase()
	left := newRenderView(RenderViewOptions{Name: "left"}, 0)
	right := newRenderView(RenderViewOptions{Name: "right"}, 1)
	box := graviton.AABBFromWidth(matrix.Vec3Zero(), 1)

	base.UpdateModelForView(left, &testViewCuller{inView: false, viewChanged: true}, box)
	base.UpdateModelForView(right, &testViewCuller{inView: true, viewChanged: true}, box)
	if base.IsInViewForView(left) {
		t.Fatalf("left view should be culled")
	}
	if !base.IsInViewForView(right) {
		t.Fatalf("right view should remain visible")
	}
	if !base.IsInView() {
		t.Fatalf("legacy default culling state should not be overwritten by named views")
	}
}

func TestDrawInstanceGroupPaddingAndSizes(t *testing.T) {
	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	mesh.MeshId = testReadyMeshID()
	group := NewDrawInstanceGroup(mesh, 12, nil)
	if !group.IsEmpty() || group.IsReady() {
		t.Fatalf("new group empty/ready state is wrong")
	}
	group.MaterialInstance = &Material{}
	group.AddInstance(newTestDrawInstance())
	group.AlterPadding(16)
	if group.rawData.padding != 4 {
		t.Fatalf("padding = %d, want 4", group.rawData.padding)
	}
	if group.TotalSize() != 16 {
		t.Fatalf("TotalSize = %d, want 16", group.TotalSize())
	}
	group.visibleCount = 1
	if !group.AnyVisible() || group.VisibleCount() != 1 || group.VisibleSize() != 16 {
		t.Fatalf("visible sizing is wrong")
	}
	if !group.IsReady() {
		t.Fatalf("ready mesh with an instance should be ready")
	}
}

func TestDrawInstanceGroupViewStatesAreIndependent(t *testing.T) {
	group := NewDrawInstanceGroup(NewMesh("mesh", testVerts(), []uint32{0, 1}), 16, nil)
	group.MaterialInstance = &Material{}
	group.AddInstance(newTestDrawInstance())
	left := newRenderView(RenderViewOptions{Name: "left"}, 0)
	right := newRenderView(RenderViewOptions{Name: "right"}, 1)

	leftState := group.viewStateForView(left)
	rightState := group.viewStateForView(right)
	if leftState == rightState {
		t.Fatalf("expected separate state objects for separate render views")
	}
	leftState.visibleCount = 1
	rightState.visibleCount = 0
	if group.VisibleCountForView(left) != 1 || group.VisibleCountForView(right) != 0 {
		t.Fatalf("visible counts leaked between view states")
	}
	var leftByte, rightByte byte
	leftState.rawData.byteMapping[0] = unsafe.Pointer(&leftByte)
	rightState.rawData.byteMapping[0] = unsafe.Pointer(&rightByte)
	if leftState.rawData.byteMapping[0] == rightState.rawData.byteMapping[0] {
		t.Fatalf("raw instance buffer mappings should be per view")
	}
}

func TestSelectMeshLodUsesScaleAwareDistance(t *testing.T) {
	base, lod1, lod2 := testMeshLodChain()
	bounds := base.Bounds()
	tests := []struct {
		name   string
		camera any
		bounds graviton.AABB
		want   *Mesh
	}{
		{name: "missing camera", bounds: bounds, want: base},
		{name: "near", camera: &testMeshLodCamera{position: matrix.Vec3{0, 0, 9.9}}, bounds: bounds, want: base},
		{name: "first transition", camera: &testMeshLodCamera{position: matrix.Vec3{0, 0, 10}}, bounds: bounds, want: lod1},
		{name: "second transition", camera: &testMeshLodCamera{position: matrix.Vec3{0, 0, 20}}, bounds: bounds, want: lod2},
		{name: "clamped far", camera: &testMeshLodCamera{position: matrix.Vec3{0, 0, 100}}, bounds: bounds, want: lod2},
		{name: "scaled object remains near", camera: &testMeshLodCamera{position: matrix.Vec3{0, 0, 15}}, bounds: graviton.NewAABB(matrix.Vec3Zero(), matrix.Vec3{2, 2, 2}), want: base},
		{name: "orthographic uses source", camera: &testMeshLodCamera{position: matrix.Vec3{0, 0, 100}, orthographic: true}, bounds: bounds, want: base},
		{name: "zero size uses source", camera: &testMeshLodCamera{position: matrix.Vec3{0, 0, 100}}, bounds: graviton.NewAABB(matrix.Vec3Zero(), matrix.Vec3Zero()), want: base},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := selectMeshLod(base, test.bounds, test.camera); got != test.want {
				t.Fatalf("selected mesh = %q, want %q", got.Key(), test.want.Key())
			}
		})
	}
}

func TestDrawInstanceGroupCaptureBatchesInstancesByLod(t *testing.T) {
	base, lod1, lod2 := testMeshLodChain()
	group := NewDrawInstanceGroup(base, newTestDrawInstance().Size(), nil)
	group.MaterialInstance = &Material{}
	group.AddInstance(translatedTestDrawInstance(matrix.Vec3{0, 0, -5}))
	group.AddInstance(translatedTestDrawInstance(matrix.Vec3{0, 0, -15}))
	group.AddInstance(translatedTestDrawInstance(matrix.Vec3{0, 0, -25}))
	view := newRenderView(RenderViewOptions{
		Name:   "lod",
		Camera: &testMeshLodCamera{inView: true},
	}, 0)

	group.CaptureDataForView(LightsForRender{}, newRenderViewFrame(view))
	state := group.viewStateForView(view)
	batches := state.frameData.lodBatches
	if state.frameData.visibleCount != 3 || len(batches) != 3 {
		t.Fatalf("visible/batches = %d/%d, want 3/3", state.frameData.visibleCount, len(batches))
	}
	wantMeshes := []*Mesh{base, lod1, lod2}
	for i := range batches {
		if batches[i].Mesh != wantMeshes[i] || batches[i].FirstInstance != uint32(i) || batches[i].InstanceCount != 1 {
			t.Fatalf("batch %d = %+v, want mesh %q first %d count 1", i, batches[i], wantMeshes[i].Key(), i)
		}
	}

	group.UpdateDataForView(&GPUDevice{}, 0, LightsForRender{}, view)
	if len(state.lodBatches) != 3 || state.lodBatches[2].Mesh != lod2 {
		t.Fatalf("captured LOD batches were not applied to render state: %+v", state.lodBatches)
	}
}

func TestDrawInstanceGroupMergesLodLevelsThatReuseMesh(t *testing.T) {
	base, lod1, _ := testMeshLodChain()
	base.lods.Levels[2].Mesh = lod1
	group := NewDrawInstanceGroup(base, newTestDrawInstance().Size(), nil)
	group.MaterialInstance = &Material{}
	group.AddInstance(translatedTestDrawInstance(matrix.Vec3{0, 0, -15}))
	group.AddInstance(translatedTestDrawInstance(matrix.Vec3{0, 0, -25}))
	view := newRenderView(RenderViewOptions{
		Name:   "lod",
		Camera: &testMeshLodCamera{inView: true},
	}, 0)

	group.CaptureDataForView(LightsForRender{}, newRenderViewFrame(view))
	batches := group.viewStateForView(view).frameData.lodBatches
	if len(batches) != 1 || batches[0].Mesh != lod1 || batches[0].FirstInstance != 0 || batches[0].InstanceCount != 2 {
		t.Fatalf("reused LOD mesh was not merged into one batch: %+v", batches)
	}
}

func TestDrawInstanceGroupSelectsLodPerRenderView(t *testing.T) {
	base, lod1, _ := testMeshLodChain()
	group := NewDrawInstanceGroup(base, newTestDrawInstance().Size(), nil)
	group.MaterialInstance = &Material{}
	group.AddInstance(translatedTestDrawInstance(matrix.Vec3{0, 0, -15}))
	far := newRenderView(RenderViewOptions{
		Name:   "far",
		Camera: &testMeshLodCamera{inView: true},
	}, 0)
	near := newRenderView(RenderViewOptions{
		Name:   "near",
		Camera: &testMeshLodCamera{position: matrix.Vec3{0, 0, -10}, inView: true},
	}, 1)

	group.CaptureDataForView(LightsForRender{}, newRenderViewFrame(far))
	group.CaptureDataForView(LightsForRender{}, newRenderViewFrame(near))
	farBatches := group.viewStateForView(far).frameData.lodBatches
	nearBatches := group.viewStateForView(near).frameData.lodBatches
	if len(farBatches) != 1 || farBatches[0].Mesh != lod1 {
		t.Fatalf("far view selected %+v, want LOD 1", farBatches)
	}
	if len(nearBatches) != 1 || nearBatches[0].Mesh != base {
		t.Fatalf("near view selected %+v, want source mesh", nearBatches)
	}
}

func TestDrawInstanceGroupUpdateDataIsPerRenderView(t *testing.T) {
	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	inst := newTestDrawInstance()
	group := NewDrawInstanceGroup(mesh, inst.Size(), nil)
	group.MaterialInstance = &Material{}
	group.AddInstance(inst)
	left := newRenderView(RenderViewOptions{
		Name:   "left",
		Camera: &testViewCuller{inView: false, viewChanged: true},
	}, 0)
	right := newRenderView(RenderViewOptions{
		Name:   "right",
		Camera: &testViewCuller{inView: true, viewChanged: true},
	}, 1)
	var leftBytes, rightBytes [64]byte
	group.viewStateForView(left).rawData.byteMapping[0] = unsafe.Pointer(&leftBytes[0])
	group.viewStateForView(right).rawData.byteMapping[0] = unsafe.Pointer(&rightBytes[0])

	group.UpdateDataForView(&GPUDevice{}, 0, LightsForRender{}, left)
	group.UpdateDataForView(&GPUDevice{}, 0, LightsForRender{}, right)
	if group.VisibleCountForView(left) != 0 {
		t.Fatalf("left view visible count = %d, want 0", group.VisibleCountForView(left))
	}
	if group.VisibleCountForView(right) != 1 {
		t.Fatalf("right view visible count = %d, want 1", group.VisibleCountForView(right))
	}
	if leftBytes != ([64]byte{}) {
		t.Fatalf("culled left view should not receive instance data")
	}
	if rightBytes == ([64]byte{}) {
		t.Fatalf("visible right view should receive instance data")
	}
}

func TestDrawInstanceGroupUIUsesFallbackCuller(t *testing.T) {
	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	inst := newTestDrawInstance()
	group := NewDrawInstanceGroup(mesh, inst.Size(), &testViewCuller{inView: true, viewChanged: true})
	group.MaterialInstance = &Material{}
	group.Layer = RenderLayerUI
	group.AddInstance(inst)
	view := newRenderView(RenderViewOptions{
		Name:   "default",
		Camera: &testViewCuller{inView: false, viewChanged: true},
	}, 0)
	var bytes [64]byte
	group.viewStateForView(view).rawData.byteMapping[0] = unsafe.Pointer(&bytes[0])

	group.UpdateDataForView(&GPUDevice{}, 0, LightsForRender{}, view)
	if group.VisibleCountForView(view) != 1 {
		t.Fatalf("UI layer should use its fallback culler; visible count = %d, want 1",
			group.VisibleCountForView(view))
	}
}

func TestDrawInstanceGroupAddInstance(t *testing.T) {
	group := NewDrawInstanceGroup(NewMesh("mesh", testVerts(), []uint32{0, 1}), 16, nil)
	group.MaterialInstance = &Material{shaderInfo: ShaderDataCompiled{LayoutGroups: []ShaderLayoutGroup{{
		Layouts: []ShaderLayout{{
			Binding:  2,
			Location: 1,
			Type:     "StorageBuffer",
			Source:   "buffer",
			Fields:   []ShaderLayoutStructField{{Type: "vec4", Name: "data"}},
		}},
	}}}}
	inst := newTestDrawInstance()
	group.AddInstance(inst)
	if len(group.Instances) != 1 || group.Instances[0] != inst {
		t.Fatalf("instance was not added")
	}
	if group.rawData.length != 16 {
		t.Fatalf("raw data length = %d, want 16", group.rawData.length)
	}
	if len(group.boundInstanceData) != 3 || group.boundInstanceData[2].length != inst.InstanceBoundDataSize() {
		t.Fatalf("bound data was not grown: %+v", group.boundInstanceData)
	}
}

func TestDrawInstanceGroupDetectsDescriptorLayoutChange(t *testing.T) {
	group := NewDrawInstanceGroup(NewMesh("mesh", testVerts(), []uint32{0, 1}), 16, nil)
	state := &DrawInstanceViewState{}
	state.generatedSets = true
	state.descriptorLayout = testDescriptorSetLayoutHandle(1)
	material := &Material{
		Shader: NewShader(ShaderDataCompiled{Name: "test"}),
	}
	material.Shader.RenderId.descriptorSetLayout = testDescriptorSetLayoutHandle(1)
	if group.instanceDescriptorLayoutChanged(material, state) {
		t.Fatalf("matching descriptor layout should not be reported as changed")
	}
	material.Shader.RenderId.descriptorSetLayout = testDescriptorSetLayoutHandle(2)
	if !group.instanceDescriptorLayoutChanged(material, state) {
		t.Fatalf("descriptor layout change was not detected")
	}
}

func TestDestroyGroupDescriptorSetsKeepsInstanceBuffers(t *testing.T) {
	state := &DrawInstanceViewState{}
	state.generatedSets = true
	state.descriptorPool = testDescriptorPoolHandle(3)
	state.descriptorSets[0] = testDescriptorSetHandle(4)
	state.descriptorLayout = testDescriptorSetLayoutHandle(5)
	state.instanceBuffer.buffers[0] = testBufferHandle(6)
	state.boundBuffers = []ShaderBuffer{{buffers: [maxFramesInFlight]GPUBuffer{testBufferHandle(7)}}}
	state.descriptorCache.ShouldWrite(0, NewDescriptorWriteSignature())

	device := &GPUDevice{}
	device.LogicalDevice.destroyGroupDescriptorSets(state)

	if state.generatedSets {
		t.Fatalf("descriptor sets should no longer be marked generated")
	}
	if state.descriptorPool.IsValid() || state.descriptorSets[0].IsValid() || state.descriptorLayout.IsValid() {
		t.Fatalf("descriptor handles were not cleared")
	}
	if !state.instanceBuffer.buffers[0].IsValid() || !state.boundBuffers[0].buffers[0].IsValid() {
		t.Fatalf("instance buffers should remain owned by the state")
	}
	if len(device.LogicalDevice.bufferTrash.trash) != 1 {
		t.Fatalf("descriptor trash count = %d, want 1", len(device.LogicalDevice.bufferTrash.trash))
	}
	trash := device.LogicalDevice.bufferTrash.trash[0]
	if !trash.pool.IsValid() || !trash.sets[0].IsValid() {
		t.Fatalf("descriptor trash did not capture pool/set: %+v", trash)
	}
	if trash.buffers[0].IsValid() {
		t.Fatalf("descriptor-only trash should not capture instance buffers")
	}
	if !state.descriptorCache.ShouldWrite(0, NewDescriptorWriteSignature()) {
		t.Fatalf("descriptor cache should be invalidated")
	}
}

func TestDrawInstanceGroupClear(t *testing.T) {
	group := NewDrawInstanceGroup(NewMesh("mesh", testVerts(), []uint32{0, 1}), 16, nil)
	group.MaterialInstance = &Material{}
	inst := newTestDrawInstance()
	group.AddInstance(inst)
	group.Clear()
	if !inst.IsDestroyed() {
		t.Fatalf("Clear should destroy instances")
	}
	group.destroyed = true
	inst.CancelDestroy()
	group.Clear()
	if inst.IsDestroyed() {
		t.Fatalf("Clear should no-op after destruction")
	}
}
