/******************************************************************************/
/* entity_id_references.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stages

import (
	"reflect"

	"github.com/KaijuEngine/uuid"

	"kaijuengine.com/engine"
)

var entityIdType = reflect.TypeFor[engine.EntityId]()

// RegenerateEntityIds replaces every non-empty entity id in desc and its
// children, returning the old-to-new id map for the duplicated subtree.
func RegenerateEntityIds(desc *EntityDescription) map[engine.EntityId]engine.EntityId {
	idMap := make(map[engine.EntityId]engine.EntityId)
	var regenerate func(d *EntityDescription)
	regenerate = func(d *EntityDescription) {
		if d.Id != "" {
			oldId := engine.EntityId(d.Id)
			newId := engine.EntityId(uuid.NewString())
			idMap[oldId] = newId
			d.Id = string(newId)
		}
		for i := range d.Children {
			regenerate(&d.Children[i])
		}
	}
	regenerate(desc)
	return idMap
}

// RegenerateEntityIdsAndRewriteReferences regenerates subtree ids and rewrites
// EntityId fields that point at entities inside the same duplicated subtree.
func RegenerateEntityIdsAndRewriteReferences(desc *EntityDescription) map[engine.EntityId]engine.EntityId {
	idMap := RegenerateEntityIds(desc)
	RewriteEntityIdReferences(desc, idMap)
	return idMap
}

// RewriteEntityIdReferences rewrites EntityId values in entity data bindings
// when their target appears in idMap. References outside idMap are left intact.
func RewriteEntityIdReferences(desc *EntityDescription, idMap map[engine.EntityId]engine.EntityId) {
	if len(idMap) == 0 {
		return
	}
	var rewrite func(d *EntityDescription)
	rewrite = func(d *EntityDescription) {
		for i := range d.DataBinding {
			rewriteEntityDataBindingReferences(&d.DataBinding[i], idMap)
		}
		for i := range d.RawDataBinding {
			d.RawDataBinding[i] = rewriteEntityIdsInData(d.RawDataBinding[i], idMap)
		}
		for i := range d.Children {
			rewrite(&d.Children[i])
		}
	}
	rewrite(desc)
}

func rewriteEntityDataBindingReferences(binding *EntityDataBinding, idMap map[engine.EntityId]engine.EntityId) {
	for name, value := range binding.Fields {
		if id, ok := value.(engine.EntityId); ok {
			if newId, exists := idMap[id]; exists {
				binding.Fields[name] = newId
			}
			continue
		}
		if s, ok := value.(string); ok && isRegisteredEntityIdField(binding.RegistraionKey, name) {
			if newId, exists := idMap[engine.EntityId(s)]; exists {
				binding.Fields[name] = newId
			}
		}
	}
}

func isRegisteredEntityIdField(registrationKey, fieldName string) bool {
	data, ok := engine.DebugEntityDataRegistry[registrationKey]
	if !ok {
		return false
	}
	t := reflect.TypeOf(data)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	field, ok := t.FieldByName(fieldName)
	return ok && field.Type == entityIdType
}

func rewriteEntityIdsInData(data any, idMap map[engine.EntityId]engine.EntityId) any {
	if data == nil {
		return nil
	}
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		rewriteEntityIdsInValue(v, idMap)
		return data
	}
	cpy := reflect.New(v.Type()).Elem()
	cpy.Set(v)
	rewriteEntityIdsInValue(cpy, idMap)
	return cpy.Interface()
}

func rewriteEntityIdsInValue(v reflect.Value, idMap map[engine.EntityId]engine.EntityId) {
	if !v.IsValid() {
		return
	}
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		elem := v.Elem()
		if elem.Kind() == reflect.Pointer {
			rewriteEntityIdsInValue(elem, idMap)
			return
		}
		cpy := reflect.New(elem.Type()).Elem()
		cpy.Set(elem)
		rewriteEntityIdsInValue(cpy, idMap)
		if cpy.Type().AssignableTo(v.Type()) {
			v.Set(cpy)
		} else if cpy.Type().Implements(v.Type()) {
			v.Set(cpy)
		}
		return
	}
	if v.Type() == entityIdType {
		if !v.CanSet() {
			return
		}
		id := v.Interface().(engine.EntityId)
		if newId, exists := idMap[id]; exists {
			v.Set(reflect.ValueOf(newId))
		}
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			rewriteEntityIdsInValue(v.Elem(), idMap)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			rewriteEntityIdsInValue(v.Field(i), idMap)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			rewriteEntityIdsInValue(v.Index(i), idMap)
		}
	}
}
