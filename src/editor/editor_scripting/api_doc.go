/******************************************************************************/
/* api_doc.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_scripting

import (
	"reflect"

	"kaijuengine.com/matrix"
	"kaijuengine.com/plugins"
)

func RegenerateEditorLuaAPI(apiFile string) error {
	types := []reflect.Type{
		reflect.TypeFor[matrix.Vec2](),
		reflect.TypeFor[matrix.Vec3](),
		reflect.TypeFor[matrix.Vec4](),
		reflect.TypeFor[matrix.Quaternion](),
		reflect.TypeFor[matrix.Mat3](),
		reflect.TypeFor[matrix.Mat4](),
	}
	types = append(types, AutomationTypes()...)
	return plugins.RegenerateAPIForTypes(apiFile, types)
}
