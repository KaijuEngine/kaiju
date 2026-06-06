/******************************************************************************/
/* constraint.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/matrix"
)

type ConstraintType uint8

const (
	ConstraintTypeUnknown ConstraintType = iota
	ConstraintTypeGeneric
	ConstraintTypeDistance
	ConstraintTypeRope
	ConstraintTypePoint
	ConstraintTypeHinge
)

// Constraint stores the lifecycle and endpoints for a future Graviton
// constraint solver. BodyA and BodyB form a body-body constraint; either body
// may be nil to represent a body-world constraint.
type Constraint struct {
	Type     ConstraintType
	BodyA    *RigidBody
	BodyB    *RigidBody
	Rows     []ConstraintSolverRow
	Distance *DistanceJoint
	Rope     *RopeJoint
	Point    *PointJoint
	Hinge    *HingeJoint
	Active   bool
	Enabled  bool
	// BreakForce and BreakTorque are optional impulse thresholds. Values <= 0
	// leave that break mode disabled.
	BreakForce  matrix.Float
	BreakTorque matrix.Float
	Broken      bool
	poolId      pooling.PoolGroupId
	id          pooling.PoolIndex
	pooled      bool
	awake       bool
}

func (c *Constraint) IsBodyBody() bool {
	return c != nil && c.BodyA != nil && c.BodyB != nil
}

func (c *Constraint) IsBodyWorld() bool {
	return c != nil && ((c.BodyA != nil && c.BodyB == nil) ||
		(c.BodyA == nil && c.BodyB != nil))
}

func (c *Constraint) BodiesValid() bool {
	if c == nil {
		return false
	}
	if c.BodyA == nil && c.BodyB == nil {
		return false
	}
	if c.BodyA != nil && !constraintBodyValid(c.BodyA) {
		return false
	}
	if c.BodyB != nil && !constraintBodyValid(c.BodyB) {
		return false
	}
	return true
}

func (c *Constraint) IsValid() bool {
	return c != nil && c.pooled && c.Active && c.Enabled && c.BodiesValid()
}

func (c *Constraint) SetBodies(bodyA, bodyB *RigidBody) {
	c.BodyA = bodyA
	c.BodyB = bodyB
	if c.Distance != nil {
		c.Distance.BodyA = bodyA
		c.Distance.BodyB = bodyB
	}
	if c.Rope != nil {
		c.Rope.BodyA = bodyA
		c.Rope.BodyB = bodyB
	}
	if c.Point != nil {
		c.Point.BodyA = bodyA
		c.Point.BodyB = bodyB
	}
	if c.Hinge != nil {
		c.Hinge.BodyA = bodyA
		c.Hinge.BodyB = bodyB
	}
	c.disableIfBodiesInvalid()
	c.syncAwakeState()
}

func (c *Constraint) SetActive(active bool) {
	if c == nil {
		return
	}
	c.Active = active
	c.syncAwakeState()
}

func (c *Constraint) SetEnabled(enabled bool) {
	if c == nil {
		return
	}
	c.Enabled = enabled
	c.syncAwakeState()
}

func (c *Constraint) SetBreakForce(threshold matrix.Float) {
	if c == nil {
		return
	}
	c.BreakForce = threshold
}

func (c *Constraint) SetBreakTorque(threshold matrix.Float) {
	if c == nil {
		return
	}
	c.BreakTorque = threshold
}

func (c *Constraint) BreakIfNeeded() bool {
	if c == nil || c.Broken || (!c.hasBreakForce() && !c.hasBreakTorque()) {
		return false
	}
	if c.hasBreakForce() && c.AccumulatedLinearImpulse() > c.BreakForce {
		c.breakConstraint()
		return true
	}
	if c.hasBreakTorque() && c.AccumulatedAngularImpulse() > c.BreakTorque {
		c.breakConstraint()
		return true
	}
	return false
}

func (c *Constraint) AccumulatedLinearImpulse() matrix.Float {
	if c == nil {
		return 0
	}
	if c.Distance != nil {
		return matrix.Abs(c.Distance.AccumulatedImpulse)
	}
	if c.Rope != nil {
		return matrix.Abs(c.Rope.AccumulatedImpulse)
	}
	if c.Point != nil {
		return c.Point.AccumulatedImpulse.Length()
	}
	if c.Hinge != nil {
		return c.Hinge.AccumulatedAnchorImpulse.Length()
	}
	var sum matrix.Float
	for i := range c.Rows {
		sum += c.Rows[i].AccumulatedImpulse * c.Rows[i].AccumulatedImpulse
	}
	return matrix.Sqrt(sum)
}

func (c *Constraint) AccumulatedAngularImpulse() matrix.Float {
	if c == nil {
		return 0
	}
	if c.Hinge != nil {
		return c.Hinge.AccumulatedAngularImpulseMagnitude()
	}
	return 0
}

func (c *Constraint) breakConstraint() {
	c.Broken = true
	c.Enabled = false
	c.Active = false
	c.awake = false
}

func (c *Constraint) hasBreakForce() bool {
	return c.BreakForce > 0
}

func (c *Constraint) hasBreakTorque() bool {
	return c.BreakTorque > 0
}

func (c *Constraint) disableIfBodiesInvalid() {
	if !c.BodiesValid() {
		c.Active = false
		c.Enabled = false
	}
	c.syncAwakeState()
}

func (c *Constraint) syncAwakeState() {
	if c == nil {
		return
	}
	awake := c.Active && c.Enabled && c.BodiesValid()
	if awake && !c.awake {
		c.WakeBodies()
	}
	c.awake = awake
}

func (c *Constraint) WakeBodies() {
	if c == nil {
		return
	}
	WakeConstrainedBodies(c.BodyA, c.BodyB)
}

func (c *Constraint) IsStretched() bool {
	return c != nil &&
		((c.Distance != nil && c.Distance.IsStretched()) ||
			(c.Rope != nil && c.Rope.IsStretched()) ||
			(c.Point != nil && c.Point.IsStretched()) ||
			(c.Hinge != nil && c.Hinge.IsStretched()))
}

func (c *Constraint) detachBody(body *RigidBody) {
	if body == nil {
		return
	}
	if c.BodyA == body {
		c.BodyA = nil
	}
	if c.BodyB == body {
		c.BodyB = nil
	}
	if c.Distance != nil {
		c.Distance.BodyA = c.BodyA
		c.Distance.BodyB = c.BodyB
	}
	if c.Rope != nil {
		c.Rope.BodyA = c.BodyA
		c.Rope.BodyB = c.BodyB
	}
	if c.Point != nil {
		c.Point.BodyA = c.BodyA
		c.Point.BodyB = c.BodyB
	}
	if c.Hinge != nil {
		c.Hinge.BodyA = c.BodyA
		c.Hinge.BodyB = c.BodyB
	}
	c.Active = false
	c.Enabled = false
	c.awake = false
}

func constraintBodyValid(body *RigidBody) bool {
	return body != nil && body.pooled
}
