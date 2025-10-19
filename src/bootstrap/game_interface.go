package bootstrap

import (
	"kaiju/engine"
	"kaiju/engine/assets"
	"reflect"
)

type GameInterface interface {
	Launch(*engine.Host)
	PluginRegistry() []reflect.Type
	ContentDatabase() (assets.Database, error)
}
