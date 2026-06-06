/******************************************************************************/
/* hinge_joint.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

// HingeJoint keeps two anchors coincident and aligns each body's hinge axis.
// Relative rotation is constrained around the two axes perpendicular to the
// hinge, leaving rotation around the hinge axis free.
type HingeJoint struct {
	BodyA                     *RigidBody
	BodyB                     *RigidBody
	LocalAnchorA              matrix.Vec3
	LocalAnchorB              matrix.Vec3
	LocalAxisA                matrix.Vec3
	LocalAxisB                matrix.Vec3
	LocalRefA                 matrix.Vec3
	LocalRefB                 matrix.Vec3
	Stiffness                 matrix.Float
	BiasFactor                matrix.Float
	PositionCorrectionFactor  matrix.Float
	Slop                      matrix.Float
	MaxCorrection             matrix.Float
	WarmStarting              bool
	EnableLimits              bool
	MinAngle                  matrix.Float
	MaxAngle                  matrix.Float
	EnableMotor               bool
	MotorTargetSpeed          matrix.Float
	MaxMotorImpulse           matrix.Float
	MaxMotorTorque            matrix.Float
	AccumulatedAnchorImpulse  matrix.Vec3
	AccumulatedAngularImpulse matrix.Vec2
	AccumulatedLimitImpulse   matrix.Float
	AccumulatedMotorImpulse   matrix.Float
	constraint                *Constraint
	anchorRows                [3]ConstraintSolverRow
	angularRows               [2]AngularConstraintSolverRow
	limitRow                  AngularConstraintSolverRow
	motorRow                  AngularConstraintSolverRow
	limitState                int
}

func NewHingeJoint(bodyA, bodyB *RigidBody, localAnchorA, localAnchorB, localAxisA, localAxisB matrix.Vec3) *HingeJoint {
	joint := &HingeJoint{
		BodyA:                    bodyA,
		BodyB:                    bodyB,
		LocalAnchorA:             localAnchorA,
		LocalAnchorB:             localAnchorB,
		LocalAxisA:               safeNormal(localAxisA, matrix.Vec3Right()),
		LocalAxisB:               safeNormal(localAxisB, matrix.Vec3Right()),
		Stiffness:                defaultDistanceJointStiffness,
		BiasFactor:               defaultDistanceJointBiasFactor,
		PositionCorrectionFactor: defaultDistanceJointPositionCorrectionFactor,
		Slop:                     defaultDistanceJointSlop,
		MaxCorrection:            defaultDistanceJointMaxCorrection,
	}
	joint.setReferenceAxesFromCurrentPose()
	return joint
}

func NewHingeJointAtWorldAnchor(bodyA, bodyB *RigidBody, worldAnchor, worldAxis matrix.Vec3) *HingeJoint {
	axis := safeNormal(worldAxis, matrix.Vec3Right())
	return NewHingeJoint(
		bodyA,
		bodyB,
		LocalAnchor(bodyA, worldAnchor),
		LocalAnchor(bodyB, worldAnchor),
		LocalAxis(bodyA, axis),
		LocalAxis(bodyB, axis),
	)
}

func NewHingeJointToWorld(body *RigidBody, localAnchor, worldAnchor, localAxis, worldAxis matrix.Vec3) *HingeJoint {
	return NewHingeJoint(
		body,
		nil,
		localAnchor,
		worldAnchor,
		localAxis,
		safeNormal(worldAxis, matrix.Vec3Right()),
	)
}

func (j *HingeJoint) WorldAnchorA() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return WorldAnchor(j.BodyA, j.LocalAnchorA)
}

func (j *HingeJoint) WorldAnchorB() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return WorldAnchor(j.BodyB, j.LocalAnchorB)
}

func (j *HingeJoint) WorldAxisA() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Right()
	}
	return WorldAxis(j.BodyA, j.LocalAxisA)
}

func (j *HingeJoint) WorldAxisB() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Right()
	}
	return WorldAxis(j.BodyB, j.LocalAxisB)
}

func (j *HingeJoint) WorldReferenceA() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Up()
	}
	return WorldAxis(j.BodyA, j.LocalRefA)
}

func (j *HingeJoint) WorldReferenceB() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Up()
	}
	return WorldAxis(j.BodyB, j.LocalRefB)
}

func (j *HingeJoint) CurrentAngle() matrix.Float {
	if j == nil {
		return 0
	}
	axis := j.hingeAxis()
	refA := projectOnHingePlane(j.WorldReferenceA(), axis)
	refB := projectOnHingePlane(j.WorldReferenceB(), axis)
	return refB.SignedAngle(refA, axis)
}

func (j *HingeJoint) CurrentAngularVelocity() matrix.Float {
	if j == nil {
		return 0
	}
	axis := j.hingeAxis()
	return constraintBodyAngularVelocity(j.BodyA, axis.Negative()) +
		constraintBodyAngularVelocity(j.BodyB, axis)
}

func (j *HingeJoint) CurrentAnchorError() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return j.WorldAnchorB().Subtract(j.WorldAnchorA())
}

func (j *HingeJoint) CurrentAngularError() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return hingeAngularError(j.WorldAxisA(), j.WorldAxisB())
}

func (j *HingeJoint) SetWorldAnchors(worldAnchorA, worldAnchorB matrix.Vec3) {
	if j == nil {
		return
	}
	j.LocalAnchorA = LocalAnchor(j.BodyA, worldAnchorA)
	j.LocalAnchorB = LocalAnchor(j.BodyB, worldAnchorB)
	j.AccumulatedAnchorImpulse = matrix.Vec3Zero()
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *HingeJoint) SetWorldAxis(worldAxis matrix.Vec3) {
	if j == nil {
		return
	}
	axis := safeNormal(worldAxis, matrix.Vec3Right())
	j.LocalAxisA = LocalAxis(j.BodyA, axis)
	j.LocalAxisB = LocalAxis(j.BodyB, axis)
	j.setReferenceAxesFromCurrentPose()
	j.AccumulatedAngularImpulse = matrix.Vec2{}
	j.AccumulatedLimitImpulse = 0
	j.AccumulatedMotorImpulse = 0
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *HingeJoint) Constraint() *Constraint {
	if j == nil {
		return nil
	}
	return j.constraint
}

func (j *HingeJoint) SetAngularLimits(minAngle, maxAngle matrix.Float) {
	if j == nil {
		return
	}
	if minAngle > maxAngle {
		minAngle, maxAngle = maxAngle, minAngle
	}
	j.MinAngle = minAngle
	j.MaxAngle = maxAngle
	j.EnableLimits = true
	j.AccumulatedLimitImpulse = 0
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *HingeJoint) DisableAngularLimits() {
	if j == nil {
		return
	}
	j.EnableLimits = false
	j.AccumulatedLimitImpulse = 0
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *HingeJoint) SetMotor(targetSpeed, maxMotorImpulse matrix.Float) {
	if j == nil {
		return
	}
	j.MotorTargetSpeed = targetSpeed
	j.MaxMotorImpulse = matrix.Max(maxMotorImpulse, 0)
	j.EnableMotor = j.MaxMotorImpulse > 0 || j.MaxMotorTorque > 0
	j.AccumulatedMotorImpulse = 0
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *HingeJoint) SetMotorTorque(targetSpeed, maxMotorTorque matrix.Float) {
	if j == nil {
		return
	}
	j.MotorTargetSpeed = targetSpeed
	j.MaxMotorTorque = matrix.Max(maxMotorTorque, 0)
	j.EnableMotor = j.MaxMotorImpulse > 0 || j.MaxMotorTorque > 0
	j.AccumulatedMotorImpulse = 0
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *HingeJoint) DisableMotor() {
	if j == nil {
		return
	}
	j.EnableMotor = false
	j.AccumulatedMotorImpulse = 0
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *HingeJoint) IsStretched() bool {
	if j == nil {
		return false
	}
	if j.CurrentAnchorError().Length() > j.slop() ||
		j.CurrentAngularError().Length() > j.slop() {
		return true
	}
	if !j.EnableLimits {
		return false
	}
	angle := j.CurrentAngle()
	return angle < j.MinAngle-j.slop() || angle > j.MaxAngle+j.slop()
}

func (j *HingeJoint) prepare(deltaTime matrix.Float) {
	if j == nil {
		return
	}
	j.prepareAnchorRows(deltaTime)
	j.prepareAngularRows(deltaTime)
	j.prepareLimitRow(deltaTime)
	j.prepareMotorRow(deltaTime)
}

func (j *HingeJoint) prepareAnchorRows(deltaTime matrix.Float) {
	anchorA := j.WorldAnchorA()
	anchorB := j.WorldAnchorB()
	error := anchorB.Subtract(anchorA)
	for i, axis := range pointJointAxes {
		row := &j.anchorRows[i]
		row.SetWorldAnchors(j.BodyA, j.BodyB, anchorA, anchorB, axis)
		row.EffectiveMass *= j.stiffness()
		row.Bias = j.bias(error.Dot(axis), deltaTime)
		row.AccumulatedImpulse = 0
		if j.WarmStarting {
			row.AccumulatedImpulse = j.AccumulatedAnchorImpulse[i]
			row.ApplyImpulse(row.AccumulatedImpulse)
		}
	}
}

func (j *HingeJoint) prepareAngularRows(deltaTime matrix.Float) {
	axisA := j.WorldAxisA()
	axisB := j.WorldAxisB()
	hingeAxis := safeNormal(axisA.Add(axisB), axisA)
	error := hingeAngularError(axisA, axisB)
	axes := hingeConstraintAxes(hingeAxis)
	for i, axis := range axes {
		row := &j.angularRows[i]
		row.SetWorldAxis(j.BodyA, j.BodyB, axis)
		row.EffectiveMass *= j.stiffness()
		row.Bias = j.bias(error.Dot(axis), deltaTime)
		row.AccumulatedImpulse = 0
		if j.WarmStarting {
			row.AccumulatedImpulse = j.AccumulatedAngularImpulse[i]
			row.ApplyImpulse(row.AccumulatedImpulse)
		}
	}
}

func (j *HingeJoint) prepareLimitRow(deltaTime matrix.Float) {
	j.limitState = 0
	j.limitRow = AngularConstraintSolverRow{}
	if !j.EnableLimits {
		j.AccumulatedLimitImpulse = 0
		return
	}
	angle := j.CurrentAngle()
	axis := j.hingeAxis()
	row := &j.limitRow
	row.SetWorldAxis(j.BodyA, j.BodyB, axis)
	row.EffectiveMass *= j.stiffness()
	if angle < j.MinAngle {
		j.limitState = -1
		row.Bias = (j.MinAngle - angle) * j.biasFactor() / j.deltaTime(deltaTime)
		row.SetImpulseLimits(-matrix.Inf(1), 0)
	} else if angle > j.MaxAngle {
		j.limitState = 1
		row.Bias = -(angle - j.MaxAngle) * j.biasFactor() / j.deltaTime(deltaTime)
		row.SetImpulseLimits(0, matrix.Inf(1))
	} else {
		j.AccumulatedLimitImpulse = 0
		return
	}
	row.AccumulatedImpulse = 0
	if j.WarmStarting {
		row.AccumulatedImpulse = j.clampedLimitWarmImpulse()
		row.ApplyImpulse(row.AccumulatedImpulse)
	}
}

func (j *HingeJoint) prepareMotorRow(deltaTime matrix.Float) {
	j.motorRow = AngularConstraintSolverRow{}
	if !j.EnableMotor {
		j.AccumulatedMotorImpulse = 0
		return
	}
	maxImpulse := j.motorImpulseLimit(deltaTime)
	if maxImpulse <= 0 {
		j.AccumulatedMotorImpulse = 0
		return
	}
	row := &j.motorRow
	row.SetWorldAxis(j.BodyA, j.BodyB, j.hingeAxis())
	row.Bias = -j.MotorTargetSpeed
	row.SetImpulseLimits(-maxImpulse, maxImpulse)
	row.AccumulatedImpulse = 0
	if j.WarmStarting {
		row.AccumulatedImpulse = matrix.Clamp(j.AccumulatedMotorImpulse, -maxImpulse, maxImpulse)
		row.ApplyImpulse(row.AccumulatedImpulse)
	}
}

func (j *HingeJoint) solveVelocity() {
	if j == nil {
		return
	}
	for i := range j.anchorRows {
		j.anchorRows[i].Solve()
		j.AccumulatedAnchorImpulse[i] = j.anchorRows[i].AccumulatedImpulse
	}
	for i := range j.angularRows {
		j.angularRows[i].Solve()
		j.AccumulatedAngularImpulse[i] = j.angularRows[i].AccumulatedImpulse
	}
	if j.limitState != 0 {
		j.limitRow.Solve()
		j.AccumulatedLimitImpulse = j.limitRow.AccumulatedImpulse
	}
	if j.EnableMotor {
		j.motorRow.Solve()
		j.AccumulatedMotorImpulse = j.motorRow.AccumulatedImpulse
	}
}

func (j *HingeJoint) solvePosition() {
	if j == nil {
		return
	}
	j.solveAnchorPosition()
	j.solveAngularPosition()
	j.solveLimitPosition()
}

func (j *HingeJoint) solveAnchorPosition() {
	error := j.CurrentAnchorError()
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

func (j *HingeJoint) solveAngularPosition() {
	error := j.CurrentAngularError()
	if error.Length() <= j.slop() {
		return
	}
	axis := safeNormal(error, matrix.Vec3Right())
	invA := AngularAxisEffectiveMass(j.BodyA, axis)
	invB := AngularAxisEffectiveMass(j.BodyB, axis)
	invSum := invA + invB
	if invSum <= contactEpsilon {
		return
	}
	correction := j.clampedAngularCorrection(error)
	rotateBody(j.BodyA, correction.Scale(invA/invSum))
	rotateBody(j.BodyB, correction.Scale(-invB/invSum))
}

func (j *HingeJoint) solveLimitPosition() {
	if !j.EnableLimits {
		return
	}
	angle := j.CurrentAngle()
	var correction matrix.Float
	if angle < j.MinAngle {
		correction = j.MinAngle - angle
	} else if angle > j.MaxAngle {
		correction = -(angle - j.MaxAngle)
	} else {
		return
	}
	if matrix.Abs(correction) <= j.slop() {
		return
	}
	axis := j.hingeAxis()
	invA := AngularAxisEffectiveMass(j.BodyA, axis)
	invB := AngularAxisEffectiveMass(j.BodyB, axis)
	invSum := invA + invB
	if invSum <= contactEpsilon {
		return
	}
	correction = matrix.Clamp(correction*j.positionCorrectionFactor()*j.stiffness(), -j.maxCorrection(), j.maxCorrection())
	correctionVector := axis.Scale(correction)
	rotateBody(j.BodyA, correctionVector.Scale(invA/invSum))
	rotateBody(j.BodyB, correctionVector.Scale(-invB/invSum))
}

func (j *HingeJoint) bias(error, deltaTime matrix.Float) matrix.Float {
	deltaTime = j.deltaTime(deltaTime)
	if matrix.Abs(error) <= j.slop() {
		return 0
	}
	return error * j.biasFactor() / deltaTime
}

func (j *HingeJoint) clampedCorrection(error matrix.Vec3) matrix.Vec3 {
	correction := error.Scale(j.positionCorrectionFactor() * j.stiffness())
	maxCorrection := j.maxCorrection()
	length := correction.Length()
	if length > maxCorrection && length > matrix.FloatSmallestNonzero {
		correction = correction.Scale(maxCorrection / length)
	}
	return correction
}

func (j *HingeJoint) clampedAngularCorrection(error matrix.Vec3) matrix.Vec3 {
	correction := error.Scale(j.positionCorrectionFactor() * j.stiffness())
	maxCorrection := j.maxCorrection()
	length := correction.Length()
	if length > maxCorrection && length > matrix.FloatSmallestNonzero {
		correction = correction.Scale(maxCorrection / length)
	}
	return correction
}

func (j *HingeJoint) stiffness() matrix.Float {
	if j.Stiffness < 0 {
		return 0
	}
	return matrix.Clamp(j.Stiffness, 0, 1)
}

func (j *HingeJoint) biasFactor() matrix.Float {
	if j.BiasFactor < 0 {
		return 0
	}
	return j.BiasFactor
}

func (j *HingeJoint) positionCorrectionFactor() matrix.Float {
	if j.PositionCorrectionFactor < 0 {
		return 0
	}
	return j.PositionCorrectionFactor
}

func (j *HingeJoint) slop() matrix.Float {
	if j.Slop <= 0 {
		return defaultDistanceJointSlop
	}
	return j.Slop
}

func (j *HingeJoint) maxCorrection() matrix.Float {
	if j.MaxCorrection <= 0 {
		return defaultDistanceJointMaxCorrection
	}
	return j.MaxCorrection
}

func (j *HingeJoint) deltaTime(deltaTime matrix.Float) matrix.Float {
	if deltaTime <= 0 {
		return defaultDistanceJointTimeStep
	}
	return deltaTime
}

func (j *HingeJoint) hingeAxis() matrix.Vec3 {
	axisA := j.WorldAxisA()
	axisB := j.WorldAxisB()
	return safeNormal(axisA.Add(axisB), axisA)
}

func (j *HingeJoint) setReferenceAxesFromCurrentPose() {
	axis := j.hingeAxis()
	reference := safeNormal(axis.Orthogonal(), matrix.Vec3Up())
	j.LocalRefA = LocalAxis(j.BodyA, reference)
	j.LocalRefB = LocalAxis(j.BodyB, reference)
}

func (j *HingeJoint) clampedLimitWarmImpulse() matrix.Float {
	if j.limitState < 0 {
		return matrix.Min(j.AccumulatedLimitImpulse, 0)
	}
	if j.limitState > 0 {
		return matrix.Max(j.AccumulatedLimitImpulse, 0)
	}
	return 0
}

func (j *HingeJoint) motorImpulseLimit(deltaTime matrix.Float) matrix.Float {
	if j.MaxMotorImpulse > 0 {
		return j.MaxMotorImpulse
	}
	if j.MaxMotorTorque > 0 {
		return j.MaxMotorTorque * j.deltaTime(deltaTime)
	}
	return 0
}

func (j *HingeJoint) AccumulatedAngularImpulseMagnitude() matrix.Float {
	if j == nil {
		return 0
	}
	sum := j.AccumulatedAngularImpulse.X()*j.AccumulatedAngularImpulse.X() +
		j.AccumulatedAngularImpulse.Y()*j.AccumulatedAngularImpulse.Y() +
		j.AccumulatedLimitImpulse*j.AccumulatedLimitImpulse +
		j.AccumulatedMotorImpulse*j.AccumulatedMotorImpulse
	return matrix.Sqrt(sum)
}

func LocalAxis(body *RigidBody, worldAxis matrix.Vec3) matrix.Vec3 {
	axis := safeNormal(worldAxis, matrix.Vec3Right())
	if body == nil {
		return axis
	}
	rotation := body.Rotation()
	rotation.Inverse()
	return safeNormal(rotation.Rotate(axis), matrix.Vec3Right())
}

func WorldAxis(body *RigidBody, localAxis matrix.Vec3) matrix.Vec3 {
	axis := safeNormal(localAxis, matrix.Vec3Right())
	if body == nil {
		return axis
	}
	return safeNormal(body.Rotation().Rotate(axis), matrix.Vec3Right())
}

func hingeConstraintAxes(hingeAxis matrix.Vec3) [2]matrix.Vec3 {
	axis := safeNormal(hingeAxis, matrix.Vec3Right())
	first := safeNormal(axis.Orthogonal(), matrix.Vec3Up())
	second := safeNormal(axis.Cross(first), matrix.Vec3Forward())
	return [2]matrix.Vec3{first, second}
}

func hingeAngularError(axisA, axisB matrix.Vec3) matrix.Vec3 {
	a := safeNormal(axisA, matrix.Vec3Right())
	b := safeNormal(axisB, matrix.Vec3Right())
	cross := a.Cross(b)
	sin := cross.Length()
	dot := matrix.Clamp(a.Dot(b), -1, 1)
	if sin <= contactEpsilon {
		if dot >= 0 {
			return matrix.Vec3Zero()
		}
		return safeNormal(a.Orthogonal(), matrix.Vec3Up()).Scale(matrix.Atan2(0, dot))
	}
	return cross.Scale(matrix.Atan2(sin, dot) / sin)
}

func projectOnHingePlane(vector, axis matrix.Vec3) matrix.Vec3 {
	normal := safeNormal(axis, matrix.Vec3Right())
	projected := vector.Subtract(normal.Scale(vector.Dot(normal)))
	return safeNormal(projected, normal.Orthogonal())
}

func rotateBody(body *RigidBody, angularCorrection matrix.Vec3) {
	if body == nil || body.inverseInertia().IsZero() {
		return
	}
	angle := angularCorrection.Length()
	if angle <= contactEpsilon {
		return
	}
	delta := matrix.QuaternionAxisAngle(angularCorrection.Scale(1.0/angle), angle)
	current := matrix.QuaternionFromEuler(body.Transform.Rotation())
	next := delta.Multiply(current)
	next.Normalize()
	body.Transform.SetRotation(next.ToEuler())
}
