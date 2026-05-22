/******************************************************************************/
/* joint_entity_data.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_physics

import (
	"log/slog"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

const (
	PhysicsJointNamedData      = "PhysicsJoint"
	PhysicsConstraintNamedData = "PhysicsConstraint"
)

type jointEntityDataCommon struct {
	ConnectedEntityId engine.EntityId
	LocalAnchorA      matrix.Vec3
	TargetAnchorB     matrix.Vec3
	Stiffness         matrix.Float
	Bias              matrix.Float
	Correction        matrix.Float
	Slop              matrix.Float
	MaxCorrection     matrix.Float
	WarmStarting      bool
	Enabled           bool
	BreakForce        matrix.Float
	BreakTorque       matrix.Float
}

type DistanceJointEntityData struct {
	ConnectedEntityId engine.EntityId
	LocalAnchorA      matrix.Vec3  // Local body anchor on this entity.
	TargetAnchorB     matrix.Vec3  // Local target anchor, or fixed world anchor when ConnectedEntityId is empty.
	Stiffness         matrix.Float `default:"1"`
	Bias              matrix.Float `default:"0.2"`
	Correction        matrix.Float `default:"0.8"`
	Slop              matrix.Float `default:"0.001"`
	MaxCorrection     matrix.Float `default:"0.5"`
	WarmStarting      bool
	Enabled           bool `default:"true"`
	BreakForce        matrix.Float
	BreakTorque       matrix.Float
	RestLength        matrix.Float
	AutoRestLength    bool `default:"true"`
}

type RopeJointEntityData struct {
	ConnectedEntityId engine.EntityId
	LocalAnchorA      matrix.Vec3  // Local body anchor on this entity.
	TargetAnchorB     matrix.Vec3  // Local target anchor, or fixed world anchor when ConnectedEntityId is empty.
	Stiffness         matrix.Float `default:"1"`
	Bias              matrix.Float `default:"0.2"`
	Correction        matrix.Float `default:"0.8"`
	Slop              matrix.Float `default:"0.001"`
	MaxCorrection     matrix.Float `default:"0.5"`
	WarmStarting      bool
	Enabled           bool `default:"true"`
	BreakForce        matrix.Float
	BreakTorque       matrix.Float
	MaxLength         matrix.Float
	AutoMaxLength     bool `default:"true"`
}

type PointJointEntityData struct {
	ConnectedEntityId engine.EntityId
	LocalAnchorA      matrix.Vec3  // Local body anchor on this entity; point joints keep this coincident with TargetAnchorB.
	TargetAnchorB     matrix.Vec3  // Local target anchor, or fixed world anchor when ConnectedEntityId is empty.
	Stiffness         matrix.Float `default:"1"`
	Bias              matrix.Float `default:"0.2"`
	Correction        matrix.Float `default:"0.8"`
	Slop              matrix.Float `default:"0.001"`
	MaxCorrection     matrix.Float `default:"0.5"`
	WarmStarting      bool
	Enabled           bool `default:"true"`
	BreakForce        matrix.Float
	BreakTorque       matrix.Float
}

type HingeJointEntityData struct {
	ConnectedEntityId engine.EntityId
	LocalAnchorA      matrix.Vec3  // Local body anchor on this entity; hinge joints keep this coincident with TargetAnchorB.
	TargetAnchorB     matrix.Vec3  // Local target anchor, or fixed world anchor when ConnectedEntityId is empty.
	Stiffness         matrix.Float `default:"1"`
	Bias              matrix.Float `default:"0.2"`
	Correction        matrix.Float `default:"0.8"`
	Slop              matrix.Float `default:"0.001"`
	MaxCorrection     matrix.Float `default:"0.5"`
	WarmStarting      bool
	Enabled           bool `default:"true"`
	BreakForce        matrix.Float
	BreakTorque       matrix.Float
	HingeAxis         matrix.Vec3 `default:"1,0,0"`
	EnableLimits      bool
	MinAngleDegrees   matrix.Float
	MaxAngleDegrees   matrix.Float
	EnableMotor       bool
	MotorSpeedDegrees matrix.Float
	MaxMotorTorque    matrix.Float
	MaxMotorImpulse   matrix.Float
}

func init() {
	pod.Register(engine.EntityId(""))
	engine.RegisterEntityData(DistanceJointEntityData{})
	engine.RegisterEntityData(RopeJointEntityData{})
	engine.RegisterEntityData(PointJointEntityData{})
	engine.RegisterEntityData(HingeJointEntityData{})
}

func (d DistanceJointEntityData) Init(e *engine.Entity, host *engine.Host) {
	host.StartPhysics()
	target, ok := d.common().targetEntity(host)
	if !ok {
		return
	}
	joint := host.Physics().AddDistanceJoint(e, target, d.LocalAnchorA, d.TargetAnchorB)
	if joint == nil {
		return
	}
	d.common().applyDistance(joint)
	if !d.AutoRestLength {
		joint.SetRestLength(d.RestLength)
	}
	storeJoint(e, joint, joint.Constraint())
}

func (d DistanceJointEntityData) EntityDataInitPhase() engine.EntityDataPhase {
	return engine.EntityDataPhasePhysicsConstraint
}

func (d DistanceJointEntityData) common() jointEntityDataCommon {
	return jointEntityDataCommon{
		ConnectedEntityId: d.ConnectedEntityId,
		LocalAnchorA:      d.LocalAnchorA,
		TargetAnchorB:     d.TargetAnchorB,
		Stiffness:         d.Stiffness,
		Bias:              d.Bias,
		Correction:        d.Correction,
		Slop:              d.Slop,
		MaxCorrection:     d.MaxCorrection,
		WarmStarting:      d.WarmStarting,
		Enabled:           d.Enabled,
		BreakForce:        d.BreakForce,
		BreakTorque:       d.BreakTorque,
	}
}

func (d RopeJointEntityData) Init(e *engine.Entity, host *engine.Host) {
	host.StartPhysics()
	target, ok := d.common().targetEntity(host)
	if !ok {
		return
	}
	joint := host.Physics().AddRopeJoint(e, target, d.LocalAnchorA, d.TargetAnchorB)
	if joint == nil {
		return
	}
	d.common().applyRope(joint)
	if !d.AutoMaxLength {
		joint.SetMaxLength(d.MaxLength)
	}
	storeJoint(e, joint, joint.Constraint())
}

func (d RopeJointEntityData) EntityDataInitPhase() engine.EntityDataPhase {
	return engine.EntityDataPhasePhysicsConstraint
}

func (d RopeJointEntityData) common() jointEntityDataCommon {
	return jointEntityDataCommon{
		ConnectedEntityId: d.ConnectedEntityId,
		LocalAnchorA:      d.LocalAnchorA,
		TargetAnchorB:     d.TargetAnchorB,
		Stiffness:         d.Stiffness,
		Bias:              d.Bias,
		Correction:        d.Correction,
		Slop:              d.Slop,
		MaxCorrection:     d.MaxCorrection,
		WarmStarting:      d.WarmStarting,
		Enabled:           d.Enabled,
		BreakForce:        d.BreakForce,
		BreakTorque:       d.BreakTorque,
	}
}

func (d PointJointEntityData) Init(e *engine.Entity, host *engine.Host) {
	host.StartPhysics()
	target, ok := d.common().targetEntity(host)
	if !ok {
		return
	}
	joint := host.Physics().AddPointJoint(e, target, d.LocalAnchorA, d.TargetAnchorB)
	if joint == nil {
		return
	}
	d.common().applyPoint(joint)
	storeJoint(e, joint, joint.Constraint())
}

func (d PointJointEntityData) EntityDataInitPhase() engine.EntityDataPhase {
	return engine.EntityDataPhasePhysicsConstraint
}

func (d PointJointEntityData) common() jointEntityDataCommon {
	return jointEntityDataCommon{
		ConnectedEntityId: d.ConnectedEntityId,
		LocalAnchorA:      d.LocalAnchorA,
		TargetAnchorB:     d.TargetAnchorB,
		Stiffness:         d.Stiffness,
		Bias:              d.Bias,
		Correction:        d.Correction,
		Slop:              d.Slop,
		MaxCorrection:     d.MaxCorrection,
		WarmStarting:      d.WarmStarting,
		Enabled:           d.Enabled,
		BreakForce:        d.BreakForce,
		BreakTorque:       d.BreakTorque,
	}
}

func (d HingeJointEntityData) Init(e *engine.Entity, host *engine.Host) {
	host.StartPhysics()
	target, ok := d.common().targetEntity(host)
	if !ok {
		return
	}
	axis := d.HingeAxis
	if axis.LengthSquared() <= matrix.FloatSmallestNonzero {
		axis = matrix.Vec3Right()
	}
	joint := host.Physics().AddHingeJoint(e, target, d.LocalAnchorA, d.TargetAnchorB, axis, axis)
	if joint == nil {
		return
	}
	d.common().applyHinge(joint)
	if d.EnableLimits {
		joint.SetAngularLimits(matrix.Deg2Rad(d.MinAngleDegrees), matrix.Deg2Rad(d.MaxAngleDegrees))
	} else {
		joint.DisableAngularLimits()
	}
	if d.EnableMotor {
		speed := matrix.Deg2Rad(d.MotorSpeedDegrees)
		if d.MaxMotorTorque > 0 {
			joint.SetMotorTorque(speed, d.MaxMotorTorque)
		} else {
			joint.SetMotor(speed, d.MaxMotorImpulse)
		}
	} else {
		joint.DisableMotor()
	}
	storeJoint(e, joint, joint.Constraint())
}

func (d HingeJointEntityData) EntityDataInitPhase() engine.EntityDataPhase {
	return engine.EntityDataPhasePhysicsConstraint
}

func (d HingeJointEntityData) common() jointEntityDataCommon {
	return jointEntityDataCommon{
		ConnectedEntityId: d.ConnectedEntityId,
		LocalAnchorA:      d.LocalAnchorA,
		TargetAnchorB:     d.TargetAnchorB,
		Stiffness:         d.Stiffness,
		Bias:              d.Bias,
		Correction:        d.Correction,
		Slop:              d.Slop,
		MaxCorrection:     d.MaxCorrection,
		WarmStarting:      d.WarmStarting,
		Enabled:           d.Enabled,
		BreakForce:        d.BreakForce,
		BreakTorque:       d.BreakTorque,
	}
}

func (d jointEntityDataCommon) targetEntity(host *engine.Host) (*engine.Entity, bool) {
	if d.ConnectedEntityId == "" {
		return nil, true
	}
	target := host.EntityById(d.ConnectedEntityId)
	if target == nil {
		slog.Error("failed to add entity physics joint, connected entity was not found",
			"connectedEntityId", d.ConnectedEntityId)
		return nil, false
	}
	return target, true
}

func (d jointEntityDataCommon) applyDistance(joint *graviton.DistanceJoint) {
	joint.Stiffness = d.Stiffness
	joint.BiasFactor = d.Bias
	joint.PositionCorrectionFactor = d.Correction
	joint.Slop = d.Slop
	joint.MaxCorrection = d.MaxCorrection
	joint.WarmStarting = d.WarmStarting
	d.applyConstraint(joint.Constraint())
}

func (d jointEntityDataCommon) applyRope(joint *graviton.RopeJoint) {
	joint.Stiffness = d.Stiffness
	joint.BiasFactor = d.Bias
	joint.PositionCorrectionFactor = d.Correction
	joint.Slop = d.Slop
	joint.MaxCorrection = d.MaxCorrection
	joint.WarmStarting = d.WarmStarting
	d.applyConstraint(joint.Constraint())
}

func (d jointEntityDataCommon) applyPoint(joint *graviton.PointJoint) {
	joint.Stiffness = d.Stiffness
	joint.BiasFactor = d.Bias
	joint.PositionCorrectionFactor = d.Correction
	joint.Slop = d.Slop
	joint.MaxCorrection = d.MaxCorrection
	joint.WarmStarting = d.WarmStarting
	d.applyConstraint(joint.Constraint())
}

func (d jointEntityDataCommon) applyHinge(joint *graviton.HingeJoint) {
	joint.Stiffness = d.Stiffness
	joint.BiasFactor = d.Bias
	joint.PositionCorrectionFactor = d.Correction
	joint.Slop = d.Slop
	joint.MaxCorrection = d.MaxCorrection
	joint.WarmStarting = d.WarmStarting
	d.applyConstraint(joint.Constraint())
}

func (d jointEntityDataCommon) applyConstraint(constraint *graviton.Constraint) {
	if constraint == nil {
		return
	}
	constraint.SetEnabled(d.Enabled)
	constraint.SetBreakForce(d.BreakForce)
	constraint.SetBreakTorque(d.BreakTorque)
}

func storeJoint(e *engine.Entity, joint any, constraint *graviton.Constraint) {
	e.AddNamedData(PhysicsJointNamedData, joint)
	if constraint != nil {
		e.AddNamedData(PhysicsConstraintNamedData, constraint)
	}
}
