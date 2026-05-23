//go:build debug

/******************************************************************************/
/* stage_entity_data_order_debug_test.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stages

import (
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/matrix"
)

var debugStageOrderFailures []string

type debugStageOrderConstraintData struct {
	Target string
}

func (d debugStageOrderConstraintData) Init(e *engine.Entity, host *engine.Host) {
	target := host.EntityById(engine.EntityId(d.Target))
	if target == nil {
		debugStageOrderFailures = append(debugStageOrderFailures, "target entity was not registered")
		return
	}
	if _, ok := host.Physics().FindEntity(e); !ok {
		debugStageOrderFailures = append(debugStageOrderFailures, "source body was not staged")
		return
	}
	if _, ok := host.Physics().FindEntity(target); !ok {
		debugStageOrderFailures = append(debugStageOrderFailures, "target body was not staged")
		return
	}
	if host.Physics().AddDistanceJoint(e, target, matrix.Vec3Zero(), matrix.Vec3Zero()) == nil {
		debugStageOrderFailures = append(debugStageOrderFailures, "distance joint was not created")
	}
}

func (d debugStageOrderConstraintData) EntityDataInitPhase() engine.EntityDataPhase {
	return engine.EntityDataPhasePhysicsConstraint
}

func TestStageLoadDebugDataBindingOrdersConstraintAfterLaterRigidBody(t *testing.T) {
	debugStageOrderFailures = nil
	key := pod.QualifiedNameForLayout(debugStageOrderConstraintData{})
	if err := engine.RegisterEntityData(debugStageOrderConstraintData{}); err != nil {
		t.Fatalf("failed to register debug constraint data: %v", err)
	}
	t.Cleanup(func() {
		delete(engine.DebugEntityDataRegistry, key)
		pod.Unregister(debugStageOrderConstraintData{})
	})
	rigidBodyBinding := EntityDataBinding{
		RegistraionKey: engine_entity_data_physics.BindingKey(),
		Fields: map[string]any{
			"Extent":   []interface{}{1.0, 1.0, 1.0},
			"IsStatic": true,
		},
	}
	stage := Stage{
		Entities: []EntityDescription{
			{
				Id: "source",
				DataBinding: []EntityDataBinding{
					{
						RegistraionKey: key,
						Fields: map[string]any{
							"Target": "target",
						},
					},
					rigidBodyBinding,
				},
			},
			{
				Id: "target",
				DataBinding: []EntityDataBinding{
					rigidBodyBinding,
				},
			},
		},
	}

	host := engine.NewHost("test", nil, nil)
	stage.Load(host)

	if len(debugStageOrderFailures) > 0 {
		t.Fatalf("expected debug constraint to find staged rigid bodies, got %v", debugStageOrderFailures)
	}
	if len(host.Physics().World().Constraints()) != 1 {
		t.Fatalf("expected 1 debug-loaded constraint, got %d", len(host.Physics().World().Constraints()))
	}
}
