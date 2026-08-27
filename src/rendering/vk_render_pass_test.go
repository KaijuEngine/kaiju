/******************************************************************************/
/* vk_render_pass_test.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"
	"unsafe"

	"kaijuengine.com/rendering/gpu_types"
)

func TestActiveSubpassDescriptorSetUsesCommandSetDescriptors(t *testing.T) {
	t.Parallel()

	defaultHandle := unsafe.Pointer(new(byte))
	targetHandle := unsafe.Pointer(new(byte))
	renderPass := RenderPass{
		subpasses: []RenderPassSubpass{{}},
	}
	renderPass.subpasses[0].descriptorSets[1].Handle = defaultHandle
	targetSets := make([][maxFramesInFlight]gpu_types.DescriptorSet, 1)
	targetSets[0][1].Handle = targetHandle
	renderPass.activeCmds = &RenderPassCommandSet{
		subpassDescriptorSets: targetSets,
	}

	got := renderPass.activeSubpassDescriptorSet(0, 1)
	if got.Handle != targetHandle {
		t.Fatalf("active target descriptor = %p, want %p", got.Handle, targetHandle)
	}
}

func TestActiveSubpassDescriptorSetUsesDefaultDescriptors(t *testing.T) {
	t.Parallel()

	defaultHandle := unsafe.Pointer(new(byte))
	renderPass := RenderPass{
		subpasses: []RenderPassSubpass{{}},
	}
	renderPass.subpasses[0].descriptorSets[2].Handle = defaultHandle

	got := renderPass.activeSubpassDescriptorSet(0, 2)
	if got.Handle != defaultHandle {
		t.Fatalf("default descriptor = %p, want %p", got.Handle, defaultHandle)
	}
}
