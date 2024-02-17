package rendering

import (
	"slices"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

type bufferTrash struct {
	delay    int
	pool     vk.DescriptorPool
	sets     [maxFramesInFlight]vk.DescriptorSet
	buffers  [maxFramesInFlight]vk.Buffer
	memories [maxFramesInFlight]vk.DeviceMemory
}

type bufferDestroyer struct {
	device *vk.Device
	trash  []bufferTrash
	dbg    *debugVulkan
}

func newBufferDestroyer(device *vk.Device, dbg *debugVulkan) bufferDestroyer {
	return bufferDestroyer{
		device: device,
		dbg:    dbg,
	}
}

func (b *bufferDestroyer) Add(pd bufferTrash) {
	b.trash = append(b.trash, pd)
}

func (b *bufferDestroyer) Purge() {
	for len(b.trash) > 0 {
		b.Cycle()
	}
}

func (b *bufferDestroyer) Cycle() {
	if len(b.trash) == 0 {
		return
	}
	for i := len(b.trash) - 1; i >= 0; i-- {
		pd := &b.trash[i]
		pd.delay--
		if pd.delay == 0 {
			for j := range maxFramesInFlight {
				vk.DestroyBuffer(*b.device, pd.buffers[j], nil)
				b.dbg.remove(uintptr(unsafe.Pointer(pd.buffers[j])))
				vk.FreeMemory(*b.device, pd.memories[j], nil)
				b.dbg.remove(uintptr(unsafe.Pointer(pd.memories[j])))
			}
			if pd.pool != vk.DescriptorPool(vk.NullHandle) {
				vk.FreeDescriptorSets(*b.device, pd.pool, uint32(len(pd.sets)), &pd.sets[0])
			}
			// TODO:  Does this need to be ordered delete?
			b.trash = slices.Delete(b.trash, i, i+1)
		}
	}
}
