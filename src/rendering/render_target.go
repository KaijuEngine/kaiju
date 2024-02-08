package rendering

type RenderTarget interface{}

func NewRenderTarget(renderer Renderer) (RenderTarget, error) {
	return newRenderTarget(renderer)
}
