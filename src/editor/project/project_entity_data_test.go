/******************************************************************************/
/* project_entity_data_test.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
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
