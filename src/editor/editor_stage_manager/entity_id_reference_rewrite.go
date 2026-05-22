/******************************************************************************/
/* entity_id_reference_rewrite.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"reflect"

	"kaijuengine.com/editor/project"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/stages"
)

var editorEntityIdType = reflect.TypeFor[engine.EntityId]()

func regenerateEntityIdsAndRewriteReferences(desc *stages.EntityDescription, proj *project.Project) map[engine.EntityId]engine.EntityId {
	idMap := stages.RegenerateEntityIds(desc)
	stages.RewriteEntityIdReferences(desc, idMap)
	rewriteProjectEntityIdReferences(desc, proj, idMap)
	return idMap
}

func rewriteProjectEntityIdReferences(desc *stages.EntityDescription, proj *project.Project, idMap map[engine.EntityId]engine.EntityId) {
	if proj == nil || len(idMap) == 0 {
		return
	}
	var rewrite func(d *stages.EntityDescription)
	rewrite = func(d *stages.EntityDescription) {
		for i := range d.DataBinding {
			binding := &d.DataBinding[i]
			g, ok := proj.EntityDataBinding(binding.RegistraionKey)
			if !ok {
				continue
			}
			for _, field := range g.Fields {
				if field.Type != editorEntityIdType {
					continue
				}
				value, exists := binding.Fields[field.Name]
				if !exists {
					continue
				}
				if newId, ok := idMap[entityIdFromBindingValue(value)]; ok {
					binding.Fields[field.Name] = newId
				}
			}
		}
		for i := range d.Children {
			rewrite(&d.Children[i])
		}
	}
	rewrite(desc)
}

func entityIdFromBindingValue(value any) engine.EntityId {
	switch v := value.(type) {
	case engine.EntityId:
		return v
	case string:
		return engine.EntityId(v)
	default:
		return ""
	}
}
