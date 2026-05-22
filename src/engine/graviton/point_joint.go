/******************************************************************************/
/* point_joint.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

var pointJointAxes = [3]matrix.Vec3{
	matrix.Vec3Right(),
	matrix.Vec3Up(),
	matrix.Vec3{0, 0, 1},
}

// PointJoint keeps two local anchors coincident while leaving relative
// orientation unconstrained. Nil bodies are treated as fixed world anchors.
type PointJoint struct {
	BodyA                    *RigidBody
	BodyB                    *RigidBody
	LocalAnchorA             matrix.Vec3
	LocalAnchorB             matrix.Vec3
	Stiffness                matrix.Float
	BiasFactor               matrix.Float
	PositionCorrectionFactor matrix.Float
	Slop                     matrix.Float
	MaxCorrection            matrix.Float
	WarmStarting             bool
	AccumulatedImpulse       matrix.Vec3
	constraint               *Constraint
	rows                     [3]ConstraintSolverRow
}

func NewPointJoint(bodyA, bodyB *RigidBody, localAnchorA, localAnchorB matrix.Vec3) *PointJoint {
	return &PointJoint{
		BodyA:                    bodyA,
		BodyB:                    bodyB,
		LocalAnchorA:             localAnchorA,
		LocalAnchorB:             localAnchorB,
		Stiffness:                defaultDistanceJointStiffness,
		BiasFactor:               defaultDistanceJointBiasFactor,
		PositionCorrectionFactor: defaultDistanceJointPositionCorrectionFactor,
		Slop:                     defaultDistanceJointSlop,
		MaxCorrection:            defaultDistanceJointMaxCorrection,
	}
}

func NewPointJointAtWorldAnchors(bodyA, bodyB *RigidBody, worldAnchorA, worldAnchorB matrix.Vec3) *PointJoint {
	return NewPointJoint(
		bodyA,
		bodyB,
		LocalAnchor(bodyA, worldAnchorA),
		LocalAnchor(bodyB, worldAnchorB),
	)
}

func NewPointJointToWorld(body *RigidBody, localAnchor, worldAnchor matrix.Vec3) *PointJoint {
	return NewPointJoint(body, nil, localAnchor, worldAnchor)
}

func (j *PointJoint) WorldAnchorA() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return WorldAnchor(j.BodyA, j.LocalAnchorA)
}

func (j *PointJoint) WorldAnchorB() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return WorldAnchor(j.BodyB, j.LocalAnchorB)
}

func (j *PointJoint) CurrentError() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return j.WorldAnchorB().Subtract(j.WorldAnchorA())
}

func (j *PointJoint) SetWorldAnchors(worldAnchorA, worldAnchorB matrix.Vec3) {
	if j == nil {
		return
	}
	j.LocalAnchorA = LocalAnchor(j.BodyA, worldAnchorA)
	j.LocalAnchorB = LocalAnchor(j.BodyB, worldAnchorB)
	j.AccumulatedImpulse = matrix.Vec3Zero()
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *PointJoint) Constraint() *Constraint {
	if j == nil {
		return nil
	}
	return j.constraint
}

func (j *PointJoint) IsStretched() bool {
	return j != nil && j.CurrentError().Length() > j.slop()
}

func (j *PointJoint) prepare(deltaTime matrix.Float) {
	if j == nil {
		return
	}
	anchorA := j.WorldAnchorA()
	anchorB := j.WorldAnchorB()
	error := anchorB.Subtract(anchorA)
	for i, axis := range pointJointAxes {
		row := &j.rows[i]
		row.SetWorldAnchors(j.BodyA, j.BodyB, anchorA, anchorB, axis)
		row.EffectiveMass *= j.stiffness()
		row.Bias = j.bias(error.Dot(axis), deltaTime)
		row.AccumulatedImpulse = 0
		if j.WarmStarting {
			row.AccumulatedImpulse = j.AccumulatedImpulse[i]
			row.ApplyImpulse(row.AccumulatedImpulse)
		}
	}
}

func (j *PointJoint) solveVelocity() {
	if j == nil {
		return
	}
	for i := range j.rows {
		j.rows[i].Solve()
		j.AccumulatedImpulse[i] = j.rows[i].AccumulatedImpulse
	}
}

func (j *PointJoint) solvePosition() {
	if j == nil {
		return
	}
	error := j.CurrentError()
	if error.Length() <= j.slop() {
		return
	}
	invMassA := j.BodyA.inverseMass()
	invMassB := j.BodyB.inverseMass()
	invMassSum := invMassA + invMassB
	if invMassSum <= contactEpsilon {
		return
	}
	correction := j.clampedCorrection(error)
	correction = correction.Scale(1.0 / invMassSum)
	moveBody(j.BodyA, correction.Scale(invMassA))
	moveBody(j.BodyB, correction.Scale(-invMassB))
}

func (j *PointJoint) bias(error, deltaTime matrix.Float) matrix.Float {
	if deltaTime <= 0 {
		deltaTime = defaultDistanceJointTimeStep
	}
	if matrix.Abs(error) <= j.slop() {
		return 0
	}
	return error * j.biasFactor() / deltaTime
}

func (j *PointJoint) clampedCorrection(error matrix.Vec3) matrix.Vec3 {
	correction := error.Scale(j.positionCorrectionFactor() * j.stiffness())
	maxCorrection := j.maxCorrection()
	length := correction.Length()
	if length > maxCorrection && length > matrix.FloatSmallestNonzero {
		correction = correction.Scale(maxCorrection / length)
	}
	return correction
}

func (j *PointJoint) stiffness() matrix.Float {
	if j.Stiffness < 0 {
		return 0
	}
	return matrix.Clamp(j.Stiffness, 0, 1)
}

func (j *PointJoint) biasFactor() matrix.Float {
	if j.BiasFactor < 0 {
		return 0
	}
	return j.BiasFactor
}

func (j *PointJoint) positionCorrectionFactor() matrix.Float {
	if j.PositionCorrectionFactor < 0 {
		return 0
	}
	return j.PositionCorrectionFactor
}

func (j *PointJoint) slop() matrix.Float {
	if j.Slop <= 0 {
		return defaultDistanceJointSlop
	}
	return j.Slop
}

func (j *PointJoint) maxCorrection() matrix.Float {
	if j.MaxCorrection <= 0 {
		return defaultDistanceJointMaxCorrection
	}
	return j.MaxCorrection
}
