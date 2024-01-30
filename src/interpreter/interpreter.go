package interpreter

import "reflect"

var Symbols map[string]map[string]reflect.Value = make(map[string]map[string]reflect.Value)
