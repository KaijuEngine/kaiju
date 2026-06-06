/******************************************************************************/
/* rigid_body_entity_data_renderer.go                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package data_binding_renderer

import (
	"errors"
	"log/slog"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/engine_entity_data/content_id"
	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

type rigidBodyGizmo struct {
	ShaderData rendering.DrawInstance
	Extent     matrix.Vec3
	Mass       float32
	Radius     float32
	Height     float32
	IsStatic   bool
	Shape      engine_entity_data_physics.Shape
	AssetKey   content_id.Mesh
	Mesh       graviton.AABB
	HasMesh    bool
	Terrain    graviton.AABB
	HasTerrain bool
}

type RigidBodyEntityDataRenderer struct {
	Wireframes map[*editor_stage_manager.StageEntity]rigidBodyGizmo
}

func init() {
	AddRenderer(engine_entity_data_physics.BindingKey(), &RigidBodyEntityDataRenderer{
		Wireframes: make(map[*editor_stage_manager.StageEntity]rigidBodyGizmo),
	})
}

func (c *RigidBodyEntityDataRenderer) Attached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	// defer tracing.NewRegion("RigidBodyEntityDataRenderer.Attached").End()
	// icon := commonAttached(host, manager, target, "light.png")
	if _, ok := c.Wireframes[target]; ok {
		slog.Error("there is an internal error in state for the editor's RigidBodyEntityDataRenderer, show was called before any hide happened. Double selected the same target?")
		c.Detatched(host, manager, target, data)
	}
	g := rigidBodyGizmo{}
	g.reloadData(data, target, host)
	var err error
	if g.ShaderData, err = rigidBodyLoadWireframe(host, g, &target.Transform); err == nil {
		c.Wireframes[target] = g
		g.ShaderData.Deactivate()
	}
	target.OnDestroy.Add(func() {
		c.Detatched(host, manager, target, data)
	})
}

func (c *RigidBodyEntityDataRenderer) Detatched(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("RigidBodyEntityDataRenderer.Detatched").End()
	if d, ok := c.Wireframes[target]; ok {
		if d.ShaderData != nil {
			d.ShaderData.Destroy()
		}
		delete(c.Wireframes, target)
	}
}

func (c *RigidBodyEntityDataRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("RigidBodyEntityDataRenderer.Show").End()
	if d, ok := c.Wireframes[target]; ok && d.ShaderData != nil {
		d.ShaderData.Activate()
	}
}

func (c *RigidBodyEntityDataRenderer) Hide(host *engine.Host, target *editor_stage_manager.StageEntity, _ *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("RigidBodyEntityDataRenderer.Hide").End()
	if d, ok := c.Wireframes[target]; ok && d.ShaderData != nil {
		d.ShaderData.Deactivate()
	}
}

func (c *RigidBodyEntityDataRenderer) Update(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	if g, ok := c.Wireframes[target]; ok {
		if g.reloadData(data, target, host) {
			if g.ShaderData != nil {
				g.ShaderData.Destroy()
			}
			var err error
			if g.ShaderData, err = rigidBodyLoadWireframe(host, g, &target.Transform); err == nil {
				c.Wireframes[target] = g
			} else {
				g.ShaderData = nil
				c.Wireframes[target] = g
			}
		}
	}
}

func rigidBodyLoadWireframe(host *engine.Host, g rigidBodyGizmo, transform *matrix.Transform) (rendering.DrawInstance, error) {
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
		wireframe = rendering.NewMeshWireSphere(host.MeshCache(), g.Radius+0.001, 8, 8)
	case engine_entity_data_physics.ShapeCapsule:
		wireframe = rendering.NewMeshCapsule(host.MeshCache(), g.Radius, g.Height, 10, 3)
	case engine_entity_data_physics.ShapeCylinder:
		rad := g.Extent.X() / 2
		height := g.Extent.Y()
		wireframe = rendering.NewMeshWireCylinder(host.MeshCache(), rad, height, 5, 1)
	case engine_entity_data_physics.ShapeCone:
		wireframe = rendering.NewMeshWireCone(host.MeshCache(), g.Radius, g.Height, 5, 1)
	case engine_entity_data_physics.ShapeMesh:
		wireframe = rendering.NewMeshWireCube(host.MeshCache(), "rigidbody_mesh_gizmo", matrix.ColorWhite())
	case engine_entity_data_physics.ShapeTerrain:
		wireframe = rendering.NewMeshWireCube(host.MeshCache(), "rigidbody_terrain_gizmo", matrix.ColorWhite())
	}
	if wireframe == nil {
		slog.Error("missing shape for rigid body wireframe")
		return nil, errors.New("could not select the correct shape")
	}
	sd := shader_data_registry.Create(material.Shader.DrawInstanceDataName())
	gsd := sd.(*shader_data_registry.ShaderDataEdTransformWire)
	gsd.Color = matrix.NewColor(0, 1, 0, 1)
	if (g.Shape == engine_entity_data_physics.ShapeMesh && !g.HasMesh) ||
		(g.Shape == engine_entity_data_physics.ShapeTerrain && !g.HasTerrain) {
		gsd.Color = matrix.ColorYellow()
	}
	if g.Shape == engine_entity_data_physics.ShapeBox {
		model := matrix.Mat4Identity()
		model.Scale(g.Extent.Scale(2))
		gsd.SetModel(model)
	} else if g.Shape == engine_entity_data_physics.ShapeMesh {
		model := matrix.Mat4Identity()
		model.Translate(g.Mesh.Center)
		model.Scale(g.Mesh.Size())
		gsd.SetModel(model)
	} else if g.Shape == engine_entity_data_physics.ShapeTerrain {
		model := matrix.Mat4Identity()
		model.Translate(g.Terrain.Center)
		model.Scale(g.Terrain.Size())
		gsd.SetModel(model)
	}
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       wireframe,
		ShaderData: gsd,
		Transform:  transform,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	})
	return gsd, nil
}

func rigidBodyTerrainBounds(target *editor_stage_manager.StageEntity) (graviton.AABB, bool) {
	if target != nil {
		for _, data := range target.NamedData("Terrain") {
			model, ok := data.(*terrain.Terrain)
			if !ok {
				continue
			}
			if collision := model.NewCollision(); collision != nil {
				return collision.LocalBounds(), true
			}
		}
	}
	return graviton.NewAABB(matrix.Vec3Zero(), matrix.NewVec3XYZ(0.5)), false
}

func rigidBodyMeshBounds(host *engine.Host, assetKey content_id.Mesh) (graviton.AABB, bool) {
	if host == nil || assetKey == "" {
		return graviton.NewAABB(matrix.Vec3Zero(), matrix.NewVec3XYZ(0.5)), false
	}
	km, err := kaiju_mesh.ReadMesh(string(assetKey), host)
	if err != nil || len(km.Verts) == 0 {
		return graviton.NewAABB(matrix.Vec3Zero(), matrix.NewVec3XYZ(0.5)), false
	}
	points := make([]matrix.Vec3, len(km.Verts))
	for i := range km.Verts {
		points[i] = km.Verts[i].Position
	}
	return graviton.AABBFromPoints(points), true
}

func (g *rigidBodyGizmo) reloadData(data *entity_data_binding.EntityDataEntry, target *editor_stage_manager.StageEntity, hosts ...*engine.Host) bool {
	var host *engine.Host
	if len(hosts) > 0 {
		host = hosts[0]
	}
	assetKey := data.FieldValueByName("AssetKey").(content_id.Mesh)
	e := data.FieldValueByName("Extent").(matrix.Vec3)
	m := data.FieldValueByName("Mass").(float32)
	r := data.FieldValueByName("Radius").(float32)
	height := data.FieldValueByName("Height").(float32)
	i := data.FieldValueByName("IsStatic").(bool)
	s := engine_entity_data_physics.Shape(data.FieldValueByName("Shape").(int))
	meshBounds, hasMesh := graviton.NewAABB(matrix.Vec3Zero(), matrix.NewVec3XYZ(0.5)), false
	if s == engine_entity_data_physics.ShapeMesh {
		meshBounds, hasMesh = rigidBodyMeshBounds(host, assetKey)
	}
	terrainBounds, hasTerrain := graviton.NewAABB(matrix.Vec3Zero(), matrix.NewVec3XYZ(0.5)), false
	if s == engine_entity_data_physics.ShapeTerrain {
		terrainBounds, hasTerrain = rigidBodyTerrainBounds(target)
	}
	changed := g.Shape != s ||
		g.AssetKey != assetKey ||
		(!g.Extent.Equals(e) &&
			(s == engine_entity_data_physics.ShapeBox ||
				s == engine_entity_data_physics.ShapeCylinder)) ||
		(g.Radius != r &&
			(s == engine_entity_data_physics.ShapeSphere ||
				s == engine_entity_data_physics.ShapeCapsule ||
				s == engine_entity_data_physics.ShapeCone)) ||
		(g.Height != height &&
			(s == engine_entity_data_physics.ShapeCapsule ||
				s == engine_entity_data_physics.ShapeCone)) ||
		(s == engine_entity_data_physics.ShapeMesh &&
			(g.HasMesh != hasMesh || g.Mesh != meshBounds)) ||
		(s == engine_entity_data_physics.ShapeTerrain &&
			(g.HasTerrain != hasTerrain || g.Terrain != terrainBounds))
	g.Extent = e
	g.Mass = m
	g.Radius = r
	g.Height = height
	g.IsStatic = i
	g.Shape = s
	g.AssetKey = assetKey
	g.Mesh = meshBounds
	g.HasMesh = hasMesh
	g.Terrain = terrainBounds
	g.HasTerrain = hasTerrain
	return changed
}
