package data_binding_renderer

import (
	"errors"
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_stage_manager"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine_entity_data/engine_entity_data_physics"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
)

type RigidBodyGizmo struct {
	ShaderData rendering.DrawInstance
	Extent     matrix.Vec3
	Mass       float32
	Radius     float32
	Height     float32
	IsStatic   bool
	Shape      engine_entity_data_physics.Shape
}

type RigidBodyEntityDataRenderer struct {
	Wireframes map[*editor_stage_manager.StageEntity]RigidBodyGizmo
}

func init() {
	AddRenderer(engine_entity_data_physics.BindingKey, &RigidBodyEntityDataRenderer{
		Wireframes: make(map[*editor_stage_manager.StageEntity]RigidBodyGizmo),
	})
}

func (c *RigidBodyEntityDataRenderer) Attached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	// commonAttached(host, manager, target, "light.png")
}

func (c *RigidBodyEntityDataRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("RigidBodyEntityDataRenderer.Show").End()
	if _, ok := c.Wireframes[target]; ok {
		slog.Error("there is an internal error in state for the editor's RigidBodyEntityDataRenderer, show was called before any hide happened. Double selected the same target?")
		c.Hide(host, target, data)
	}

	g := RigidBodyGizmo{}
	g.reloadData(data)
	var err error
	if g.ShaderData, err = rigidBodyLoadWireframe(host, g, &target.Transform); err == nil {
		c.Wireframes[target] = g
	}
}

func (c *RigidBodyEntityDataRenderer) Update(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	if g, ok := c.Wireframes[target]; ok {
		if g.reloadData(data) {
			g.ShaderData.Destroy()
			var err error
			if g.ShaderData, err = rigidBodyLoadWireframe(host, g, &target.Transform); err == nil {
				c.Wireframes[target] = g
			}
		}
	}
}

func rigidBodyLoadWireframe(host *engine.Host, g RigidBodyGizmo, transform *matrix.Transform) (rendering.DrawInstance, error) {
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionEdTransformWire)
	if err != nil {
		slog.Error("failed to load the grid material", "error", err)
		return nil, errors.New("failed to load the material")
	}
	var wireframe *rendering.Mesh
	switch g.Shape {
	case engine_entity_data_physics.ShapeBox:
		wireframe = rendering.NewMeshWireCube(host.MeshCache(), "rigidbody_gizmo", matrix.ColorWhite())
	case engine_entity_data_physics.ShapeSphere:
		wireframe = rendering.NewMeshWireSphereLatLon(host.MeshCache(), 1, 8, 8)
	case engine_entity_data_physics.ShapeCapsule:
		wireframe = rendering.NewMeshCapsule(host.MeshCache(), g.Radius, g.Height, 10, 3)
	case engine_entity_data_physics.ShapeCylinder:
		rad := g.Extent.X() / 2
		height := g.Extent.Y()
		wireframe = rendering.NewMeshWireCylinder(host.MeshCache(), rad, height, 5, 1)
	case engine_entity_data_physics.ShapeCone:
		wireframe = rendering.NewMeshWireCone(host.MeshCache(), g.Radius, g.Height, 5, 1)
	}
	if wireframe == nil {
		slog.Error("missing shape for rigid body wireframe")
		return nil, errors.New("could not select the correct shape")
	}
	sd := shader_data_registry.Create(material.Shader.ShaderDataName())
	gsd := sd.(*shader_data_registry.ShaderDataEdTransformWire)
	gsd.Color = matrix.NewColor(0, 1, 0, 1)
	host.Drawings.AddDrawing(rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Material:   material,
		Mesh:       wireframe,
		ShaderData: gsd,
		Transform:  transform,
		ViewCuller: &host.Cameras.Primary,
	})
	return gsd, nil
}

func (c *RigidBodyEntityDataRenderer) Hide(host *engine.Host, target *editor_stage_manager.StageEntity, _ *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("RigidBodyEntityDataRenderer.Hide").End()
	if d, ok := c.Wireframes[target]; ok {
		d.ShaderData.Destroy()
		delete(c.Wireframes, target)
	}
}

func (g *RigidBodyGizmo) reloadData(data *entity_data_binding.EntityDataEntry) bool {
	e := data.FieldValueByName("Extent").(matrix.Vec3)
	m := data.FieldValueByName("Mass").(float32)
	r := data.FieldValueByName("Radius").(float32)
	h := data.FieldValueByName("Height").(float32)
	i := data.FieldValueByName("IsStatic").(bool)
	s := engine_entity_data_physics.Shape(data.FieldValueByName("Shape").(int))
	changed := g.Shape != s ||
		(!g.Extent.Equals(e) &&
			(s == engine_entity_data_physics.ShapeBox ||
				s == engine_entity_data_physics.ShapeCylinder)) ||
		(g.Radius != r &&
			(s == engine_entity_data_physics.ShapeSphere ||
				s == engine_entity_data_physics.ShapeCapsule ||
				s == engine_entity_data_physics.ShapeCone)) ||
		(g.Height != h &&
			(s == engine_entity_data_physics.ShapeCapsule ||
				s == engine_entity_data_physics.ShapeCone))
	g.Extent = e
	g.Mass = m
	g.Radius = r
	g.Height = h
	g.IsStatic = i
	g.Shape = s
	return changed
}
