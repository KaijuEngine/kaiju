/******************************************************************************/
/* constraint_entity_data_renderer_test.go                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package data_binding_renderer

import (
	"reflect"
	"testing"
	"unsafe"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

func TestConstraintRendererAttachEachTypeCreatesGizmo(t *testing.T) {
	host, manager, owner, target := newConstraintRendererTestStage()
	renderer := &ConstraintEntityDataRenderer{
		Gizmos: make(map[*entity_data_binding.EntityDataEntry]*constraintGizmo),
	}
	entries := []*entity_data_binding.EntityDataEntry{
		constraintTestEntry(&engine_entity_data_physics.DistanceJointEntityData{
			ConnectedEntityId: engine.EntityId(target.StageData.Description.Id),
			TargetAnchorB:     matrix.NewVec3(0, 1, 0),
		}),
		constraintTestEntry(&engine_entity_data_physics.RopeJointEntityData{
			ConnectedEntityId: engine.EntityId(target.StageData.Description.Id),
			TargetAnchorB:     matrix.NewVec3(0, 1, 0),
		}),
		constraintTestEntry(&engine_entity_data_physics.PointJointEntityData{
			ConnectedEntityId: engine.EntityId(target.StageData.Description.Id),
			TargetAnchorB:     matrix.NewVec3(0, 1, 0),
		}),
		constraintTestEntry(&engine_entity_data_physics.HingeJointEntityData{
			ConnectedEntityId: engine.EntityId(target.StageData.Description.Id),
			TargetAnchorB:     matrix.NewVec3(0, 1, 0),
			HingeAxis:         matrix.Vec3Right(),
			EnableLimits:      true,
			MinAngleDegrees:   -35,
			MaxAngleDegrees:   45,
		}),
	}
	for _, entry := range entries {
		renderer.Attached(host, manager, owner, entry)
		g := renderer.Gizmos[entry]
		if g == nil {
			t.Fatalf("expected gizmo for %s", entry.Gen.RegisterKey)
		}
		if g.anchorA == nil || g.anchorB == nil || g.link == nil {
			t.Fatalf("expected anchors and link for %s", entry.Gen.RegisterKey)
		}
		if g.data.kind == constraintGizmoHinge && (g.axis == nil || g.arc == nil) {
			t.Fatalf("expected hinge axis and limit arc")
		}
	}
}

func TestConstraintRendererShowHideDetachDestroyCleansDraws(t *testing.T) {
	host, manager, owner, target := newConstraintRendererTestStage()
	renderer := &ConstraintEntityDataRenderer{
		Gizmos: make(map[*entity_data_binding.EntityDataEntry]*constraintGizmo),
	}
	entry := constraintTestEntry(&engine_entity_data_physics.DistanceJointEntityData{
		ConnectedEntityId: engine.EntityId(target.StageData.Description.Id),
	})
	renderer.Attached(host, manager, owner, entry)
	g := renderer.Gizmos[entry]
	if g == nil {
		t.Fatal("expected gizmo")
	}
	renderer.Show(host, owner, entry)
	if !g.anchorA.IsInView() || !g.anchorB.IsInView() || !g.link.IsInView() {
		t.Fatal("show should activate all draw instances")
	}
	renderer.Hide(host, owner, entry)
	if g.anchorA.IsInView() || g.anchorB.IsInView() || g.link.IsInView() {
		t.Fatal("hide should deactivate all draw instances")
	}
	anchorA, anchorB, link := g.anchorA, g.anchorB, g.link
	renderer.Detatched(host, manager, owner, entry)
	if !anchorA.IsDestroyed() || !anchorB.IsDestroyed() || !link.IsDestroyed() {
		t.Fatal("detach should destroy draw instances")
	}

	entry = constraintTestEntry(&engine_entity_data_physics.DistanceJointEntityData{
		ConnectedEntityId: engine.EntityId(target.StageData.Description.Id),
	})
	renderer.Attached(host, manager, owner, entry)
	g = renderer.Gizmos[entry]
	anchorA, anchorB, link = g.anchorA, g.anchorB, g.link
	owner.OnDestroy.Execute()
	if _, ok := renderer.Gizmos[entry]; ok {
		t.Fatal("owner destroy should remove the gizmo")
	}
	if !anchorA.IsDestroyed() || !anchorB.IsDestroyed() || !link.IsDestroyed() {
		t.Fatal("owner destroy should destroy draw instances")
	}
}

func TestConstraintRendererMovingOwnerOrTargetUpdatesAnchors(t *testing.T) {
	host, manager, owner, target := newConstraintRendererTestStage()
	renderer := &ConstraintEntityDataRenderer{
		Gizmos: make(map[*entity_data_binding.EntityDataEntry]*constraintGizmo),
	}
	entry := constraintTestEntry(&engine_entity_data_physics.DistanceJointEntityData{
		ConnectedEntityId: engine.EntityId(target.StageData.Description.Id),
		LocalAnchorA:      matrix.NewVec3(1, 0, 0),
		TargetAnchorB:     matrix.NewVec3(0, 1, 0),
	})
	renderer.Attached(host, manager, owner, entry)
	g := renderer.Gizmos[entry]
	oldA := g.lastAnchorA
	oldB := g.lastAnchorB

	owner.Transform.SetPosition(matrix.NewVec3(2, 0, 0))
	renderer.update(host, manager)
	if g.lastAnchorA.Equals(oldA) {
		t.Fatal("moving owner should update anchor A")
	}

	target.Transform.SetPosition(matrix.NewVec3(0, 3, 0))
	renderer.update(host, manager)
	if g.lastAnchorB.Equals(oldB) {
		t.Fatal("moving target should update anchor B")
	}
}

func TestConstraintRendererDataChangeAndInvalidTargetRefreshes(t *testing.T) {
	host, manager, owner, target := newConstraintRendererTestStage()
	renderer := &ConstraintEntityDataRenderer{
		Gizmos: make(map[*entity_data_binding.EntityDataEntry]*constraintGizmo),
	}
	entry := constraintTestEntry(&engine_entity_data_physics.PointJointEntityData{
		ConnectedEntityId: engine.EntityId("missing"),
		LocalAnchorA:      matrix.NewVec3(1, 0, 0),
		TargetAnchorB:     matrix.NewVec3(0, 1, 0),
	})
	renderer.Attached(host, manager, owner, entry)
	g := renderer.Gizmos[entry]
	if !g.invalidTarget {
		t.Fatal("missing connected entity should be marked invalid")
	}
	if g.link.(*shader_data_registry.ShaderDataEdTransformWire).Color != matrix.ColorOrange() {
		t.Fatal("invalid target should use warning color")
	}
	oldA := g.lastAnchorA
	oldB := g.lastAnchorB
	entry.SetFieldByName("LocalAnchorA", matrix.NewVec3(2, 0, 0))
	renderer.Update(host, owner, entry)
	if g.lastAnchorA.Equals(oldA) {
		t.Fatal("local anchor edits should refresh anchor A")
	}
	entry.SetFieldByName("TargetAnchorB", matrix.NewVec3(0, 4, 0))
	renderer.Update(host, owner, entry)
	if g.lastAnchorB.Equals(oldB) {
		t.Fatal("data field changes should refresh anchor B")
	}

	entry.SetFieldByName("ConnectedEntityId", engine.EntityId(target.StageData.Description.Id))
	renderer.Update(host, owner, entry)
	if g.invalidTarget || g.target != target {
		t.Fatal("target ID edits should resolve and refresh the target")
	}
	if g.link.(*shader_data_registry.ShaderDataEdTransformWire).Color != matrix.ColorGreen() {
		t.Fatal("valid target should restore normal gizmo color")
	}
	entry.SetFieldByName("ConnectedEntityId", engine.EntityId("still-missing"))
	renderer.Update(host, owner, entry)
	if !g.invalidTarget {
		t.Fatal("editing target ID to a missing entity should restore warning state")
	}
}

func TestConstraintRendererHingeAxisAndLimitEditsRefresh(t *testing.T) {
	host, manager, owner, target := newConstraintRendererTestStage()
	renderer := &ConstraintEntityDataRenderer{
		Gizmos: make(map[*entity_data_binding.EntityDataEntry]*constraintGizmo),
	}
	entry := constraintTestEntry(&engine_entity_data_physics.HingeJointEntityData{
		ConnectedEntityId: engine.EntityId(target.StageData.Description.Id),
		HingeAxis:         matrix.Vec3Right(),
		EnableLimits:      false,
	})
	renderer.Attached(host, manager, owner, entry)
	g := renderer.Gizmos[entry]
	if g == nil || g.axis == nil {
		t.Fatal("expected hinge axis drawing")
	}
	if g.arc != nil {
		t.Fatal("limits disabled should not create an arc")
	}

	oldAxis := g.lastAxis
	entry.SetFieldByName("HingeAxis", matrix.Vec3Up())
	renderer.Update(host, owner, entry)
	if g.lastAxis.Equals(oldAxis) {
		t.Fatal("axis edits should refresh the hinge axis")
	}

	entry.SetFieldByName("EnableLimits", true)
	entry.SetFieldByName("MinAngleDegrees", matrix.Float(-20))
	entry.SetFieldByName("MaxAngleDegrees", matrix.Float(60))
	renderer.Update(host, owner, entry)
	if g.arc == nil {
		t.Fatal("enabling limits should create a limit arc")
	}
	oldArc := g.arc
	oldArcKey := g.arcKey
	entry.SetFieldByName("MaxAngleDegrees", matrix.Float(90))
	renderer.Update(host, owner, entry)
	if g.arc == nil || g.arc == oldArc || g.arcKey == oldArcKey {
		t.Fatal("limit angle edits should rebuild the limit arc")
	}

	arcToDestroy := g.arc
	entry.SetFieldByName("EnableLimits", false)
	renderer.Update(host, owner, entry)
	if g.arc != nil || !arcToDestroy.IsDestroyed() {
		t.Fatal("disabling limits should destroy and clear the limit arc")
	}
}

func newConstraintRendererTestStage() (*engine.Host, *editor_stage_manager.StageManager, *editor_stage_manager.StageEntity, *editor_stage_manager.StageEntity) {
	host := engine.NewHost("constraint-renderer-test", nil, assets.NewMockDB(map[string][]byte{}))
	setHostField(host, "materialCache", rendering.NewMaterialCache(nil, host.AssetDatabase()))
	setHostField(host, "meshCache", rendering.NewMeshCache(nil, host.AssetDatabase()))
	host.MaterialCache().AddMaterial(&rendering.Material{
		Id:        assets.MaterialDefinitionEdTransformWire,
		Instances: make(map[string]*rendering.Material),
	})
	history := &memento.History{}
	history.Initialize(64)
	manager := &editor_stage_manager.StageManager{}
	manager.Initialize(host, history, nil)
	owner := manager.AddEntityWithId("owner", "owner", matrix.Vec3Zero())
	target := manager.AddEntityWithId("target", "target", matrix.NewVec3(0, 2, 0))
	return host, manager, owner, target
}

func setHostField[T any](host *engine.Host, name string, value T) {
	field := reflect.ValueOf(host).Elem().FieldByName(name)
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

func constraintTestEntry(target any) *entity_data_binding.EntityDataEntry {
	entry := entity_data_binding.ToDataBinding("", target)
	switch target.(type) {
	case *engine_entity_data_physics.DistanceJointEntityData:
		entry.Gen.RegisterKey = pod.QualifiedNameForLayout(engine_entity_data_physics.DistanceJointEntityData{})
	case *engine_entity_data_physics.RopeJointEntityData:
		entry.Gen.RegisterKey = pod.QualifiedNameForLayout(engine_entity_data_physics.RopeJointEntityData{})
	case *engine_entity_data_physics.PointJointEntityData:
		entry.Gen.RegisterKey = pod.QualifiedNameForLayout(engine_entity_data_physics.PointJointEntityData{})
	case *engine_entity_data_physics.HingeJointEntityData:
		entry.Gen.RegisterKey = pod.QualifiedNameForLayout(engine_entity_data_physics.HingeJointEntityData{})
	}
	return &entry
}
