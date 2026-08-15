/******************************************************************************/
/* vk_render_pass_test.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"
	"unsafe"
)

func TestActiveSubpassDescriptorSetUsesCommandSetDescriptors(t *testing.T) {
	t.Parallel()

	defaultHandle := unsafe.Pointer(new(byte))
	targetHandle := unsafe.Pointer(new(byte))
	renderPass := RenderPass{
		subpasses: []RenderPassSubpass{{}},
	}
	renderPass.subpasses[0].descriptorSets[1].handle = defaultHandle
	targetSets := make([][maxFramesInFlight]GPUDescriptorSet, 1)
	targetSets[0][1].handle = targetHandle
	renderPass.activeCmds = &RenderPassCommandSet{
		subpassDescriptorSets: targetSets,
	}

	got := renderPass.activeSubpassDescriptorSet(0, 1)
	if got.handle != targetHandle {
		t.Fatalf("active target descriptor = %p, want %p", got.handle, targetHandle)
	}
}

func TestActiveSubpassDescriptorSetUsesDefaultDescriptors(t *testing.T) {
	t.Parallel()

	defaultHandle := unsafe.Pointer(new(byte))
	renderPass := RenderPass{
		subpasses: []RenderPassSubpass{{}},
	}
	renderPass.subpasses[0].descriptorSets[2].handle = defaultHandle

	got := renderPass.activeSubpassDescriptorSet(0, 2)
	if got.handle != defaultHandle {
		t.Fatalf("default descriptor = %p, want %p", got.handle, defaultHandle)
	}
}
