/******************************************************************************/
/* lighting_information.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package lighting

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type LightingInformation struct {
	Lights LightCollection
}

func NewLightingInformation(lightCacheCapacity int) LightingInformation {
	return LightingInformation{
		Lights: LightCollection{
			Cache: make([]rendering.Light, lightCacheCapacity),
		},
	}
}

func (l *LightingInformation) Update(point matrix.Vec3) {
	l.Lights.UpdateCache(point)
}
