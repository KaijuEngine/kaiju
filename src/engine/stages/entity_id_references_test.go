/******************************************************************************/
/* entity_id_references_test.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stages

import (
	"encoding/json"
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
)

type entityIdReferenceTestData struct {
	ConnectedEntityId engine.EntityId
}

var entityIdReferenceTestKey = pod.QualifiedNameForLayout(entityIdReferenceTestData{})

func init() {
	_ = engine.RegisterEntityData(entityIdReferenceTestData{})
}

func (d entityIdReferenceTestData) Init(e *engine.Entity, host *engine.Host) {
	target := host.EntityById(d.ConnectedEntityId)
	if target == nil {
		e.AddNamedData("resolvedTarget", engine.EntityId(""))
		return
	}
	e.AddNamedData("resolvedTarget", target.Id())
}

func TestRegenerateEntityIdsRewritesTwoBodyConstraintTarget(t *testing.T) {
	desc := EntityDescription{
		Id: "source",
		DataBinding: []EntityDataBinding{{
			RegistraionKey: "test.constraint",
			Fields: map[string]any{
				"ConnectedEntityId": engine.EntityId("target"),
			},
		}},
		Children: []EntityDescription{{Id: "target"}},
	}

	idMap := RegenerateEntityIdsAndRewriteReferences(&desc)

	got := desc.DataBinding[0].Fields["ConnectedEntityId"]
	if got != idMap["target"] {
		t.Fatalf("expected target to be rewritten to %q, got %q", idMap["target"], got)
	}
}

func TestRegenerateEntityIdsRewritesInternalChainReferences(t *testing.T) {
	desc := EntityDescription{
		Id: "link-a",
		DataBinding: []EntityDataBinding{{
			RegistraionKey: "test.link",
			Fields: map[string]any{
				"ConnectedEntityId": engine.EntityId("link-b"),
			},
		}},
		Children: []EntityDescription{{
			Id: "link-b",
			DataBinding: []EntityDataBinding{{
				RegistraionKey: "test.link",
				Fields: map[string]any{
					"ConnectedEntityId": engine.EntityId("link-c"),
				},
			}},
			Children: []EntityDescription{{Id: "link-c"}},
		}},
	}

	idMap := RegenerateEntityIdsAndRewriteReferences(&desc)

	if got := desc.DataBinding[0].Fields["ConnectedEntityId"]; got != idMap["link-b"] {
		t.Fatalf("expected first link to target duplicate link-b %q, got %q", idMap["link-b"], got)
	}
	if got := desc.Children[0].DataBinding[0].Fields["ConnectedEntityId"]; got != idMap["link-c"] {
		t.Fatalf("expected second link to target duplicate link-c %q, got %q", idMap["link-c"], got)
	}
}

func TestRegenerateEntityIdsKeepsExternalAnchorReferences(t *testing.T) {
	desc := EntityDescription{
		Id: "source",
		DataBinding: []EntityDataBinding{{
			RegistraionKey: "test.constraint",
			Fields: map[string]any{
				"ConnectedEntityId": engine.EntityId("external-anchor"),
			},
		}},
		Children: []EntityDescription{{Id: "target"}},
	}

	RegenerateEntityIdsAndRewriteReferences(&desc)

	if got := desc.DataBinding[0].Fields["ConnectedEntityId"]; got != engine.EntityId("external-anchor") {
		t.Fatalf("expected external reference to stay unchanged, got %q", got)
	}
}

func TestTemplateEntityIdsInstantiateWithFreshInternalReferences(t *testing.T) {
	desc := EntityDescription{
		Id: "source",
		DataBinding: []EntityDataBinding{{
			RegistraionKey: entityIdReferenceTestKey,
			Fields: map[string]any{
				"ConnectedEntityId": "target",
			},
		}},
		RawDataBinding: []any{
			entityIdReferenceTestData{ConnectedEntityId: "target"},
		},
		Children: []EntityDescription{{Id: "target"}},
	}
	idMap := RegenerateEntityIdsAndRewriteReferences(&desc)

	host := engine.NewHost("test", nil, nil)
	stage := Stage{Entities: []EntityDescription{desc}}
	res := stage.Load(host)
	source := res.EntitiesById[idMap["source"]]
	if source == nil {
		t.Fatal("expected duplicated source entity to load")
	}
	resolved := source.NamedData("resolvedTarget")
	if len(resolved) == 0 {
		t.Fatal("expected template constraint data to resolve target")
	}
	if resolved[0] != idMap["target"] {
		t.Fatalf("expected resolved duplicate target %q, got %q", idMap["target"], resolved[0])
	}
}

func TestSaveReloadAfterDuplicationKeepsRewrittenConstraintReference(t *testing.T) {
	desc := EntityDescription{
		Id: "source",
		DataBinding: []EntityDataBinding{{
			RegistraionKey: "test.constraint",
			Fields: map[string]any{
				"ConnectedEntityId": engine.EntityId("target"),
			},
		}},
		Children: []EntityDescription{{Id: "target"}},
	}
	idMap := RegenerateEntityIdsAndRewriteReferences(&desc)
	stage := Stage{Entities: []EntityDescription{desc}}

	data, err := json.Marshal(stage.ToMinimized())
	if err != nil {
		t.Fatal(err)
	}
	var minimized StageJson
	if err := json.Unmarshal(data, &minimized); err != nil {
		t.Fatal(err)
	}
	var reloaded Stage
	reloaded.FromMinimized(minimized)

	got := reloaded.Entities[0].DataBinding[0].Fields["ConnectedEntityId"]
	if got != string(idMap["target"]) {
		t.Fatalf("expected reloaded target to stay %q, got %q", idMap["target"], got)
	}
}
