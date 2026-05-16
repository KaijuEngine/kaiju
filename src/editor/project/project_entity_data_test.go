package project

import (
	"testing"

	"kaijuengine.com/engine_entity_data/engine_entity_data_terrain"
)

func TestEnsureBuiltInEntityDataBindingsIncludesTerrain(t *testing.T) {
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

	p.ensureBuiltInEntityDataBindings()
	terrainCount = 0
	for i := range p.entityData {
		if p.entityData[i].RegisterKey == engine_entity_data_terrain.BindingKey() {
			terrainCount++
		}
	}
	if terrainCount != 1 {
		t.Fatalf("expected repeated ensure to keep one terrain binding, got %d", terrainCount)
	}
}
