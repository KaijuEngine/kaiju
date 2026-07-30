/******************************************************************************/
/* entity_data_binding.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import (
	"errors"
	"log/slog"
	"reflect"

	"kaijuengine.com/build"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/matrix"
)

var DebugEntityDataRegistry = map[string]EntityData{}

type EntityData interface {
	Init(entity *Entity, host *Host)
}

type EntityDataPhase int

const (
	EntityDataPhaseDefault EntityDataPhase = iota * 100
	EntityDataPhasePhysicsConstraint
)

const EntityDataPhasePhysicsBody = EntityDataPhaseDefault

type EntityDataInitPhaser interface {
	EntityDataInitPhase() EntityDataPhase
}

func EntityDataInitPhase(data EntityData) EntityDataPhase {
	if phased, ok := data.(EntityDataInitPhaser); ok {
		return phased.EntityDataInitPhase()
	}
	return EntityDataPhaseDefault
}

func RegisterEntityData(value EntityData) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
		}
	}()
	pod.Register(value)
	if build.Debug {
		DebugEntityDataRegistry[pod.QualifiedNameForLayout(value)] = value
	}
	return err
}

func ReflectValueFromJson(v any, f reflect.Value) {
	switch f.Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		elemType := f.Type().Elem()
		if elemType.Kind() == reflect.Float32 || elemType.Kind() == reflect.Float64 {
			if ivs, ok := v.([]interface{}); ok && len(ivs) == f.Len() {
				for i := 0; i < f.Len(); i++ {
					if num, ok := ivs[i].(float64); ok {
						f.Index(i).SetFloat(num)
					} else {
						slog.Error("invalid float in array of floats", "index", i)
					}
				}
			} else if ivs, ok := v.([]matrix.Float); ok && len(ivs) == f.Len() {
				for i := 0; i < f.Len(); i++ {
					f.Index(i).SetFloat(float64(ivs[i]))
				}
			} else if ivs, ok := v.([]float64); ok && len(ivs) == f.Len() {
				for i := 0; i < f.Len(); i++ {
					f.Index(i).SetFloat(ivs[i])
				}
			} else if vec, ok := v.(matrix.Vec2); ok {
				for i := 0; i < len(vec); i++ {
					f.Index(i).SetFloat(float64(vec[i]))
				}
			} else if vec, ok := v.(matrix.Vec3); ok {
				for i := 0; i < len(vec); i++ {
					f.Index(i).SetFloat(float64(vec[i]))
				}
			} else if vec, ok := v.(matrix.Vec4); ok {
				for i := 0; i < len(vec); i++ {
					f.Index(i).SetFloat(float64(vec[i]))
				}
			} else if vec, ok := v.(matrix.Color); ok {
				for i := 0; i < len(vec); i++ {
					f.Index(i).SetFloat(float64(vec[i]))
				}
			}
		}
	default:
		if f.IsValid() {
			f.Set(reflect.ValueOf(v).Convert(f.Type()))
		}
	}
}
