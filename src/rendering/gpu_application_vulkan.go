package rendering

import (
	"kaiju/klib"
	vk "kaiju/rendering/vulkan"
)

func init() {
	klib.Must(vk.SetDefaultGetInstanceProcAddr())
	klib.Must(vk.Init())
}
