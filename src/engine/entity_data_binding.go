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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"kaiju/matrix"
	"log/slog"
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
			} else if ivs, ok := v.([]float32); ok && len(ivs) == f.Len() {
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
	case reflect.Float32, reflect.Int, reflect.Uint, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f.Set(reflect.ValueOf(v).Convert(f.Type()))
	default:
		if f.IsValid() {
			f.Set(reflect.ValueOf(v))
		}
	}
}
