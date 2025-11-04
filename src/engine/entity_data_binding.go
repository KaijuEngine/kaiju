/******************************************************************************/
/* entity_data_binding.go                                                     */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package engine

import (
	"errors"
	"kaiju/build"
	"kaiju/engine/runtime/encoding/gob"
	"reflect"
)

var DebugEntityDataRegistry = map[string]EntityData{}

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
	if build.Debug {
		DebugEntityDataRegistry[name] = value
	}
	return err
}

func ReflectEntityDataBindingValueFromJson(v any, f reflect.Value) {
	switch f.Kind() {
	case reflect.Float32: // JSON reads float64 only
		reflectTo[float64, float32](v, f)
	case reflect.Int: // JSON reads int64 only
		reflectTo[int64, int](v, f)
	case reflect.Uint:
		reflectTo[int64, uint](v, f)
	case reflect.Int8:
		reflectTo[int64, int8](v, f)
	case reflect.Int16:
		reflectTo[int64, int16](v, f)
	case reflect.Int32:
		reflectTo[int64, int32](v, f)
	case reflect.Uint8:
		reflectTo[int64, uint8](v, f)
	case reflect.Uint16:
		reflectTo[int64, uint16](v, f)
	case reflect.Uint32:
		reflectTo[int64, uint32](v, f)
	case reflect.Uint64:
		reflectTo[int64, uint64](v, f)
	default:
		if f.IsValid() {
			f.Set(reflect.ValueOf(v))
		}
	}
}

func reflectTo[F, T any](v any, f reflect.Value) {
	if vv, ok := v.(F); ok {
		f.Set(reflect.ValueOf(vv).Convert(reflect.TypeFor[T]()))
	}
}
