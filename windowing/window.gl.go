//go:build OPENGL

package windowing

import (
	"kaiju/rendering"
)

func selectRenderer() (rendering.Renderer, error) {
	return rendering.NewGLRenderer(), nil
}
