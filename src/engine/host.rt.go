//go:build !editor

package engine

type EditorEntities struct{}

func newEditorEntities() EditorEntities {
	return EditorEntities{}
}

func (e EditorEntities) TickCleanup() {}
func (e EditorEntities) ResetDirty()  {}

func (host *Host) addEntity(entity *Entity) {
	host.entities = append(host.entities, entity)
}

func (host *Host) addEntities(entities ...*Entity) {
	host.entities = append(host.entities, entities...)
}
