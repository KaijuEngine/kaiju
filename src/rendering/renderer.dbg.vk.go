//go:build !js && !OPENGL

package rendering

import (
	"fmt"
	"kaiju/klib"
)

type debugVulkan map[uintptr]string

func debugVulkanNew() debugVulkan {
	return make(debugVulkan)
}

func (d debugVulkan) add(handle uintptr) {
	d[handle] = klib.TraceString(fmt.Sprintf("VK Resource %x leak", handle))
}

func (d debugVulkan) remove(handle uintptr) {
	delete(d, handle)
}

func (d debugVulkan) print() {
	for _, trace := range d {
		fmt.Println(trace)
	}
}
