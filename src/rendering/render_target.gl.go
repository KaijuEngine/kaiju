//go:build OPENGL

package rendering

type GLRenderTarget struct {
}

func newRenderTarget(renderer Renderer) (GLRenderTarget, error) {
	return GLRenderTarget{}, nil
}
