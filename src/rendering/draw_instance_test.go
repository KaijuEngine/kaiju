/******************************************************************************/
/* draw_instance_test.go                                                      */
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

func aabbApprox(a, b graviton.AABB) bool {
	return matrix.Vec3ApproxTo(a.Min(), b.Min(), 0.0001) &&
		matrix.Vec3ApproxTo(a.Max(), b.Max(), 0.0001)
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

func TestVisibilityStateDefaultsAndOverrides(t *testing.T) {
	base := NewShaderDataBase()
	visibility := base.VisibilityState()
	if !visibility.FrustumVisible || visibility.OcclusionEligible ||
		!visibility.LastOcclusionVisible || visibility.ForceVisible {
		t.Fatalf("default visibility state = %+v", *visibility)
	}
	if !base.IsInView() {
		t.Fatalf("new base should be visible by default")
	}
	visibility.FrustumVisible = false
	if base.IsInView() {
		t.Fatalf("frustum-hidden instance should not be visible")
	}
	visibility.ForceVisible = true
	if !base.IsInView() {
		t.Fatalf("force visible should bypass culling state")
	}
	base.Deactivate()
	if base.IsInView() {
		t.Fatalf("deactivation should still hide force-visible instances")
	}
	base.Activate()
	visibility.ForceVisible = false
	visibility.FrustumVisible = true
	visibility.OcclusionEligible = true
	visibility.LastOcclusionVisible = false
	if base.IsInView() {
		t.Fatalf("occlusion-hidden eligible instance should not be visible")
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
	container := graviton.AABBFromMinMax(matrix.Vec3{-2, -1, -0.5}, matrix.Vec3{2, 1, 0.5})

	var transform matrix.Transform
	transform.SetupRawTransform()
	transform.SetScale(matrix.Vec3{1, 3, 2})
	transform.SetRotation(matrix.Vec3{0, 0, 45})
	transform.SetPosition(matrix.Vec3{5, -2, 1})
	base.setTransform(&transform)

	culler := &testViewCuller{inView: true}
	base.UpdateModel(culler, container)

	want := container.Transform(base.Model())
	if !aabbApprox(base.renderBounds(), want) {
		t.Fatalf("bounds = %+v, want %+v", base.renderBounds(), want)
	}
	if !aabbApprox(culler.seen, want) {
		t.Fatalf("culler saw bounds = %+v, want %+v", culler.seen, want)
	}
	renderBounds := base.renderBounds()
	for _, corner := range container.Transform(base.Model()).Corners() {
		if !renderBounds.Contains(corner) {
			t.Fatalf("bounds %+v did not contain transformed corner %v", renderBounds, corner)
		}
	}
}

func TestShaderDataBaseCulling(t *testing.T) {
	base := NewShaderDataBase()
	culler := &testViewCuller{inView: false, viewChanged: true}
	base.UpdateModel(culler, graviton.AABBFromWidth(matrix.Vec3Zero(), 1))
	if !base.viewCulled || base.VisibilityState().FrustumVisible || base.IsInView() {
		t.Fatalf("out-of-view culling was not applied")
	}
	culler.inView = true
	base.UpdateModel(culler, graviton.AABBFromWidth(matrix.Vec3Zero(), 1))
	if base.viewCulled || !base.VisibilityState().FrustumVisible || !base.IsInView() {
		t.Fatalf("in-view culling was not applied")
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

func TestDrawInstanceGroupVisibilityCounters(t *testing.T) {
	group := NewDrawInstanceGroup(NewMesh("mesh", testVerts(), []uint32{0, 1}), 16, nil)
	visible := NewShaderDataBase()
	group.countVisibility(&visible)

	frustumCulled := NewShaderDataBase()
	frustumCulled.VisibilityState().FrustumVisible = false
	group.countVisibility(&frustumCulled)

	occlusionCulled := NewShaderDataBase()
	occlusionCulled.VisibilityState().OcclusionEligible = true
	occlusionCulled.VisibilityState().LastOcclusionVisible = false
	group.countVisibility(&occlusionCulled)

	forced := NewShaderDataBase()
	forced.VisibilityState().FrustumVisible = false
	forced.VisibilityState().OcclusionEligible = true
	forced.VisibilityState().LastOcclusionVisible = false
	forced.VisibilityState().ForceVisible = true
	group.countVisibility(&forced)

	deactivated := NewShaderDataBase()
	deactivated.Deactivate()
	group.countVisibility(&deactivated)

	want := VisibilityCounters{
		TotalInstances:  5,
		FrustumCulled:   1,
		OcclusionTested: 1,
		OcclusionCulled: 1,
		Visible:         2,
	}
	if got := group.VisibilityCounters(); got != want {
		t.Fatalf("visibility counters = %+v, want %+v", got, want)
	}
}

func TestDrawingsVisibilityCounters(t *testing.T) {
	first := DrawInstanceGroup{visibilityCounters: VisibilityCounters{
		TotalInstances: 2,
		FrustumCulled:  1,
		Visible:        1,
	}}
	second := DrawInstanceGroup{visibilityCounters: VisibilityCounters{
		TotalInstances:  3,
		OcclusionTested: 2,
		OcclusionCulled: 1,
		Visible:         2,
	}}
	drawings := Drawings{renderPassGroups: []RenderPassGroup{{
		draws: []ShaderDraw{{instanceGroups: []DrawInstanceGroup{first, second}}},
	}}}
	want := VisibilityCounters{
		TotalInstances:  5,
		FrustumCulled:   1,
		OcclusionTested: 2,
		OcclusionCulled: 1,
		Visible:         3,
	}
	if got := drawings.VisibilityCounters(); got != want {
		t.Fatalf("visibility counters = %+v, want %+v", got, want)
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
