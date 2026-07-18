package rendering

import "testing"

func TestGPUCapabilitiesAreConservative(t *testing.T) {
	device := GPUDevice{PhysicalDevice: GPUPhysicalDevice{
		Properties: GPUPhysicalDeviceProperties{ApiVersion: (1 << 22) | (2 << 12)},
		QueueFamilies: []GPUQueueFamily{
			{Index: 0, IsGraphics: true, IsCompute: true},
			{Index: 1, IsCompute: true},
		},
		Extensions: []GPUPhysicalDeviceExtension{
			{Name: extensionBufferDeviceAddress},
			{Name: extensionAccelerationStructure},
			{Name: extensionDeferredHostOperations},
			{Name: extensionRayQuery},
		},
		MemoryProperties: GPUPhysicalDeviceMemoryProperties{MemoryHeaps: []GPUMemoryHeap{
			{Size: 256 * 1024 * 1024, Flags: GPUMemoryHeapDeviceLocalBit},
			{Size: 64 * 1024 * 1024},
		}},
	}}
	capabilities := device.Capabilities()
	if capabilities.VulkanMajor != 1 || capabilities.VulkanMinor != 2 || capabilities.DeviceLocalMemoryMB != 256 {
		t.Fatalf("basic capabilities = %+v", capabilities)
	}
	if !capabilities.DedicatedComputeQueue || !capabilities.RayQueryAdvertised {
		t.Fatalf("advertised capabilities = %+v", capabilities)
	}
	if capabilities.RayQuery || capabilities.AccelerationStructure || capabilities.BufferDeviceAddress {
		t.Fatal("advertised but unenabled features were reported as usable")
	}
}
