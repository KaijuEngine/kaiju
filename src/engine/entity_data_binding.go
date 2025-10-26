package engine

import (
	"errors"
	"kaiju/engine/runtime/encoding/gob"
)

type EntityData interface {
	Init(entity *Entity, host *Host)
}

func RegisterEntityData(name string, value EntityData) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
		}
	}()
	gob.RegisterName(name, value)
	return err
}
