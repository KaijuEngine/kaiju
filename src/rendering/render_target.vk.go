//go:build !js && !OPENGL

package rendering

import "errors"

type VKRenderTarget struct {
	oit oitFrameBuffers
}

func newRenderTarget(renderer Renderer) (VKRenderTarget, error) {
	vr := renderer.(*Vulkan)
	target := VKRenderTarget{}
	if !target.oit.createImages(vr) {
		return target, errors.New("failed to create render target images")
	}
	if !target.oit.createBuffers(vr, &vr.oitPass) {
		return target, errors.New("failed to create render target buffers")
	}
	return target, nil
}

func (r *VKRenderTarget) reset(vr *Vulkan) {
	r.oit.reset(vr)
}
