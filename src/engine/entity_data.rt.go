//go:build !editor

package engine

import (
	"encoding/gob"
	"reflect"
)

type EntityData interface {
	Init(entity *Entity, host *Host)
}

func RegisterEntityData(value EntityData) {
	t := reflect.TypeOf(value)
	name := "*" + t.PkgPath() + "." + t.Name()
	gob.RegisterName(name, value)
}
