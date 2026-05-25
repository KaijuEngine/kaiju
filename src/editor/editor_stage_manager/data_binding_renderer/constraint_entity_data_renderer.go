/******************************************************************************/
/* constraint_entity_data_renderer.go                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package data_binding_renderer

import (
	"errors"
	"fmt"
	"log/slog"
	"math"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	constraintAnchorRadius = matrix.Float(0.08)
	constraintAxisLength   = matrix.Float(0.8)
	constraintArcRadius    = matrix.Float(0.45)
	constraintArcSegments  = 24
)

type constraintGizmoKind int

const (
	constraintGizmoDistance constraintGizmoKind = iota
	constraintGizmoRope
	constraintGizmoPoint
	constraintGizmoHinge
)

var constraintDataKeys = map[string]constraintGizmoKind{
	pod.QualifiedNameForLayout(engine_entity_data_physics.DistanceJointEntityData{}): constraintGizmoDistance,
	pod.QualifiedNameForLayout(engine_entity_data_physics.RopeJointEntityData{}):     constraintGizmoRope,
	pod.QualifiedNameForLayout(engine_entity_data_physics.PointJointEntityData{}):    constraintGizmoPoint,
	pod.QualifiedNameForLayout(engine_entity_data_physics.HingeJointEntityData{}):    constraintGizmoHinge,
}

type constraintGizmoData struct {
	kind              constraintGizmoKind
	connectedEntityId engine.EntityId
	localAnchorA      matrix.Vec3
	targetAnchorB     matrix.Vec3
	hingeAxis         matrix.Vec3
	enableLimits      bool
	minAngleDegrees   matrix.Float
	maxAngleDegrees   matrix.Float
}

type constraintGizmo struct {
	data             constraintGizmoData
	owner            *editor_stage_manager.StageEntity
	target           *editor_stage_manager.StageEntity
	manager          *editor_stage_manager.StageManager
	anchorATransform matrix.Transform
	anchorBTransform matrix.Transform
	anchorA          rendering.DrawInstance
	anchorB          rendering.DrawInstance
	link             rendering.DrawInstance
	axis             rendering.DrawInstance
	arc              rendering.DrawInstance
	lineKey          string
	axisKey          string
	arcKey           string
	lastAnchorA      matrix.Vec3
	lastAnchorB      matrix.Vec3
	lastAxis         matrix.Vec3
	invalidTarget    bool
	visible          bool
	version          int
}

type ConstraintEntityDataRenderer struct {
	Gizmos   map[*entity_data_binding.EntityDataEntry]*constraintGizmo
	updateId engine.UpdateId
}

func init() {
	r := &ConstraintEntityDataRenderer{
		Gizmos: make(map[*entity_data_binding.EntityDataEntry]*constraintGizmo),
	}
	for key := range constraintDataKeys {
		AddRenderer(key, r)
	}
}

func (c *ConstraintEntityDataRenderer) Attached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("ConstraintEntityDataRenderer.Attached").End()
	if _, ok := c.Gizmos[data]; ok {
		c.Detatched(host, manager, target, data)
	}
	g := &constraintGizmo{
		owner:   target,
		manager: manager,
		data:    constraintGizmoDataFromEntry(data),
	}
	g.anchorATransform.SetupRawTransform()
	g.anchorBTransform.SetupRawTransform()
	if err := g.create(host, manager); err != nil {
		slog.Error("failed to create constraint gizmo", "error", err)
		return
	}
	g.deactivate()
	c.Gizmos[data] = g
	if !c.updateId.IsValid() {
		c.updateId = host.Updater.AddUpdate(func(float64) {
			c.update(host, manager)
		})
	}
	target.OnDestroy.Add(func() {
		c.Detatched(host, manager, target, data)
	})
}

func (c *ConstraintEntityDataRenderer) Detatched(host *engine.Host, _ *editor_stage_manager.StageManager, _ *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("ConstraintEntityDataRenderer.Detatched").End()
	if g, ok := c.Gizmos[data]; ok {
		g.destroy(host)
		delete(c.Gizmos, data)
	}
	if len(c.Gizmos) == 0 && c.updateId.IsValid() {
		host.Updater.RemoveUpdate(&c.updateId)
	}
}

func (c *ConstraintEntityDataRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("ConstraintEntityDataRenderer.Show").End()
	if g, ok := c.Gizmos[data]; ok {
		g.visible = true
		g.refresh(host, g.manager, false)
		g.activate()
	}
}

func (c *ConstraintEntityDataRenderer) Hide(_ *engine.Host, _ *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("ConstraintEntityDataRenderer.Hide").End()
	if g, ok := c.Gizmos[data]; ok {
		g.visible = false
		g.deactivate()
	}
}

func (c *ConstraintEntityDataRenderer) Update(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("ConstraintEntityDataRenderer.Update").End()
	if g, ok := c.Gizmos[data]; ok {
		g.data = constraintGizmoDataFromEntry(data)
		g.owner = target
		g.refresh(host, g.manager, true)
	}
}

func (c *ConstraintEntityDataRenderer) update(host *engine.Host, manager *editor_stage_manager.StageManager) {
	defer tracing.NewRegion("ConstraintEntityDataRenderer.update").End()
	for _, g := range c.Gizmos {
		if g.owner == nil || g.owner.IsDeleted() {
			continue
		}
		g.refresh(host, manager, false)
	}
}

func (g *constraintGizmo) create(host *engine.Host, manager *editor_stage_manager.StageManager) error {
	g.resolveTarget(manager)
	a, b, axis := g.currentWorldData()
	g.lastAnchorA = a
	g.lastAnchorB = b
	g.lastAxis = axis
	g.anchorATransform.SetWorldPosition(a)
	g.anchorBTransform.SetWorldPosition(b)
	color := g.color()
	var err error
	if g.anchorA, err = constraintDrawWireSphere(host, &g.anchorATransform, color); err != nil {
		return err
	}
	if g.anchorB, err = constraintDrawWireSphere(host, &g.anchorBTransform, color); err != nil {
		g.destroy(host)
		return err
	}
	return g.rebuildWorldDrawings(host, a, b, axis)
}

func (g *constraintGizmo) refresh(host *engine.Host, manager *editor_stage_manager.StageManager, force bool) {
	if manager != nil {
		g.resolveTarget(manager)
	}
	a, b, axis := g.currentWorldData()
	if !force && a.Equals(g.lastAnchorA) && b.Equals(g.lastAnchorB) && axis.Equals(g.lastAxis) {
		g.applyColor()
		return
	}
	g.lastAnchorA = a
	g.lastAnchorB = b
	g.lastAxis = axis
	g.anchorATransform.SetWorldPosition(a)
	g.anchorBTransform.SetWorldPosition(b)
	if err := g.rebuildWorldDrawings(host, a, b, axis); err != nil {
		slog.Error("failed to refresh constraint gizmo", "error", err)
	}
	if !g.visible {
		g.deactivate()
	}
}

func (g *constraintGizmo) resolveTarget(manager *editor_stage_manager.StageManager) {
	g.target = nil
	g.invalidTarget = false
	if g.data.connectedEntityId == "" || manager == nil {
		return
	}
	if target, ok := manager.EntityById(string(g.data.connectedEntityId)); ok && !target.IsDeleted() {
		g.target = target
		return
	}
	g.invalidTarget = true
}

func (g *constraintGizmo) currentWorldData() (matrix.Vec3, matrix.Vec3, matrix.Vec3) {
	anchorA := g.owner.Transform.WorldMatrix().TransformPoint(g.data.localAnchorA)
	anchorB := g.data.targetAnchorB
	if g.target != nil {
		anchorB = g.target.Transform.WorldMatrix().TransformPoint(g.data.targetAnchorB)
	}
	axis := g.data.hingeAxis
	if axis.LengthSquared() <= matrix.FloatSmallestNonzero {
		axis = matrix.Vec3Right()
	}
	axis = axis.Normal()
	ownerOrigin := g.owner.Transform.WorldMatrix().TransformPoint(matrix.Vec3Zero())
	axisEnd := g.owner.Transform.WorldMatrix().TransformPoint(axis)
	worldAxis := axisEnd.Subtract(ownerOrigin)
	if worldAxis.LengthSquared() <= matrix.FloatSmallestNonzero {
		worldAxis = matrix.Vec3Right()
	} else {
		worldAxis = worldAxis.Normal()
	}
	return anchorA, anchorB, worldAxis
}

func (g *constraintGizmo) rebuildWorldDrawings(host *engine.Host, a, b, axis matrix.Vec3) error {
	g.destroyWorldDrawings(host)
	g.version++
	color := g.color()
	link, key, err := constraintDrawLine(host, g.meshKey("link"), a, b, color)
	if err != nil {
		return err
	}
	g.link = link
	g.lineKey = key
	if g.data.kind != constraintGizmoHinge {
		g.applyVisibilityTo(g.link)
		g.applyColor()
		return nil
	}
	axisStart := a.Subtract(axis.Scale(constraintAxisLength))
	axisEnd := a.Add(axis.Scale(constraintAxisLength))
	g.axis, g.axisKey, err = constraintDrawLine(host, g.meshKey("axis"), axisStart, axisEnd, matrix.ColorDeepSkyBlue())
	if err != nil {
		return err
	}
	if g.data.enableLimits {
		g.arc, g.arcKey, err = constraintDrawArc(host, g.meshKey("arc"), a, axis, g.data.minAngleDegrees, g.data.maxAngleDegrees, matrix.ColorGold())
		if err != nil {
			return err
		}
	}
	g.applyVisibilityTo(g.link, g.axis, g.arc)
	g.applyColor()
	return nil
}

func (g *constraintGizmo) meshKey(kind string) string {
	return fmt.Sprintf("constraint_gizmo_%p_%s_%d", g, kind, g.version)
}

func (g *constraintGizmo) destroy(host *engine.Host) {
	g.destroyWorldDrawings(host)
	for _, d := range []rendering.DrawInstance{g.anchorA, g.anchorB} {
		if d != nil {
			d.Destroy()
		}
	}
	g.anchorA = nil
	g.anchorB = nil
}

func (g *constraintGizmo) destroyWorldDrawings(host *engine.Host) {
	for _, d := range []rendering.DrawInstance{g.link, g.axis, g.arc} {
		if d != nil {
			d.Destroy()
		}
	}
	for _, key := range []string{g.lineKey, g.axisKey, g.arcKey} {
		if key != "" {
			host.MeshCache().RemoveMesh(key)
		}
	}
	g.link = nil
	g.axis = nil
	g.arc = nil
	g.lineKey = ""
	g.axisKey = ""
	g.arcKey = ""
}

func (g *constraintGizmo) activate() {
	g.applyVisibilityTo(g.anchorA, g.anchorB, g.link, g.axis, g.arc)
}

func (g *constraintGizmo) deactivate() {
	for _, d := range []rendering.DrawInstance{g.anchorA, g.anchorB, g.link, g.axis, g.arc} {
		if d != nil {
			d.Deactivate()
		}
	}
}

func (g *constraintGizmo) applyVisibilityTo(draws ...rendering.DrawInstance) {
	for _, d := range draws {
		if d == nil {
			continue
		}
		if g.visible {
			d.Activate()
		} else {
			d.Deactivate()
		}
	}
}

func (g *constraintGizmo) applyColor() {
	color := g.color()
	for _, d := range []rendering.DrawInstance{g.anchorA, g.anchorB, g.link} {
		if sd, ok := d.(*shader_data_registry.ShaderDataEdTransformWire); ok {
			sd.Color = color
		}
	}
}

func (g *constraintGizmo) color() matrix.Color {
	if g.invalidTarget {
		return matrix.ColorOrange()
	}
	return matrix.ColorGreen()
}

func constraintDrawWireSphere(host *engine.Host, transform *matrix.Transform, color matrix.Color) (rendering.DrawInstance, error) {
	mesh := rendering.NewMeshWireSphere(host.MeshCache(), constraintAnchorRadius, 8, 8)
	return constraintAddWireDrawing(host, mesh, transform, color)
}

func constraintDrawLine(host *engine.Host, key string, a, b matrix.Vec3, color matrix.Color) (rendering.DrawInstance, string, error) {
	mesh := rendering.NewMeshLine(host.MeshCache(), key, a, b, matrix.ColorWhite())
	sd, err := constraintAddWireDrawing(host, mesh, nil, color)
	return sd, key, err
}

func constraintDrawArc(host *engine.Host, key string, center, axis matrix.Vec3, minDegrees, maxDegrees matrix.Float, color matrix.Color) (rendering.DrawInstance, string, error) {
	points := constraintArcPoints(center, axis, minDegrees, maxDegrees)
	mesh := rendering.NewMeshGrid(host.MeshCache(), key, points, matrix.ColorWhite())
	sd, err := constraintAddWireDrawing(host, mesh, nil, color)
	return sd, key, err
}

func constraintAddWireDrawing(host *engine.Host, mesh *rendering.Mesh, transform *matrix.Transform, color matrix.Color) (rendering.DrawInstance, error) {
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionEdTransformWire)
	if err != nil {
		slog.Error("failed to load constraint gizmo wire material", "error", err)
		return nil, errors.New("failed to load constraint gizmo wire material")
	}
	sd := shader_data_registry.Create("ed_transform_wire")
	gsd := sd.(*shader_data_registry.ShaderDataEdTransformWire)
	gsd.Color = color
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       mesh,
		ShaderData: gsd,
		Transform:  transform,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	})
	return gsd, nil
}

func constraintArcPoints(center, axis matrix.Vec3, minDegrees, maxDegrees matrix.Float) []matrix.Vec3 {
	if axis.LengthSquared() <= matrix.FloatSmallestNonzero {
		axis = matrix.Vec3Right()
	}
	axis = axis.Normal()
	ref := matrix.Vec3Up()
	if matrix.Abs(axis.Dot(ref)) > 0.95 {
		ref = matrix.Vec3Forward()
	}
	u := axis.Cross(ref).Normal()
	v := axis.Cross(u).Normal()
	minRad := float64(matrix.Deg2Rad(minDegrees))
	maxRad := float64(matrix.Deg2Rad(maxDegrees))
	if maxRad < minRad {
		maxRad = minRad
	}
	points := make([]matrix.Vec3, 0, constraintArcSegments*2)
	last := center.Add(u.Scale(constraintArcRadius * matrix.Float(math.Cos(minRad))).
		Add(v.Scale(constraintArcRadius * matrix.Float(math.Sin(minRad)))))
	for i := 1; i <= constraintArcSegments; i++ {
		t := float64(i) / float64(constraintArcSegments)
		ang := minRad + ((maxRad - minRad) * t)
		next := center.Add(u.Scale(constraintArcRadius * matrix.Float(math.Cos(ang))).
			Add(v.Scale(constraintArcRadius * matrix.Float(math.Sin(ang)))))
		points = append(points, last, next)
		last = next
	}
	return points
}

func constraintGizmoDataFromEntry(data *entity_data_binding.EntityDataEntry) constraintGizmoData {
	g := constraintGizmoData{
		kind:              constraintDataKeys[data.Gen.RegisterKey],
		connectedEntityId: data.FieldValueByName("ConnectedEntityId").(engine.EntityId),
		localAnchorA:      data.FieldValueByName("LocalAnchorA").(matrix.Vec3),
		targetAnchorB:     data.FieldValueByName("TargetAnchorB").(matrix.Vec3),
		hingeAxis:         matrix.Vec3Right(),
	}
	if g.kind == constraintGizmoHinge {
		g.hingeAxis = data.FieldValueByName("HingeAxis").(matrix.Vec3)
		g.enableLimits = data.FieldValueByName("EnableLimits").(bool)
		g.minAngleDegrees = data.FieldValueByName("MinAngleDegrees").(matrix.Float)
		g.maxAngleDegrees = data.FieldValueByName("MaxAngleDegrees").(matrix.Float)
	}
	return g
}
