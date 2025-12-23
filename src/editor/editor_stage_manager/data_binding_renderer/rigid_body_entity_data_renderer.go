/******************************************************************************/
/* rigid_body_entity_data_renderer.go                                         */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

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

type rigidBodyGizmo struct {
	ShaderData rendering.DrawInstance
	Extent     matrix.Vec3
	Mass       float32
	Radius     float32
	Height     float32
	IsStatic   bool
	Shape      engine_entity_data_physics.Shape
}

type RigidBodyEntityDataRenderer struct {
	Wireframes map[*editor_stage_manager.StageEntity]rigidBodyGizmo
}

func init() {
	AddRenderer(engine_entity_data_physics.BindingKey, &RigidBodyEntityDataRenderer{
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
	g.reloadData(data)
	var err error
	if g.ShaderData, err = rigidBodyLoadWireframe(host, g, &target.Transform); err == nil {
		c.Wireframes[target] = g
	}
}

func (c *RigidBodyEntityDataRenderer) Detatched(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("RigidBodyEntityDataRenderer.Detatched").End()
	if d, ok := c.Wireframes[target]; ok {
		d.ShaderData.Destroy()
		delete(c.Wireframes, target)
	}
}

func (c *RigidBodyEntityDataRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("RigidBodyEntityDataRenderer.Show").End()
	if d, ok := c.Wireframes[target]; ok {
		d.ShaderData.Activate()
	}
}

func (c *RigidBodyEntityDataRenderer) Hide(host *engine.Host, target *editor_stage_manager.StageEntity, _ *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("RigidBodyEntityDataRenderer.Hide").End()
	if d, ok := c.Wireframes[target]; ok {
		d.ShaderData.Deactivate()
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
		Material:   material,
		Mesh:       wireframe,
		ShaderData: gsd,
		Transform:  transform,
		ViewCuller: &host.Cameras.Primary,
	})
	return gsd, nil
}

func (g *rigidBodyGizmo) reloadData(data *entity_data_binding.EntityDataEntry) bool {
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
