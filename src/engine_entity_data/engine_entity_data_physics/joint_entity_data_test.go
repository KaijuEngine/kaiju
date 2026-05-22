/******************************************************************************/
/* joint_entity_data_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_physics

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

func TestJointEntityDataCreatesCorrectGravitonJoints(t *testing.T) {
	tests := []struct {
		name     string
		init     func(*engine.Entity, *engine.Host)
		wantType graviton.ConstraintType
		assert   func(*testing.T, *graviton.Constraint)
	}{
		{
			name: "distance",
			init: func(e *engine.Entity, host *engine.Host) {
				DistanceJointEntityData{
					ConnectedEntityId: "target",
					LocalAnchorA:      matrix.NewVec3(1, 0, 0),
					TargetAnchorB:     matrix.NewVec3(0, 1, 0),
					Stiffness:         0.75,
					Bias:              0.25,
					Correction:        0.5,
					Slop:              0.02,
					MaxCorrection:     0.4,
					WarmStarting:      true,
					Enabled:           true,
					BreakForce:        7,
					BreakTorque:       8,
					RestLength:        4,
					AutoRestLength:    false,
				}.Init(e, host)
			},
			wantType: graviton.ConstraintTypeDistance,
			assert: func(t *testing.T, c *graviton.Constraint) {
				if c.Distance == nil {
					t.Fatal("expected distance joint")
				}
				if c.Distance.RestLength != 4 || c.Distance.Stiffness != 0.75 ||
					c.Distance.BiasFactor != 0.25 || c.Distance.PositionCorrectionFactor != 0.5 ||
					c.Distance.Slop != 0.02 || c.Distance.MaxCorrection != 0.4 ||
					!c.Distance.WarmStarting {
					t.Fatalf("distance joint fields were not applied: %#v", c.Distance)
				}
			},
		},
		{
			name: "rope",
			init: func(e *engine.Entity, host *engine.Host) {
				RopeJointEntityData{
					ConnectedEntityId: "target",
					LocalAnchorA:      matrix.NewVec3(1, 0, 0),
					TargetAnchorB:     matrix.NewVec3(0, 1, 0),
					Stiffness:         0.8,
					Bias:              0.3,
					Correction:        0.6,
					Slop:              0.03,
					MaxCorrection:     0.7,
					WarmStarting:      true,
					Enabled:           true,
					BreakForce:        7,
					BreakTorque:       8,
					MaxLength:         5,
					AutoMaxLength:     false,
				}.Init(e, host)
			},
			wantType: graviton.ConstraintTypeRope,
			assert: func(t *testing.T, c *graviton.Constraint) {
				if c.Rope == nil {
					t.Fatal("expected rope joint")
				}
				if c.Rope.MaxLength != 5 || c.Rope.Stiffness != 0.8 ||
					c.Rope.BiasFactor != 0.3 || c.Rope.PositionCorrectionFactor != 0.6 ||
					c.Rope.Slop != 0.03 || c.Rope.MaxCorrection != 0.7 ||
					!c.Rope.WarmStarting {
					t.Fatalf("rope joint fields were not applied: %#v", c.Rope)
				}
			},
		},
		{
			name: "point",
			init: func(e *engine.Entity, host *engine.Host) {
				PointJointEntityData{
					ConnectedEntityId: "target",
					LocalAnchorA:      matrix.NewVec3(1, 0, 0),
					TargetAnchorB:     matrix.NewVec3(0, 1, 0),
					Stiffness:         0.9,
					Bias:              0.35,
					Correction:        0.65,
					Slop:              0.04,
					MaxCorrection:     0.8,
					WarmStarting:      true,
					Enabled:           true,
					BreakForce:        7,
					BreakTorque:       8,
				}.Init(e, host)
			},
			wantType: graviton.ConstraintTypePoint,
			assert: func(t *testing.T, c *graviton.Constraint) {
				if c.Point == nil {
					t.Fatal("expected point joint")
				}
				if c.Point.Stiffness != 0.9 || c.Point.BiasFactor != 0.35 ||
					c.Point.PositionCorrectionFactor != 0.65 ||
					c.Point.Slop != 0.04 || c.Point.MaxCorrection != 0.8 ||
					!c.Point.WarmStarting {
					t.Fatalf("point joint fields were not applied: %#v", c.Point)
				}
			},
		},
		{
			name: "hinge",
			init: func(e *engine.Entity, host *engine.Host) {
				HingeJointEntityData{
					ConnectedEntityId: "target",
					LocalAnchorA:      matrix.NewVec3(1, 0, 0),
					TargetAnchorB:     matrix.NewVec3(0, 1, 0),
					Stiffness:         0.95,
					Bias:              0.4,
					Correction:        0.7,
					Slop:              0.05,
					MaxCorrection:     0.9,
					WarmStarting:      true,
					Enabled:           true,
					BreakForce:        7,
					BreakTorque:       8,
					HingeAxis:         matrix.Vec3Up(),
					EnableLimits:      true,
					MinAngleDegrees:   -45,
					MaxAngleDegrees:   30,
					EnableMotor:       true,
					MotorSpeedDegrees: 90,
					MaxMotorTorque:    12,
					MaxMotorImpulse:   3,
				}.Init(e, host)
			},
			wantType: graviton.ConstraintTypeHinge,
			assert: func(t *testing.T, c *graviton.Constraint) {
				if c.Hinge == nil {
					t.Fatal("expected hinge joint")
				}
				if c.Hinge.Stiffness != 0.95 || c.Hinge.BiasFactor != 0.4 ||
					c.Hinge.PositionCorrectionFactor != 0.7 ||
					c.Hinge.Slop != 0.05 || c.Hinge.MaxCorrection != 0.9 ||
					!c.Hinge.WarmStarting || !c.Hinge.EnableLimits ||
					!c.Hinge.EnableMotor {
					t.Fatalf("hinge joint fields were not applied: %#v", c.Hinge)
				}
				if !matrix.Vec3ApproxTo(c.Hinge.LocalAxisA, matrix.Vec3Up(), 0.0001) ||
					!matrix.Vec3ApproxTo(c.Hinge.LocalAxisB, matrix.Vec3Up(), 0.0001) {
					t.Fatalf("expected hinge axes to use data axis, got %v and %v",
						c.Hinge.LocalAxisA, c.Hinge.LocalAxisB)
				}
				if matrix.Abs(c.Hinge.MinAngle-matrix.Deg2Rad(-45)) > 0.0001 ||
					matrix.Abs(c.Hinge.MaxAngle-matrix.Deg2Rad(30)) > 0.0001 ||
					matrix.Abs(c.Hinge.MotorTargetSpeed-matrix.Deg2Rad(90)) > 0.0001 ||
					c.Hinge.MaxMotorTorque != 12 {
					t.Fatalf("expected hinge degrees to convert to radians, got %#v", c.Hinge)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, source, _ := jointTestHostWithBodies(t, true)
			tt.init(source, host)
			constraints := host.Physics().World().Constraints()
			if len(constraints) != 1 {
				t.Fatalf("expected 1 constraint, got %d", len(constraints))
			}
			constraint := constraints[0]
			if constraint.Type != tt.wantType {
				t.Fatalf("expected constraint type %d, got %d", tt.wantType, constraint.Type)
			}
			if !constraint.Enabled || !constraint.Active || constraint.BreakForce != 7 || constraint.BreakTorque != 8 {
				t.Fatalf("expected common constraint fields to be applied, got %#v", constraint)
			}
			if len(source.NamedData(PhysicsJointNamedData)) != 1 {
				t.Fatal("expected joint to be stored as named data")
			}
			if got := source.NamedData(PhysicsConstraintNamedData); len(got) != 1 || got[0] != constraint {
				t.Fatal("expected constraint to be stored as named data")
			}
			tt.assert(t, constraint)
		})
	}
}

func TestJointEntityDataEmptyTargetCreatesBodyWorldJoint(t *testing.T) {
	tests := []struct {
		name string
		init func(*engine.Entity, *engine.Host)
	}{
		{"distance", func(e *engine.Entity, host *engine.Host) {
			DistanceJointEntityData{Enabled: true, AutoRestLength: true}.Init(e, host)
		}},
		{"rope", func(e *engine.Entity, host *engine.Host) {
			RopeJointEntityData{Enabled: true, AutoMaxLength: true}.Init(e, host)
		}},
		{"point", func(e *engine.Entity, host *engine.Host) {
			PointJointEntityData{Enabled: true}.Init(e, host)
		}},
		{"hinge", func(e *engine.Entity, host *engine.Host) {
			HingeJointEntityData{Enabled: true, HingeAxis: matrix.Vec3Right()}.Init(e, host)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, source, _ := jointTestHostWithBodies(t, false)
			tt.init(source, host)
			constraints := host.Physics().World().Constraints()
			if len(constraints) != 1 {
				t.Fatalf("expected 1 constraint, got %d", len(constraints))
			}
			if !constraints[0].IsBodyWorld() {
				t.Fatalf("expected body-world joint, got bodyA %p bodyB %p",
					constraints[0].BodyA, constraints[0].BodyB)
			}
		})
	}
}

func TestDistanceJointEntityDataAutoRestLengthUsesInitialBodyWorldOffset(t *testing.T) {
	host := engine.NewHost("test", nil, nil)
	source := engine.NewEntity(host.WorkGroup())
	host.SetEntityId(source, "source")
	source.Transform.SetPosition(matrix.NewVec3(2, 0, 0))
	addJointTestRigidBody(source, host)

	worldAnchor := matrix.NewVec3(0, 2, 0)
	DistanceJointEntityData{
		LocalAnchorA:   matrix.Vec3Zero(),
		TargetAnchorB:  worldAnchor,
		Stiffness:      1,
		Bias:           0.2,
		Correction:     0.8,
		Slop:           0.001,
		MaxCorrection:  0.5,
		Enabled:        true,
		AutoRestLength: true,
	}.Init(source, host)

	constraints := host.Physics().World().Constraints()
	if len(constraints) != 1 || constraints[0].Distance == nil {
		t.Fatalf("expected one body-world distance joint, got %#v", constraints)
	}
	joint := constraints[0].Distance
	wantRestLength := source.Transform.WorldPosition().Distance(worldAnchor)
	if matrix.Abs(joint.RestLength-wantRestLength) > 0.0001 {
		t.Fatalf("expected auto rest length %f, got %f", wantRestLength, joint.RestLength)
	}
	if !matrix.Vec3ApproxTo(joint.WorldAnchorA(), source.Transform.WorldPosition(), 0.0001) {
		t.Fatalf("expected local anchor A to start at body position %v, got %v",
			source.Transform.WorldPosition(), joint.WorldAnchorA())
	}
	if !matrix.Vec3ApproxTo(joint.WorldAnchorB(), worldAnchor, 0.0001) {
		t.Fatalf("expected target anchor B to be fixed world anchor %v, got %v",
			worldAnchor, joint.WorldAnchorB())
	}
}

func TestJointEntityDataMissingBodyOrTargetDoesNotPanic(t *testing.T) {
	host := engine.NewHost("test", nil, nil)
	source := engine.NewEntity(host.WorkGroup())
	host.SetEntityId(source, "source")

	PointJointEntityData{
		ConnectedEntityId: "missing",
		Enabled:           true,
	}.Init(source, host)
	if len(host.Physics().World().Constraints()) != 0 {
		t.Fatalf("expected missing target to skip joint creation")
	}

	PointJointEntityData{Enabled: true}.Init(source, host)
	if len(host.Physics().World().Constraints()) != 0 {
		t.Fatalf("expected missing source rigid body to skip joint creation")
	}

	target := engine.NewEntity(host.WorkGroup())
	host.SetEntityId(target, "target")
	addJointTestRigidBody(source, host)
	PointJointEntityData{
		ConnectedEntityId: "target",
		Enabled:           true,
	}.Init(source, host)
	if len(host.Physics().World().Constraints()) != 0 {
		t.Fatalf("expected missing target rigid body to skip joint creation")
	}
}

func TestJointEntityDataRoundTripsJSONAndArchivedPODStages(t *testing.T) {
	values := []any{
		DistanceJointEntityData{
			ConnectedEntityId: "target",
			LocalAnchorA:      matrix.NewVec3(1, 2, 3),
			TargetAnchorB:     matrix.NewVec3(4, 5, 6),
			Stiffness:         0.1,
			Bias:              0.2,
			Correction:        0.3,
			Slop:              0.4,
			MaxCorrection:     0.5,
			WarmStarting:      true,
			Enabled:           true,
			BreakForce:        6,
			BreakTorque:       7,
			RestLength:        8,
			AutoRestLength:    true,
		},
		RopeJointEntityData{
			ConnectedEntityId: "target",
			LocalAnchorA:      matrix.NewVec3(2, 3, 4),
			TargetAnchorB:     matrix.NewVec3(5, 6, 7),
			Stiffness:         0.11,
			Bias:              0.22,
			Correction:        0.33,
			Slop:              0.44,
			MaxCorrection:     0.55,
			WarmStarting:      true,
			Enabled:           true,
			BreakForce:        6,
			BreakTorque:       7,
			MaxLength:         9,
			AutoMaxLength:     true,
		},
		PointJointEntityData{
			ConnectedEntityId: "target",
			LocalAnchorA:      matrix.NewVec3(3, 4, 5),
			TargetAnchorB:     matrix.NewVec3(6, 7, 8),
			Stiffness:         0.12,
			Bias:              0.23,
			Correction:        0.34,
			Slop:              0.45,
			MaxCorrection:     0.56,
			WarmStarting:      true,
			Enabled:           true,
			BreakForce:        6,
			BreakTorque:       7,
		},
		HingeJointEntityData{
			ConnectedEntityId: "target",
			LocalAnchorA:      matrix.NewVec3(4, 5, 6),
			TargetAnchorB:     matrix.NewVec3(7, 8, 9),
			Stiffness:         0.13,
			Bias:              0.24,
			Correction:        0.35,
			Slop:              0.46,
			MaxCorrection:     0.57,
			WarmStarting:      true,
			Enabled:           true,
			BreakForce:        6,
			BreakTorque:       7,
			HingeAxis:         matrix.Vec3Up(),
			EnableLimits:      true,
			MinAngleDegrees:   -20,
			MaxAngleDegrees:   40,
			EnableMotor:       true,
			MotorSpeedDegrees: 80,
			MaxMotorTorque:    10,
			MaxMotorImpulse:   11,
		},
	}

	for _, value := range values {
		t.Run(reflect.TypeOf(value).Name(), func(t *testing.T) {
			data, err := json.Marshal(value)
			if err != nil {
				t.Fatalf("failed to marshal JSON: %v", err)
			}
			decoded := reflect.New(reflect.TypeOf(value)).Interface()
			if err := json.Unmarshal(data, decoded); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}
			if !reflect.DeepEqual(reflect.ValueOf(decoded).Elem().Interface(), value) {
				t.Fatalf("JSON round trip mismatch: got %#v want %#v",
					reflect.ValueOf(decoded).Elem().Interface(), value)
			}

			stage := stages.Stage{
				Entities: []stages.EntityDescription{
					{
						Id:             "entity",
						RawDataBinding: []any{value},
					},
				},
			}
			buf := bytes.Buffer{}
			if err := pod.NewEncoder(&buf).Encode(stage); err != nil {
				t.Fatalf("failed to encode POD stage: %v", err)
			}
			loaded, err := stages.ArchiveDeserializer(buf.Bytes())
			if err != nil {
				t.Fatalf("failed to decode POD stage: %v", err)
			}
			if len(loaded.Entities) != 1 || len(loaded.Entities[0].RawDataBinding) != 1 {
				t.Fatalf("expected archived stage binding to round trip, got %#v", loaded)
			}
			if !reflect.DeepEqual(loaded.Entities[0].RawDataBinding[0], value) {
				t.Fatalf("POD stage round trip mismatch: got %#v want %#v",
					loaded.Entities[0].RawDataBinding[0], value)
			}
		})
	}
}

func TestJointEntityDataBreakThresholdsDisableConstraints(t *testing.T) {
	t.Run("break force", func(t *testing.T) {
		host, source, _ := jointTestHostWithBodies(t, false)
		PointJointEntityData{
			Stiffness:     1,
			Bias:          0.2,
			Correction:    0.8,
			Slop:          0.001,
			MaxCorrection: 0.5,
			Enabled:       true,
			BreakForce:    0.1,
		}.Init(source, host)
		entry, _ := host.Physics().FindEntity(source)
		entry.Body.MotionState.LinearVelocity = matrix.Vec3Right().Scale(20)

		stepHostPhysics(t, host)

		constraint := host.Physics().World().Constraints()[0]
		if !constraint.Broken || constraint.Enabled || constraint.Active {
			t.Fatalf("expected break force to disable constraint, got %#v", constraint)
		}
	})

	t.Run("break torque", func(t *testing.T) {
		host, source, _ := jointTestHostWithBodies(t, false)
		HingeJointEntityData{
			Stiffness:     1,
			Bias:          0.2,
			Correction:    0.8,
			Slop:          0.001,
			MaxCorrection: 0.5,
			Enabled:       true,
			BreakTorque:   0.1,
			HingeAxis:     matrix.Vec3Up(),
		}.Init(source, host)
		entry, _ := host.Physics().FindEntity(source)
		entry.Body.MotionState.AngularVelocity = matrix.Vec3Right().Scale(20)

		stepHostPhysics(t, host)

		constraint := host.Physics().World().Constraints()[0]
		if !constraint.Broken || constraint.Enabled || constraint.Active {
			t.Fatalf("expected break torque to disable constraint, got %#v", constraint)
		}
	})
}

func jointTestHostWithBodies(t *testing.T, includeTarget bool) (*engine.Host, *engine.Entity, *engine.Entity) {
	t.Helper()
	host := engine.NewHost("test", nil, nil)
	source := engine.NewEntity(host.WorkGroup())
	host.SetEntityId(source, "source")
	addJointTestRigidBody(source, host)
	var target *engine.Entity
	if includeTarget {
		target = engine.NewEntity(host.WorkGroup())
		host.SetEntityId(target, "target")
		target.Transform.SetPosition(matrix.NewVec3(2, 0, 0))
		addJointTestRigidBody(target, host)
	}
	return host, source, target
}

func addJointTestRigidBody(entity *engine.Entity, host *engine.Host) {
	RigidBodyEntityData{
		Extent:   matrix.Vec3One(),
		Mass:     1,
		Radius:   1,
		Height:   1,
		Shape:    ShapeSphere,
		IsStatic: false,
	}.Init(entity, host)
}

func stepHostPhysics(t *testing.T, host *engine.Host) {
	t.Helper()
	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()
	host.Physics().SetMaxSubSteps(1)
	host.Physics().Update(&workGroup, &threads, host.Physics().FixedTimeStep())
}
