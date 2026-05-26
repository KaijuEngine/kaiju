//go:build !debug

/******************************************************************************/
/* stage_entity_data_order_release_test.go                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stages

import (
	"bytes"
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/matrix"
)

var (
	stageOrderLog      []string
	stageOrderFailures []string
)

type stageOrderRecordData struct {
	Name string
}

func (d stageOrderRecordData) Init(e *engine.Entity, host *engine.Host) {
	stageOrderLog = append(stageOrderLog, d.Name)
}

type stageOrderLateData struct {
	Name string
}

func (d stageOrderLateData) Init(e *engine.Entity, host *engine.Host) {
	stageOrderLog = append(stageOrderLog, d.Name)
}

func (d stageOrderLateData) EntityDataInitPhase() engine.EntityDataPhase {
	return engine.EntityDataPhasePhysicsConstraint
}

type stageOrderConstraintData struct {
	Target string
}

func (d stageOrderConstraintData) Init(e *engine.Entity, host *engine.Host) {
	stageOrderLog = append(stageOrderLog, "constraint")
	target := host.EntityById(engine.EntityId(d.Target))
	if target == nil {
		stageOrderFailures = append(stageOrderFailures, "target entity was not registered")
		return
	}
	if _, ok := host.Physics().FindEntity(e); !ok {
		stageOrderFailures = append(stageOrderFailures, "source body was not staged")
		return
	}
	if _, ok := host.Physics().FindEntity(target); !ok {
		stageOrderFailures = append(stageOrderFailures, "target body was not staged")
		return
	}
	if host.Physics().AddDistanceJoint(e, target, matrix.Vec3Zero(), matrix.Vec3Zero()) == nil {
		stageOrderFailures = append(stageOrderFailures, "distance joint was not created")
	}
}

func (d stageOrderConstraintData) EntityDataInitPhase() engine.EntityDataPhase {
	return engine.EntityDataPhasePhysicsConstraint
}

func TestStageLoadRawDataBindingOrdersConstraintAfterLaterRigidBody(t *testing.T) {
	resetStageOrderTestState()
	stage := Stage{
		Entities: []EntityDescription{
			{
				Id: "source",
				RawDataBinding: []any{
					stageOrderConstraintData{Target: "target"},
					releaseTestRigidBodyData(),
				},
			},
			{
				Id: "target",
				RawDataBinding: []any{
					releaseTestRigidBodyData(),
				},
			},
		},
	}

	host := engine.NewHost("test", nil, nil)
	stage.Load(host)

	if len(stageOrderFailures) > 0 {
		t.Fatalf("expected constraint to find staged rigid bodies, got %v", stageOrderFailures)
	}
	if len(host.Physics().World().Constraints()) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(host.Physics().World().Constraints()))
	}
	if len(stageOrderLog) != 1 || stageOrderLog[0] != "constraint" {
		t.Fatalf("expected constraint init to run once, got %v", stageOrderLog)
	}
}

func TestStageLoadRawDataBindingPreservesSamePhaseBindingOrder(t *testing.T) {
	resetStageOrderTestState()
	stage := Stage{
		Entities: []EntityDescription{
			{
				Id: "entity",
				RawDataBinding: []any{
					stageOrderRecordData{Name: "first"},
					stageOrderLateData{Name: "late"},
					stageOrderRecordData{Name: "second"},
				},
			},
		},
	}

	stage.Load(engine.NewHost("test", nil, nil))

	want := []string{"first", "second", "late"}
	if !sameStrings(stageOrderLog, want) {
		t.Fatalf("expected init order %v, got %v", want, stageOrderLog)
	}
}

func TestStageArchiveDeserializerPreservesRawDataBindingPhaseOrder(t *testing.T) {
	resetStageOrderTestState()
	registerStageOrderArchiveTypes(t)
	stage := Stage{
		Entities: []EntityDescription{
			{
				Id: "source",
				RawDataBinding: []any{
					stageOrderConstraintData{Target: "target"},
					releaseTestRigidBodyData(),
				},
			},
			{
				Id: "target",
				RawDataBinding: []any{
					releaseTestRigidBodyData(),
				},
			},
		},
	}
	buf := bytes.Buffer{}
	if err := pod.NewEncoder(&buf).Encode(stage); err != nil {
		t.Fatalf("failed to encode stage archive: %v", err)
	}
	loaded, err := ArchiveDeserializer(buf.Bytes())
	if err != nil {
		t.Fatalf("failed to decode stage archive: %v", err)
	}

	host := engine.NewHost("test", nil, nil)
	loaded.Load(host)

	if len(stageOrderFailures) > 0 {
		t.Fatalf("expected archive-loaded constraint to find staged rigid bodies, got %v", stageOrderFailures)
	}
	if len(host.Physics().World().Constraints()) != 1 {
		t.Fatalf("expected 1 archive-loaded constraint, got %d", len(host.Physics().World().Constraints()))
	}
}

func releaseTestRigidBodyData() engine_entity_data_physics.RigidBodyEntityData {
	return engine_entity_data_physics.RigidBodyEntityData{
		Extent:   matrix.Vec3One(),
		IsStatic: true,
	}
}

func resetStageOrderTestState() {
	stageOrderLog = nil
	stageOrderFailures = nil
}

func registerStageOrderArchiveTypes(t *testing.T) {
	t.Helper()
	if err := pod.Register(stageOrderRecordData{}); err != nil {
		t.Fatalf("failed to register stageOrderRecordData: %v", err)
	}
	t.Cleanup(func() { pod.Unregister(stageOrderRecordData{}) })
	if err := pod.Register(stageOrderLateData{}); err != nil {
		t.Fatalf("failed to register stageOrderLateData: %v", err)
	}
	t.Cleanup(func() { pod.Unregister(stageOrderLateData{}) })
	if err := pod.Register(stageOrderConstraintData{}); err != nil {
		t.Fatalf("failed to register stageOrderConstraintData: %v", err)
	}
	t.Cleanup(func() { pod.Unregister(stageOrderConstraintData{}) })
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
