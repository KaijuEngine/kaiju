//go:build !vulkanValidation

package rendering

type debugVulkan struct{}

func debugVulkanNew() debugVulkan           { return debugVulkan{} }
func (d debugVulkan) add(handle uintptr)    {}
func (d debugVulkan) remove(handle uintptr) {}
func (d debugVulkan) print()                {}
