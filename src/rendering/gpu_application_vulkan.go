package rendering

import (
	"kaijuengine.com/klib"
	vk "kaijuengine.com/rendering/vulkan"
)

func init() {
	klib.Must(vk.SetDefaultGetInstanceProcAddr())
	klib.Must(vk.Init())
}
