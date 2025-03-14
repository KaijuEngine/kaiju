//go:build !editor

package collision_module

import "kaiju/engine"

func init() {
	engine.RegisterEntityData(&OOBBModuleBinding{})
}
