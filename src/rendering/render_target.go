package rendering

import "kaiju/matrix"

type RenderTarget interface{}

type RenderTargetDraw struct {
	Target RenderTarget
	Rect   matrix.Vec4
}

func NewRenderTarget(renderer Renderer) (RenderTarget, error) {
	return newRenderTarget(renderer)
}
