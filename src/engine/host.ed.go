//go:build editor

package engine

type EditorEntities []*Entity

func newEditorEntities() EditorEntities {
	return make([]*Entity, 0)
}

func (e EditorEntities) TickCleanup() {
	for _, t := range e {
		t.TickCleanup()
	}
}

func (e EditorEntities) ResetDirty() {
	for _, t := range e {
		t.Transform.ResetDirty()
	}
}

func (host *Host) addEntity(entity *Entity) {
	if host.inEditorEntity {
		host.editorEntities = append(host.editorEntities, entity)
	} else {
		host.entities = append(host.entities, entity)
	}
}

func (host *Host) addEntities(entities ...*Entity) {
	if host.inEditorEntity {
		host.editorEntities = append(host.editorEntities, entities...)
	} else {
		host.entities = append(host.entities, entities...)
	}
}
