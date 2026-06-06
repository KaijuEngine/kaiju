/******************************************************************************/
/* distance_joint.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

const (
	defaultDistanceJointStiffness                = matrix.Float(1)
	defaultDistanceJointBiasFactor               = matrix.Float(0.2)
	defaultDistanceJointPositionCorrectionFactor = matrix.Float(0.8)
	defaultDistanceJointSlop                     = matrix.Float(0.001)
	defaultDistanceJointMaxCorrection            = matrix.Float(0.5)
	defaultDistanceJointTimeStep                 = matrix.Float(1.0 / 60.0)
	distanceJointMinLength                       = matrix.Float(0.000001)
)

// DistanceJoint keeps two body anchors, or one body anchor and one fixed world
// anchor, at a target distance. Nil bodies are treated as fixed world anchors.
type DistanceJoint struct {
	BodyA                    *RigidBody
	BodyB                    *RigidBody
	LocalAnchorA             matrix.Vec3
	LocalAnchorB             matrix.Vec3
	RestLength               matrix.Float
	Stiffness                matrix.Float
	BiasFactor               matrix.Float
	PositionCorrectionFactor matrix.Float
	Slop                     matrix.Float
	MaxCorrection            matrix.Float
	WarmStarting             bool
	AccumulatedImpulse       matrix.Float
	constraint               *Constraint
	row                      ConstraintSolverRow
	lastAxis                 matrix.Vec3
}

func NewDistanceJoint(bodyA, bodyB *RigidBody, localAnchorA, localAnchorB matrix.Vec3) *DistanceJoint {
	joint := &DistanceJoint{
		BodyA:                    bodyA,
		BodyB:                    bodyB,
		LocalAnchorA:             localAnchorA,
		LocalAnchorB:             localAnchorB,
		Stiffness:                defaultDistanceJointStiffness,
		BiasFactor:               defaultDistanceJointBiasFactor,
		PositionCorrectionFactor: defaultDistanceJointPositionCorrectionFactor,
		Slop:                     defaultDistanceJointSlop,
		MaxCorrection:            defaultDistanceJointMaxCorrection,
		lastAxis:                 matrix.Vec3Right(),
	}
	joint.RestLength = joint.CurrentLength()
	return joint
}

func NewDistanceJointAtWorldAnchors(bodyA, bodyB *RigidBody, worldAnchorA, worldAnchorB matrix.Vec3) *DistanceJoint {
	return NewDistanceJoint(
		bodyA,
		bodyB,
		LocalAnchor(bodyA, worldAnchorA),
		LocalAnchor(bodyB, worldAnchorB),
	)
}

func NewDistanceJointToWorld(body *RigidBody, localAnchor, worldAnchor matrix.Vec3) *DistanceJoint {
	return NewDistanceJoint(body, nil, localAnchor, worldAnchor)
}

func (j *DistanceJoint) WorldAnchorA() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return WorldAnchor(j.BodyA, j.LocalAnchorA)
}

func (j *DistanceJoint) WorldAnchorB() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return WorldAnchor(j.BodyB, j.LocalAnchorB)
}

func (j *DistanceJoint) CurrentLength() matrix.Float {
	if j == nil {
		return 0
	}
	return j.WorldAnchorB().Subtract(j.WorldAnchorA()).Length()
}

func (j *DistanceJoint) SetWorldAnchors(worldAnchorA, worldAnchorB matrix.Vec3) {
	if j == nil {
		return
	}
	j.LocalAnchorA = LocalAnchor(j.BodyA, worldAnchorA)
	j.LocalAnchorB = LocalAnchor(j.BodyB, worldAnchorB)
	j.RestLength = j.CurrentLength()
	j.AccumulatedImpulse = 0
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *DistanceJoint) SetRestLength(restLength matrix.Float) {
	if j == nil {
		return
	}
	j.RestLength = matrix.Max(restLength, 0)
	j.AccumulatedImpulse = 0
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *DistanceJoint) Constraint() *Constraint {
	if j == nil {
		return nil
	}
	return j.constraint
}

func (j *DistanceJoint) IsStretched() bool {
	return j != nil && matrix.Abs(j.CurrentLength()-j.RestLength) > j.slop()
}

func (j *DistanceJoint) prepare(deltaTime matrix.Float) {
	if j == nil {
		return
	}
	anchorA := j.WorldAnchorA()
	anchorB := j.WorldAnchorB()
	axis, length, ok := j.axisAndLength(anchorA, anchorB)
	if !ok && j.RestLength <= j.slop() {
		j.row = ConstraintSolverRow{}
		j.AccumulatedImpulse = 0
		return
	}
	j.row.SetWorldAnchors(j.BodyA, j.BodyB, anchorA, anchorB, axis)
	j.row.EffectiveMass *= j.stiffness()
	j.row.Bias = j.bias(length, deltaTime)
	j.row.AccumulatedImpulse = 0
	if j.WarmStarting {
		j.row.AccumulatedImpulse = j.AccumulatedImpulse
		j.row.ApplyImpulse(j.AccumulatedImpulse)
	}
}

func (j *DistanceJoint) solveVelocity() {
	if j == nil {
		return
	}
	j.row.Solve()
	j.AccumulatedImpulse = j.row.AccumulatedImpulse
}

func (j *DistanceJoint) solvePosition() {
	if j == nil {
		return
	}
	anchorA := j.WorldAnchorA()
	anchorB := j.WorldAnchorB()
	axis, length, ok := j.axisAndLength(anchorA, anchorB)
	if !ok && j.RestLength <= j.slop() {
		return
	}
	error := length - j.RestLength
	if matrix.Abs(error) <= j.slop() {
		return
	}
	invMassA := j.BodyA.inverseMass()
	invMassB := j.BodyB.inverseMass()
	invMassSum := invMassA + invMassB
	if invMassSum <= contactEpsilon {
		return
	}
	correction := error * j.positionCorrectionFactor() * j.stiffness()
	correction = matrix.Clamp(correction, -j.maxCorrection(), j.maxCorrection())
	correction /= invMassSum
	moveBody(j.BodyA, axis.Scale(correction*invMassA))
	moveBody(j.BodyB, axis.Scale(-correction*invMassB))
}

func (j *DistanceJoint) axisAndLength(anchorA, anchorB matrix.Vec3) (matrix.Vec3, matrix.Float, bool) {
	delta := anchorB.Subtract(anchorA)
	lengthSq := delta.LengthSquared()
	if lengthSq <= distanceJointMinLength*distanceJointMinLength {
		axis := j.lastAxis
		if axis.LengthSquared() <= distanceJointMinLength*distanceJointMinLength {
			axis = matrix.Vec3Right()
		}
		j.lastAxis = axis.Normal()
		return j.lastAxis, 0, false
	}
	length := matrix.Sqrt(lengthSq)
	j.lastAxis = delta.Scale(1.0 / length)
	return j.lastAxis, length, true
}

func (j *DistanceJoint) bias(length, deltaTime matrix.Float) matrix.Float {
	if deltaTime <= 0 {
		deltaTime = defaultDistanceJointTimeStep
	}
	return (length - j.RestLength) * j.biasFactor() / deltaTime
}

func (j *DistanceJoint) stiffness() matrix.Float {
	if j.Stiffness < 0 {
		return 0
	}
	return matrix.Clamp(j.Stiffness, 0, 1)
}

func (j *DistanceJoint) biasFactor() matrix.Float {
	if j.BiasFactor < 0 {
		return 0
	}
	return j.BiasFactor
}

func (j *DistanceJoint) positionCorrectionFactor() matrix.Float {
	if j.PositionCorrectionFactor < 0 {
		return 0
	}
	return j.PositionCorrectionFactor
}

func (j *DistanceJoint) slop() matrix.Float {
	if j.Slop <= 0 {
		return defaultDistanceJointSlop
	}
	return j.Slop
}

func (j *DistanceJoint) maxCorrection() matrix.Float {
	if j.MaxCorrection <= 0 {
		return defaultDistanceJointMaxCorrection
	}
	return j.MaxCorrection
}

func LocalAnchor(body *RigidBody, worldAnchor matrix.Vec3) matrix.Vec3 {
	if body == nil {
		return worldAnchor
	}
	return body.Transform.InverseWorldMatrix().TransformPoint(worldAnchor)
}
