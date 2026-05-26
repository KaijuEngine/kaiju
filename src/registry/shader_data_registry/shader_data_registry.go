/******************************************************************************/
/* shader_data_registry.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_data_registry

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/debug"
	"kaijuengine.com/rendering"
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

func register(factory func() rendering.DrawInstance, names ...string) {
	for _, name := range names {
		_, ok := registry[name]
		debug.Assert(!ok, fmt.Sprintf("duplicate name '%s' used for shader_data_registry", name))
		if ok {
			return
		}
		registry[name] = factory
	}
}
