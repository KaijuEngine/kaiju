/******************************************************************************/
/* rigid_body_entity_data_test.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_physics

import (
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/matrix"
)

func TestTerrainRigidBodyUsesTerrainModelBounds(t *testing.T) {
	entity := engine.NewEntity(nil)
	entity.Transform.SetPosition(matrix.NewVec3(1, 2, 3))
	model, err := terrain.NewModel(terrain.TerrainConfig{
		Resolution:    2,
		WorldSize:     matrix.NewVec2(8, 6),
		MinHeight:     -3,
		MaxHeight:     9,
		InitialHeight: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	entity.AddNamedData("Terrain", model)
	body := RigidBodyEntityData{Shape: ShapeTerrain, Mass: 5}.gravitonRigidBody(entity, nil)
	if body.Collision.Shape.Type != graviton.ShapeTypeTerrain {
		t.Fatalf("expected terrain shape, got %v", body.Collision.Shape.Type)
	}
	if body.Collision.Terrain == nil {
		t.Fatal("expected terrain collision")
	}
	if !body.IsStatic() {
		t.Fatal("expected terrain body to be static")
	}
	bounds := body.WorldAABB()
	if !matrix.Vec3ApproxTo(bounds.Min(), matrix.NewVec3(-3, -1, 0), 0.0001) {
		t.Fatalf("expected world terrain min -3,-1,0, got %v", bounds.Min())
	}
	if !matrix.Vec3ApproxTo(bounds.Max(), matrix.NewVec3(5, 11, 6), 0.0001) {
		t.Fatalf("expected world terrain max 5,11,6, got %v", bounds.Max())
	}
}
