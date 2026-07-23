/******************************************************************************/
/* project_entity_data_test.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"kaijuengine.com/editor/codegen"
	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/engine_entity_data/engine_entity_data_light"
	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/engine_entity_data/engine_entity_data_terrain"
	"kaijuengine.com/matrix"
)

func TestWalkedLightEntityDataAppliesVectorDefaults(t *testing.T) {
	g := walkedLightEntityData(t)
	entry := (&entity_data_binding.EntityDataEntry{}).ReadEntityDataBindingType(g)
	if !entry.Fields[0].IsVec3() {
		t.Fatalf("Ambient field type %q was not classified as Vec3", entry.Fields[0].Type)
	}
	if got, want := entry.FieldValueByName("Ambient"), matrix.NewVec3(0.1, 0.1, 0.1); got != want {
		t.Fatalf("Ambient default = %v, want %v", got, want)
	}
	if got, want := entry.FieldValueByName("Diffuse"), matrix.Vec3One(); got != want {
		t.Fatalf("Diffuse default = %v, want %v", got, want)
	}
	if got, want := entry.FieldValueByName("Specular"), matrix.Vec3One(); got != want {
		t.Fatalf("Specular default = %v, want %v", got, want)
	}
}

func TestWalkedLightEntityDataSurvivesStageJSONRoundTrip(t *testing.T) {
	g := walkedLightEntityData(t)
	spawned := (&entity_data_binding.EntityDataEntry{}).ReadEntityDataBindingType(g)
	fields := make(map[string]any, len(spawned.Fields))
	for i := range spawned.Fields {
		fields[spawned.Fields[i].Name] = spawned.FieldValue(i)
	}
	stage := stages.Stage{Entities: []stages.EntityDescription{{
		DataBinding: []stages.EntityDataBinding{{
			RegistraionKey: g.RegisterKey,
			Fields:         fields,
		}},
	}}}
	raw, err := json.Marshal(stage.ToMinimized())
	if err != nil {
		t.Fatal(err)
	}
	var minimized stages.StageJson
	if err := json.Unmarshal(raw, &minimized); err != nil {
		t.Fatal(err)
	}
	var reloadedStage stages.Stage
	reloadedStage.FromMinimized(minimized)
	reloaded := (&entity_data_binding.EntityDataEntry{}).ReadEntityDataBindingType(g)
	for name, value := range reloadedStage.Entities[0].DataBinding[0].Fields {
		reloaded.SetFieldByName(name, value)
	}
	if got, want := reloaded.FieldValueByName("Ambient"), matrix.NewVec3(0.1, 0.1, 0.1); got != want {
		t.Fatalf("Ambient after stage round trip = %v, want %v", got, want)
	}
	if got, want := reloaded.FieldValueByName("Diffuse"), matrix.Vec3One(); got != want {
		t.Fatalf("Diffuse after stage round trip = %v, want %v", got, want)
	}
	if got, want := reloaded.FieldValueByName("Specular"), matrix.Vec3One(); got != want {
		t.Fatalf("Specular after stage round trip = %v, want %v", got, want)
	}
}

func walkedLightEntityData(t *testing.T) codegen.GeneratedType {
	t.Helper()
	srcRoot, err := os.OpenRoot(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	defer srcRoot.Close()
	bindingsRoot, err := os.OpenRoot(filepath.Join("..", "..", "engine_entity_data"))
	if err != nil {
		t.Fatal(err)
	}
	defer bindingsRoot.Close()

	gens, err := codegen.Walk(srcRoot, bindingsRoot, "kaijuengine.com")
	if err != nil {
		t.Fatal(err)
	}
	for i := range gens {
		if gens[i].RegisterKey != engine_entity_data_light.BindingKey() {
			continue
		}
		return gens[i]
	}
	t.Fatal("walked light entity data binding was not found")
	return codegen.GeneratedType{}
}

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
