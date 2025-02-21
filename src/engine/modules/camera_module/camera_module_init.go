//go:build !editor

package camera_module

import "kaiju/engine"

func init() {
	engine.RegisterEntityData(&CameraModuleBinding{})
}
