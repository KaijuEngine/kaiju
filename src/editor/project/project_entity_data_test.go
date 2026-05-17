/******************************************************************************/
/* project_entity_data_test.go                                                */
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

package project

import (
	"testing"

	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/engine_entity_data/engine_entity_data_terrain"
)

func TestEnsureBuiltInEntityDataBindingsIncludesTerrainAndPhysics(t *testing.T) {
	p := Project{}
	p.ensureBuiltInEntityDataBindings()

	terrainCount := 0
	for i := range p.entityData {
		if p.entityData[i].RegisterKey == engine_entity_data_terrain.BindingKey() {
			terrainCount++
			if len(p.entityData[i].Fields) == 0 {
				t.Fatal("expected terrain fallback binding to include fields")
			}
			if len(p.entityData[i].FieldGens) != len(p.entityData[i].Fields) {
				t.Fatalf("expected field gen count %d, got %d",
					len(p.entityData[i].Fields), len(p.entityData[i].FieldGens))
			}
		}
	}
	if terrainCount != 1 {
		t.Fatalf("expected one terrain binding, got %d", terrainCount)
	}
	physicsCount := 0
	for i := range p.entityData {
		if p.entityData[i].RegisterKey == engine_entity_data_physics.BindingKey() {
			physicsCount++
			if len(p.entityData[i].Fields) == 0 {
				t.Fatal("expected physics fallback binding to include fields")
			}
			if len(p.entityData[i].FieldGens) != len(p.entityData[i].Fields) {
				t.Fatalf("expected physics field gen count %d, got %d",
					len(p.entityData[i].Fields), len(p.entityData[i].FieldGens))
			}
		}
	}
	if physicsCount != 1 {
		t.Fatalf("expected one physics binding, got %d", physicsCount)
	}

	p.ensureBuiltInEntityDataBindings()
	terrainCount = 0
	physicsCount = 0
	for i := range p.entityData {
		if p.entityData[i].RegisterKey == engine_entity_data_terrain.BindingKey() {
			terrainCount++
		}
		if p.entityData[i].RegisterKey == engine_entity_data_physics.BindingKey() {
			physicsCount++
		}
	}
	if terrainCount != 1 {
		t.Fatalf("expected repeated ensure to keep one terrain binding, got %d", terrainCount)
	}
	if physicsCount != 1 {
		t.Fatalf("expected repeated ensure to keep one physics binding, got %d", physicsCount)
	}
}
