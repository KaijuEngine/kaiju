# Physics Constraint Entity Data

Kaiju's physics joints are authored as entity data on the source body. The
source entity must also have `RigidBodyEntityData`. `ConnectedEntityId` points to
the other body; leave it empty to anchor the source body to a world-space point.

All distances, anchors, lengths, corrections, forces, and torques are in engine
world units. Angles exposed on entity data use degrees. Runtime graviton APIs use
radians where their names do not explicitly say degrees.

## Shared Fields

- `ConnectedEntityId`: target entity id. Empty means body-to-world constraint.
- `LocalAnchorA`: source-body local anchor, in source local units.
- `TargetAnchorB`: target-body local anchor when `ConnectedEntityId` is set, or
  a world-space anchor when it is empty.
- `Stiffness`: solver stiffness multiplier. `1` is the usual rigid setting.
- `Bias`: positional bias factor applied during solving.
- `Correction`: positional correction factor.
- `Slop`: allowed positional error in world units before correction is applied.
- `MaxCorrection`: largest positional correction per solve step in world units.
- `WarmStarting`: reuse previous impulses for faster convergence.
- `Enabled`: whether the constraint starts active.
- `BreakForce`: linear break threshold. `0` disables force breaking.
- `BreakTorque`: angular break threshold. `0` disables torque breaking.

## Joint-Specific Fields

- `DistanceJointEntityData.RestLength`: target distance in world units.
- `DistanceJointEntityData.AutoRestLength`: derive `RestLength` from the initial
  anchor separation.
- `RopeJointEntityData.MaxLength`: maximum allowed distance in world units.
- `RopeJointEntityData.AutoMaxLength`: derive `MaxLength` from the initial anchor
  separation.
- `PointJointEntityData`: keeps the two anchors coincident.
- `HingeJointEntityData.HingeAxis`: local hinge axis on the source body.
- `HingeJointEntityData.EnableLimits`: clamps hinge angle when true.
- `HingeJointEntityData.MinAngleDegrees` / `MaxAngleDegrees`: angular limits in
  degrees.
- `HingeJointEntityData.EnableMotor`: enables motorized hinge motion.
- `HingeJointEntityData.MotorSpeedDegrees`: motor target speed in degrees per
  second.
- `HingeJointEntityData.MaxMotorTorque`: torque cap. When this is positive it is
  used instead of `MaxMotorImpulse`.
- `HingeJointEntityData.MaxMotorImpulse`: per-step motor impulse cap.

## Editor Helpers

The Create menu includes:

- `Connect selected as distance chain`
- `Connect selected as rope`
- `Connect selected as hinge chain`

The helpers use deterministic hierarchy order for the current selection and add
one constraint data binding for each adjacent pair. Distance and rope helpers
connect body centers and derive rest/max length from the current layout. The
hinge helper places each hinge at the midpoint between adjacent bodies and uses
the local backward axis so chains laid out along X can swing in the XY plane.

Sample stages are available from
`kaijuengine.com/engine_entity_data/engine_entity_data_physics/samples` via
`samples.ConstraintSampleStages()`: chain, rope, bridge, hinge pendulum, and
body-world anchor.
