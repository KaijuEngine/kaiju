/******************************************************************************/
/* descriptor_texture_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"
	"unsafe"
)

var descriptorTextureHandles [3]byte

func TestDescriptorTextureOrFallbackUsesReadyTexture(t *testing.T) {
	t.Parallel()

	texture := testDescriptorTexture()
	fallback := testDescriptorTexture()

	if got := descriptorTextureOrFallback(texture, fallback); got != texture {
		t.Fatalf("descriptor texture = %p, want primary %p", got, texture)
	}
}

func TestDescriptorTextureOrFallbackUsesFallbackForInvalidTexture(t *testing.T) {
	t.Parallel()

	invalid := testDescriptorTexture()
	invalid.RenderId.View.Reset()
	fallback := testDescriptorTexture()

	if got := descriptorTextureOrFallback(invalid, fallback); got != fallback {
		t.Fatalf("descriptor texture = %p, want fallback %p", got, fallback)
	}
}

func TestDescriptorTextureOrFallbackRejectsInvalidFallback(t *testing.T) {
	t.Parallel()

	invalid := testDescriptorTexture()
	invalid.RenderId.Sampler.Reset()

	if got := descriptorTextureOrFallback(invalid, &Texture{}); got != nil {
		t.Fatalf("descriptor texture = %p, want nil", got)
	}
}

func testDescriptorTexture() *Texture {
	return &Texture{
		RenderId: TextureId{
			Image:   GPUImage{GPUHandle{handle: unsafe.Pointer(&descriptorTextureHandles[0])}},
			View:    GPUImageView{GPUHandle{handle: unsafe.Pointer(&descriptorTextureHandles[1])}},
			Sampler: GPUSampler{GPUHandle{handle: unsafe.Pointer(&descriptorTextureHandles[2])}},
		},
	}
}
