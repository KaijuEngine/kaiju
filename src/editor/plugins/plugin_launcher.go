//go:build editor

package plugins

import (
	"kaiju/editor/interfaces"
	"kaiju/matrix"
	"log/slog"
	"path/filepath"
	"reflect"

	"github.com/dop251/goja"
)

type jsLogType = int

const (
	jsLogTypeDebug = jsLogType(iota)
	jsLogTypeWarn
	jsLogTypeError
)

type jsvm struct {
	runtime      *goja.Runtime
	instanceMap  map[int64]any
	nextInstance int64
}

func (vm *jsvm) throwError(message string) goja.Value {
	if errorConstructor, ok := goja.AssertFunction(vm.runtime.Get("Error")); ok {
		if res, err := errorConstructor(goja.Null(), vm.runtime.ToValue(message)); err == nil {
			return res
		}
	}
	return nil
}

func (vm *jsvm) logging(call goja.FunctionCall, logType jsLogType) goja.Value {
	if len(call.Arguments) < 1 {
		return vm.throwError("Missing arguments")
	}
	msg := call.Arguments[0].String()
	prints := make([]any, len(call.Arguments[1:]))
	for i, a := range call.Arguments[1:] {
		prints[i] = a.Export()
	}
	switch logType {
	case jsLogTypeDebug:
		slog.Debug(msg, prints...)
	case jsLogTypeWarn:
		slog.Warn(msg, prints...)
	case jsLogTypeError:
		slog.Error(msg, prints...)
	}
	return nil
}

func (vm *jsvm) jsDebug(call goja.FunctionCall) goja.Value {
	return vm.logging(call, jsLogTypeDebug)
}

func (vm *jsvm) jsWarn(call goja.FunctionCall) goja.Value {
	return vm.logging(call, jsLogTypeWarn)
}

func (vm *jsvm) jsError(call goja.FunctionCall) goja.Value {
	return vm.logging(call, jsLogTypeError)
}

func reflectStructToJS[T any](vm *jsvm) {
	t := reflect.TypeFor[T]()
	name := t.Name()
	proto := vm.runtime.NewObject()
	constructorFunc := vm.runtime.ToValue(func(call goja.ConstructorCall) *goja.Object {
		instance := reflect.New(t).Interface()
		id := vm.nextInstance
		vm.nextInstance++
		vm.instanceMap[id] = instance
		call.This.Set("_id", id)
		return call.This
	})
	constructorObj := constructorFunc.ToObject(vm.runtime)
	constructorObj.Set("prototype", proto)
	vm.runtime.Set(name, constructorFunc)
	pt := reflect.PointerTo(t)
	methods := make([]reflect.Method, 0, t.NumMethod()+pt.NumMethod())
	for i := range t.NumMethod() {
		methods = append(methods, t.Method(i))
	}
	for i := range pt.NumMethod() {
		methods = append(methods, pt.Method(i))
	}
	for _, m := range methods {
		methodName := m.Name
		mt := m.Type
		argTypes := make([]reflect.Type, mt.NumIn()-1)
		for j := range mt.NumIn() - 1 {
			argTypes[j] = mt.In(j + 1)
		}
		var dynFields func(r reflect.Value) goja.Value
		dynFields = func(r reflect.Value) goja.Value {
			rt := r.Type()
			mCount := rt.NumMethod()
			kind := rt.Kind()
			switch kind {
			case reflect.Array:
			case reflect.Pointer:
				mCount += rt.Elem().NumMethod()
			default:
				mCount += reflect.PointerTo(rt).NumMethod()
			}
			if mCount > 0 {
				s := vm.runtime.Get(rt.Name())
				v, _ := vm.runtime.New(s)
				goV := vm.instanceMap[v.Get("_id").ToInteger()]
				goR := reflect.ValueOf(goV).Elem()
				switch kind {
				case reflect.Array:
					for i := range r.Len() {
						goR.Index(i).Set(r.Index(i))
					}
				default:
					for j := range rt.NumField() {
						// TODO:  Handle if this is a pointer
						f := r.Field(j)
						goR.Field(j).Set(r.Field(j))
						v.Set(f.Type().Name(), dynFields(f).Export())
					}
				}
				return v
			} else {
				return vm.runtime.ToValue(r.Interface())
			}
		}
		proto.Set(methodName, func(call goja.FunctionCall) goja.Value {
			obj := call.This.ToObject(vm.runtime)
			idVal := obj.Get("_id")
			id := idVal.ToInteger()
			if instance, ok := vm.instanceMap[id]; ok {
				v := reflect.ValueOf(instance)
				args := make([]reflect.Value, len(call.Arguments))
				for j := range call.Arguments {
					args[j] = reflect.ValueOf(call.Arguments[j].Export())
					args[j] = args[j].Convert(argTypes[j])
				}
				res := v.MethodByName(methodName).Call(args)
				if len(res) > 1 {
					slog.Error("reflected go method to JS contains too many returns", "count", len(res), "method", methodName)
				} else if len(res) == 0 {
					return nil
				}
				return dynFields(res[0])
			}
			return nil
		})
	}
}

func LaunchPlugins(ed interfaces.Editor) error {
	const plugins = "editor/plugins"
	js, err := ed.Host().AssetDatabase().ReadText(filepath.Join(plugins, "test.js"))
	if err != nil {
		return err
	}
	vm := &jsvm{
		runtime:      goja.New(),
		instanceMap:  make(map[int64]any),
		nextInstance: 1,
	}
	reflectStructToJS[matrix.Vec2](vm)
	reflectStructToJS[matrix.Vec3](vm)
	reflectStructToJS[matrix.Vec3i](vm)
	reflectStructToJS[matrix.Vec4](vm)
	reflectStructToJS[matrix.Vec4i](vm)
	reflectStructToJS[matrix.Quaternion](vm)
	reflectStructToJS[matrix.Mat3](vm)
	reflectStructToJS[matrix.Mat4](vm)
	vm.runtime.Set("debug", vm.jsDebug)
	vm.runtime.Set("warn", vm.jsWarn)
	vm.runtime.Set("error", vm.jsError)
	if _, err := vm.runtime.RunString(js); err != nil {
		return err
	}
	return nil
}
