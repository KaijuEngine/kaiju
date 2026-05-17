/******************************************************************************/
/* rigid_body_entity_data_test.go                                             */
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
