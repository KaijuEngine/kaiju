package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/editor/memento"
	"kaijuengine.com/matrix"
)

func TestRenderGraphNodeFieldValueHistoryUndoRedoTextAndNumber(t *testing.T) {
	tests := []struct {
		name      string
		fieldType renderGraphNodeFieldType
		from      renderGraphNodeFieldValue
		to        renderGraphNodeFieldValue
	}{
		{
			name:      "text",
			fieldType: renderGraphNodeFieldText,
			from:      renderGraphNodeFieldValue{Text: "roughness"},
			to:        renderGraphNodeFieldValue{Text: "metallic"},
		},
		{
			name:      "number",
			fieldType: renderGraphNodeFieldNumber,
			from:      renderGraphNodeFieldValue{Text: "0.25"},
			to:        renderGraphNodeFieldValue{Text: "0.75"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			history, _, node, field := newFieldValueHistoryTest(test.fieldType, test.from)

			field.beginFieldValueEdit()
			node.setFieldValue(field.spec.ID, test.to)
			if !field.commitFieldValueEdit() {
				t.Fatal("commitFieldValueEdit() should add history")
			}

			history.Undo()
			requireFieldValue(t, node.FieldValue(field.spec.ID), test.from)
			history.Redo()
			requireFieldValue(t, node.FieldValue(field.spec.ID), test.to)
		})
	}
}

func TestRenderGraphNodeFieldValueHistoryVectorCollapsesAndClonesParts(t *testing.T) {
	from := renderGraphNodeFieldValue{Parts: []string{"0", "0", "0"}}
	to := renderGraphNodeFieldValue{Parts: []string{"1", "2", "0"}}
	history, _, node, field := newFieldValueHistoryTest(renderGraphNodeFieldVector3, from)

	field.beginFieldValueEdit()
	value := node.FieldValue(field.spec.ID)
	value.Parts[0] = "1"
	node.setFieldValue(field.spec.ID, value)
	value = node.FieldValue(field.spec.ID)
	value.Parts[1] = "2"
	node.setFieldValue(field.spec.ID, value)
	if !field.commitFieldValueEdit() {
		t.Fatal("commitFieldValueEdit() should add history")
	}

	last, ok := history.Last()
	if !ok {
		t.Fatal("history should have an entry")
	}
	stored, ok := last.(*renderGraphNodeFieldValueHistory)
	if !ok {
		t.Fatalf("history entry type = %T, want *renderGraphNodeFieldValueHistory", last)
	}
	current := node.values[field.spec.ID]
	current.Parts[0] = "99"
	if stored.to.Parts[0] == "99" {
		t.Fatal("history to.Parts aliases the node value parts")
	}

	history.Undo()
	requireFieldValue(t, node.FieldValue(field.spec.ID), from)
	history.Redo()
	requireFieldValue(t, node.FieldValue(field.spec.ID), to)
}

func TestRenderGraphNodeFieldValueHistorySkipsNoOpEdits(t *testing.T) {
	history, _, _, field := newFieldValueHistoryTest(renderGraphNodeFieldText,
		renderGraphNodeFieldValue{Text: "unchanged"})

	field.beginFieldValueEdit()
	if field.commitFieldValueEdit() {
		t.Fatal("commitFieldValueEdit() should skip unchanged values")
	}
	if _, ok := history.Last(); ok {
		t.Fatal("unchanged value should not add history")
	}
}

func TestRenderGraphNodeFieldValueHistorySubmitThenBlurDoesNotDuplicate(t *testing.T) {
	history, _, node, field := newFieldValueHistoryTest(renderGraphNodeFieldText,
		renderGraphNodeFieldValue{Text: "before"})

	field.beginFieldValueEdit()
	node.setFieldValue(field.spec.ID, renderGraphNodeFieldValue{Text: "after"})
	if !field.commitFieldValueEdit() {
		t.Fatal("submit commit should add history")
	}
	last, ok := history.Last()
	if !ok {
		t.Fatal("submit commit should leave a history entry")
	}
	field.scheduleDeferredFieldValueCommit()

	if next, ok := history.Last(); !ok || next != last {
		t.Fatal("blur after submit should not add another history entry")
	}
}

func TestRenderGraphNodeFieldValueHistoryEscapeStyleRestoreSkipsHistory(t *testing.T) {
	from := renderGraphNodeFieldValue{Text: "focus-start"}
	history, _, node, field := newFieldValueHistoryTest(renderGraphNodeFieldText, from)

	field.beginFieldValueEdit()
	node.setFieldValue(field.spec.ID, renderGraphNodeFieldValue{Text: "typed"})
	node.setFieldValue(field.spec.ID, from)
	if field.commitFieldValueEdit() {
		t.Fatal("restored value should not add history")
	}
	requireFieldValue(t, node.FieldValue(field.spec.ID), from)
	if _, ok := history.Last(); ok {
		t.Fatal("restored value should not leave a history entry")
	}
}

func TestRenderGraphNodeFieldValueHistoryDiscreteChanges(t *testing.T) {
	tests := []struct {
		name      string
		fieldType renderGraphNodeFieldType
		from      renderGraphNodeFieldValue
		to        renderGraphNodeFieldValue
	}{
		{
			name:      "checkbox",
			fieldType: renderGraphNodeFieldBool,
			from:      renderGraphNodeFieldValue{Bool: false},
			to:        renderGraphNodeFieldValue{Bool: true},
		},
		{
			name:      "select",
			fieldType: renderGraphNodeFieldSelect,
			from:      renderGraphNodeFieldValue{Option: "mix"},
			to:        renderGraphNodeFieldValue{Option: "multiply"},
		},
		{
			name:      "texture",
			fieldType: renderGraphNodeFieldTexture,
			from:      renderGraphNodeFieldValue{Text: "albedo.png"},
			to:        renderGraphNodeFieldValue{Text: "normal.png"},
		},
		{
			name:      "color",
			fieldType: renderGraphNodeFieldColor,
			from:      renderGraphNodeFieldValue{Color: matrix.ColorWhite()},
			to:        renderGraphNodeFieldValue{Color: matrix.ColorRed()},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			history, _, node, field := newFieldValueHistoryTest(test.fieldType, test.from)

			if !field.commitDiscreteFieldValue(test.to) {
				t.Fatal("commitDiscreteFieldValue() should add history")
			}
			last, ok := history.Last()
			if !ok {
				t.Fatal("discrete change should leave a history entry")
			}
			if field.commitDiscreteFieldValue(test.to) {
				t.Fatal("same discrete value should not add history")
			}
			if next, ok := history.Last(); !ok || next != last {
				t.Fatal("same discrete value should not add another history entry")
			}

			history.Undo()
			requireFieldValue(t, node.FieldValue(field.spec.ID), test.from)
			history.Redo()
			requireFieldValue(t, node.FieldValue(field.spec.ID), test.to)
		})
	}
}

func newFieldValueHistoryTest(fieldType renderGraphNodeFieldType, value renderGraphNodeFieldValue) (
	*memento.History, *renderGraph, *renderGraphNode, *renderGraphNodeField,
) {
	history := &memento.History{}
	history.Initialize(8)
	graph := &renderGraph{history: history}
	node := &renderGraphNode{
		graph:  graph,
		id:     "node-value",
		typeID: "value",
		values: map[string]renderGraphNodeFieldValue{},
	}
	field := &renderGraphNodeField{
		node: node,
		spec: renderGraphNodeFieldSpec{
			ID:   "value",
			Type: fieldType,
		},
	}
	node.fields = []*renderGraphNodeField{field}
	node.setFieldValue(field.spec.ID, value)
	graph.nodes = []*renderGraphNode{node}
	return history, graph, node, field
}

func requireFieldValue(t *testing.T, got, want renderGraphNodeFieldValue) {
	t.Helper()
	if !got.Equals(want) {
		t.Fatalf("field value = %#v, want %#v", got, want)
	}
}
