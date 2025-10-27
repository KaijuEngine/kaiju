package shader_data_registry

import (
	"fmt"
	"kaiju/debug"
	"kaiju/rendering"
	"log/slog"
)

const fallback = "basic"

var registry = map[string]func() rendering.DrawInstance{}

func Create(name string) rendering.DrawInstance {
	r, ok := registry[name]
	if !ok {
		slog.Warn("missing shader_data_registry factory method for target, using fallback", "name", name)
		r = registry[fallback]
	}
	return r()
}

func register(name string, factory func() rendering.DrawInstance) {
	_, ok := registry[name]
	debug.Assert(!ok, fmt.Sprintf("duplicate name '%s' used for shader_data_registry", name))
	if ok {
		return
	}
	registry[name] = factory
}
