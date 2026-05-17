/******************************************************************************/
/* rigid_body_data_binding.go                                                 */
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

package engine_entity_data_physics

import (
	"log/slog"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/engine_entity_data/content_id"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

var bindingKey = ""

type Shape int

const (
	ShapeBox Shape = iota
	ShapeSphere
	ShapeCapsule
	ShapeCylinder
	ShapeCone
	ShapeMesh
	ShapeTerrain
)

func init() {
	pod.Register(Shape(0))
	pod.Register(content_id.Mesh(""))
	engine.RegisterEntityData(RigidBodyEntityData{})
}

func BindingKey() string {
	if bindingKey == "" {
		bindingKey = pod.QualifiedNameForLayout(RigidBodyEntityData{})
	}
	return bindingKey
}

type RigidBodyEntityData struct {
	AssetKey content_id.Mesh
	Extent   matrix.Vec3 `default:"1,1,1"`
	Mass     float32     `default:"1"`
	Radius   float32     `default:"1"`
	Height   float32     `default:"1"`
	Shape    Shape
	IsStatic bool
}

func (r RigidBodyEntityData) Init(e *engine.Entity, host *engine.Host) {
	host.StartPhysics()
	body := r.gravitonRigidBody(e, host)
	host.Physics().AddEntity(e, body)
}

func (r RigidBodyEntityData) EntityDataInitPhase() engine.EntityDataPhase {
	return engine.EntityDataPhasePhysicsBody
}

func (r RigidBodyEntityData) gravitonRigidBody(e *engine.Entity, host *engine.Host) *graviton.RigidBody {
	body := &graviton.RigidBody{}
	body.Transform.SetupRawTransform()
	body.Transform.SetPosition(e.Transform.Position())
	body.Transform.SetRotation(e.Transform.Rotation())
	if r.Shape == ShapeMesh {
		body.SetStaticMesh(r.gravitonMesh(host))
		return body
	}
	if r.Shape == ShapeTerrain {
		body.SetStaticTerrain(r.gravitonTerrain(e))
		return body
	}
	shape := r.gravitonShape(e.Transform.Scale())
	// Scale is baked into the shape dimensions to match the existing behavior.
	body.SetShape(shape)
	if r.IsStatic {
		body.SetStatic()
	} else {
		mass := matrix.Float(r.Mass)
		body.SetDynamic(mass, graviton.CalculateLocalInertia(shape, mass))
	}
	return body
}

func (r RigidBodyEntityData) gravitonMesh(host *engine.Host) *graviton.MeshCollision {
	if r.AssetKey == "" {
		slog.Warn("graviton mesh physics shape has no asset key")
		return graviton.NewMeshCollision(nil)
	}
	km, err := kaiju_mesh.ReadMesh(string(r.AssetKey), host)
	if err != nil {
		slog.Error("failed to read graviton mesh physics shape", "assetKey", r.AssetKey, "error", err)
		return graviton.NewMeshCollision(nil)
	}
	positions := make([]matrix.Vec3, len(km.Verts))
	for i := range km.Verts {
		positions[i] = km.Verts[i].Position
	}
	mesh := graviton.NewMeshCollisionFromVertices(positions, km.Indexes)
	if len(mesh.Triangles) == 0 {
		slog.Warn("graviton mesh physics shape has no triangles", "assetKey", r.AssetKey)
	}
	return mesh
}

func (r RigidBodyEntityData) gravitonTerrain(e *engine.Entity) *graviton.TerrainCollision {
	if e == nil {
		slog.Warn("graviton terrain physics shape has no entity")
		return nil
	}
	for _, data := range e.NamedData("Terrain") {
		model, ok := data.(*terrain.Terrain)
		if !ok {
			continue
		}
		collision := model.NewCollision()
		if collision == nil {
			slog.Error("failed to create graviton terrain physics shape")
			return nil
		}
		return collision
	}
	slog.Warn("graviton terrain physics shape has no terrain entity data")
	return nil
}

func (r RigidBodyEntityData) gravitonShape(scale matrix.Vec3) graviton.Shape {
	scale = matrix.Vec3Abs(scale)
	switch r.Shape {
	case ShapeBox:
		return graviton.NewBoxShape(r.Extent.Multiply(scale))
	case ShapeSphere:
		radius := matrix.Float(r.Radius) * scale.LongestAxisValue()
		return graviton.NewSphereShape(radius)
	case ShapeCapsule:
		radius := matrix.Float(r.Radius) * scale.LongestAxisValue()
		height := matrix.Float(r.Height) * scale.Y()
		return graviton.NewCapsuleShape(radius, height)
	case ShapeCylinder:
		size := r.Extent.Multiply(scale)
		radius := matrix.Max(size.X(), size.Z())
		height := size.Y() * 2
		return graviton.NewCylinderShape(radius, height)
	case ShapeCone:
		radius := matrix.Float(r.Radius) * scale.LongestAxisValue()
		height := matrix.Float(r.Height) * scale.Y()
		return graviton.NewConeShape(radius, height)
	case ShapeMesh:
		return graviton.NewMeshShape(graviton.NewAABB(matrix.Vec3Zero(), matrix.Vec3Zero()))
	case ShapeTerrain:
		return graviton.NewTerrainShape(graviton.NewAABB(matrix.Vec3Zero(), matrix.Vec3Zero()))
	}
	return graviton.NewBoxShape(r.Extent.Multiply(scale))
}
