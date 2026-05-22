/******************************************************************************/
/* constraint_authoring_test.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"testing"

	"kaijuengine.com/editor/memento"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/matrix"
)

func TestConnectSelectedAsConstraintChainCreatesBindingsInHierarchyOrder(t *testing.T) {
	tests := []struct {
		name     string
		kind     ConstraintChainKind
		wantKey  string
		validate func(*testing.T, any, string)
	}{
		{
			name:    "distance",
			kind:    ConstraintChainDistance,
			wantKey: pod.QualifiedNameForLayout(engine_entity_data_physics.DistanceJointEntityData{}),
			validate: func(t *testing.T, data any, target string) {
				joint, ok := data.(*engine_entity_data_physics.DistanceJointEntityData)
				if !ok {
					t.Fatalf("expected distance data, got %T", data)
				}
				if joint.ConnectedEntityId != engine.EntityId(target) || !joint.AutoRestLength || !joint.Enabled {
					t.Fatalf("unexpected distance data: %#v", joint)
				}
			},
		},
		{
			name:    "rope",
			kind:    ConstraintChainRope,
			wantKey: pod.QualifiedNameForLayout(engine_entity_data_physics.RopeJointEntityData{}),
			validate: func(t *testing.T, data any, target string) {
				joint, ok := data.(*engine_entity_data_physics.RopeJointEntityData)
				if !ok {
					t.Fatalf("expected rope data, got %T", data)
				}
				if joint.ConnectedEntityId != engine.EntityId(target) || !joint.AutoMaxLength || !joint.Enabled {
					t.Fatalf("unexpected rope data: %#v", joint)
				}
			},
		},
		{
			name:    "hinge",
			kind:    ConstraintChainHinge,
			wantKey: pod.QualifiedNameForLayout(engine_entity_data_physics.HingeJointEntityData{}),
			validate: func(t *testing.T, data any, target string) {
				joint, ok := data.(*engine_entity_data_physics.HingeJointEntityData)
				if !ok {
					t.Fatalf("expected hinge data, got %T", data)
				}
				if joint.ConnectedEntityId != engine.EntityId(target) || !joint.Enabled {
					t.Fatalf("unexpected hinge data: %#v", joint)
				}
				if !matrix.Vec3ApproxTo(joint.HingeAxis, matrix.Vec3Backward(), 0.0001) {
					t.Fatalf("expected hinge axis %v, got %v", matrix.Vec3Backward(), joint.HingeAxis)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, first, second, third := newConstraintAuthoringStage(t)
			manager.SelectEntity(third)
			manager.SelectEntity(first)
			manager.SelectEntity(second)

			attachments := manager.ConnectSelectedAsConstraintChain(tt.kind)
			if len(attachments) != 2 {
				t.Fatalf("expected 2 attachments, got %d", len(attachments))
			}
			if attachments[0].Entity != first || attachments[1].Entity != second {
				t.Fatalf("expected hierarchy order first->second, got %s->%s",
					attachments[0].Entity.Name(), attachments[1].Entity.Name())
			}
			for i, attachment := range attachments {
				if attachment.Data.Gen.RegisterKey != tt.wantKey {
					t.Fatalf("expected key %q, got %q", tt.wantKey, attachment.Data.Gen.RegisterKey)
				}
				if got := attachment.Entity.DataBindings(); len(got) != 1 || got[0] != attachment.Data {
					t.Fatalf("expected attachment to be on source entity, got %#v", got)
				}
				target := []string{"second", "third"}[i]
				tt.validate(t, attachment.Data.BoundData, target)
			}
		})
	}
}

func TestConnectSelectedAsConstraintChainRequiresTwoEntities(t *testing.T) {
	manager, first, _, _ := newConstraintAuthoringStage(t)
	manager.SelectEntity(first)
	if got := manager.ConnectSelectedAsDistanceChain(); len(got) != 0 {
		t.Fatalf("expected no attachments for a single selection, got %d", len(got))
	}
}

func newConstraintAuthoringStage(t *testing.T) (*StageManager, *StageEntity, *StageEntity, *StageEntity) {
	t.Helper()
	host := engine.NewHost("constraint-authoring-test", nil, nil)
	history := &memento.History{}
	history.Initialize(64)
	manager := &StageManager{}
	manager.Initialize(host, history, nil)
	first := manager.AddEntityWithId("first", "first", matrix.NewVec3(0, 0, 0))
	second := manager.AddEntityWithId("second", "second", matrix.NewVec3(2, 0, 0))
	third := manager.AddEntityWithId("third", "third", matrix.NewVec3(4, 0, 0))
	return manager, first, second, third
}
