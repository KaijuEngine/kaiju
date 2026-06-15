/******************************************************************************/
/* shader_data.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package sprite

import (
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type ShaderData struct {
	rendering.ShaderDataBase
	UVs     matrix.Vec4
	FgColor matrix.Color
}

func (ShaderData) Size() int {
	return int(rendering.ShaderBaseDataSize +
		unsafe.Sizeof(ShaderData{}.UVs) +
		unsafe.Sizeof(ShaderData{}.FgColor))
}
