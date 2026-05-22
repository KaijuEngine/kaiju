/******************************************************************************/
/* gpu_application_vulkan.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/klib"
	vk "kaijuengine.com/rendering/vulkan"
)

func init() {
	klib.Must(vk.SetDefaultGetInstanceProcAddr())
	klib.Must(vk.Init())
}
