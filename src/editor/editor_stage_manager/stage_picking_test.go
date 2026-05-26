/******************************************************************************/
/* stage_picking_test.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"testing"

	"kaijuengine.com/editor/memento"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

func TestStageManagerAssignPickIDNonzeroUniqueStable(t *testing.T) {
	manager := &StageManager{}
	a := &StageEntity{}
	b := &StageEntity{}

	aID := manager.AssignPickID(a)
	bID := manager.AssignPickID(b)

	if aID == 0 || bID == 0 {
		t.Fatalf("pick IDs must be nonzero, got %d and %d", aID, bID)
	}
	if aID == bID {
		t.Fatalf("pick IDs must be unique, got %d for both entities", aID)
	}
	if got := manager.AssignPickID(a); got != aID {
		t.Fatalf("pick ID changed from %d to %d", aID, got)
	}
}

func TestStageManagerEntityByPickIDIgnoresLockedAndDeleted(t *testing.T) {
	manager := &StageManager{}
	entity := &StageEntity{}
	id := manager.AssignPickID(entity)

	if got, ok := manager.EntityByPickID(id); !ok || got != entity {
		t.Fatalf("EntityByPickID(%d) = %v, %v; want entity, true", id, got, ok)
	}
	entity.Lock()
	if got, ok := manager.EntityByPickID(id); ok || got != nil {
		t.Fatalf("locked entity resolved by pick ID: %v, %v", got, ok)
	}
	entity.Unlock()
	entity.isDeleted = true
	if got, ok := manager.EntityByPickID(id); ok || got != nil {
		t.Fatalf("deleted entity resolved by pick ID: %v, %v", got, ok)
	}
}

func TestStageManagerEntityByPickIDDeterministic(t *testing.T) {
	manager := &StageManager{}
	a := &StageEntity{}
	b := &StageEntity{}
	aID := manager.AssignPickID(a)
	bID := manager.AssignPickID(b)

	for i := 0; i < 10; i++ {
		if got, ok := manager.EntityByPickID(aID); !ok || got != a {
			t.Fatalf("iteration %d: first entity resolution = %v, %v", i, got, ok)
		}
		if got, ok := manager.EntityByPickID(bID); !ok || got != b {
			t.Fatalf("iteration %d: second entity resolution = %v, %v", i, got, ok)
		}
	}
}

func TestStageManagerEntitiesByPickIDsUsesStageOrder(t *testing.T) {
	manager := &StageManager{}
	a := &StageEntity{}
	b := &StageEntity{}
	c := &StageEntity{}
	manager.entities = []*StageEntity{a, b, c}
	aID := manager.AssignPickID(a)
	bID := manager.AssignPickID(b)
	cID := manager.AssignPickID(c)

	got := manager.EntitiesByPickIDs([]uint32{cID, aID, bID, aID})
	want := []*StageEntity{a, b, c}
	if len(got) != len(want) {
		t.Fatalf("picked entity count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("picked entity %d = %p, want %p", i, got[i], want[i])
		}
	}
}

func TestStageManagerEntitiesByPickIDsSkipsLockedDeletedAndStaleIDs(t *testing.T) {
	manager := &StageManager{}
	a := &StageEntity{}
	b := &StageEntity{}
	c := &StageEntity{}
	manager.entities = []*StageEntity{a, b, c}
	aID := manager.AssignPickID(a)
	bID := manager.AssignPickID(b)
	cID := manager.AssignPickID(c)
	b.Lock()
	c.isDeleted = true

	got := manager.EntitiesByPickIDs([]uint32{cID, bID, aID, 99})
	if len(got) != 1 || got[0] != a {
		t.Fatalf("picked entities = %v, want only a", got)
	}
}

func TestStageManagerPickingDrawingOnlyForMeshBackedEntities(t *testing.T) {
	manager := &StageManager{}
	material := &rendering.Material{}

	if _, ok := manager.newPickingDrawing(&StageEntity{}, material); ok {
		t.Fatalf("entity without mesh should not create a picking drawing")
	}

	entity := &StageEntity{}
	entity.StageData.Mesh = &rendering.Mesh{}
	draw, ok := manager.newPickingDrawing(entity, material)
	if !ok {
		t.Fatalf("mesh-backed entity should create a picking drawing")
	}
	if draw.Layer != rendering.RenderLayerEditorPicking {
		t.Fatalf("picking drawing layer = %v, want editor picking", draw.Layer)
	}
	if draw.Mesh != entity.StageData.Mesh {
		t.Fatalf("picking drawing used the wrong mesh")
	}
	if draw.ShaderData != entity.StageData.PickingShaderData {
		t.Fatalf("picking shader data should be stored on the stage entity")
	}
	sd := draw.ShaderData.(*shader_data_registry.ShaderDataEditorPicking)
	if sd.PickID == 0 || sd.PickID != entity.PickID {
		t.Fatalf("picking shader data PickID = %d, entity PickID = %d", sd.PickID, entity.PickID)
	}
	if _, ok := manager.newPickingDrawing(entity, material); ok {
		t.Fatalf("entity with existing picking shader data should not create a duplicate drawing")
	}
}

func TestStageManagerSelectEntitiesReplaceAppendToggle(t *testing.T) {
	history := &memento.History{}
	manager := &StageManager{history: history}
	a := &StageEntity{}
	b := &StageEntity{}
	a.Init(nil)
	b.Init(nil)

	manager.SelectEntities([]*StageEntity{a}, SelectionModeReplace)
	if !manager.IsSelected(a) || manager.IsSelected(b) {
		t.Fatalf("replace selection should select only a")
	}

	manager.SelectEntities([]*StageEntity{b}, SelectionModeAppend)
	if !manager.IsSelected(a) || !manager.IsSelected(b) {
		t.Fatalf("append selection should keep a and add b")
	}

	manager.SelectEntities([]*StageEntity{a}, SelectionModeToggle)
	if manager.IsSelected(a) || !manager.IsSelected(b) {
		t.Fatalf("toggle selection should remove a and keep b")
	}
}

func TestStageManagerSelectEntitiesEmptyReplaceClearsSelection(t *testing.T) {
	history := &memento.History{}
	manager := &StageManager{history: history}
	entity := &StageEntity{}
	entity.Init(nil)
	manager.SelectEntities([]*StageEntity{entity}, SelectionModeReplace)

	manager.SelectEntities(nil, SelectionModeReplace)

	if manager.HasSelection() {
		t.Fatalf("empty replace selection should clear selection")
	}
}
