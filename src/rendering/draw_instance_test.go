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
