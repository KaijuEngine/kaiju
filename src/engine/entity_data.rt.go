//go:build !editor

package engine

import (
	"encoding/gob"
	"errors"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type EntityData interface {
	Init(entity *Entity, host *Host)
}

func RegisterEntityData(value EntityData) error {
	_, fileName, _, ok := runtime.Caller(1)
	if !ok {
		return errors.New("failed to get the caller's information")
	}
	pkg := strings.ReplaceAll(filepath.Dir(fileName), "\\", "/")
	const lookFor = "/source/"
	const pkgPrefix = "kaiju/source/"
	start := strings.Index(pkg, lookFor)
	if start == -1 {
		return errors.New("failed to find the source package")
	}
	pkg = "*" + pkgPrefix + pkg[start+len(lookFor):]
	typ := reflect.TypeOf(value).Elem()
	pkg += "." + typ.Name()
	gob.RegisterName(pkg, value)
	return nil
}
