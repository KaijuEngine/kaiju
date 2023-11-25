//go:build !editor

package bootstrap

import (
	"kaiju/engine"
	"kaiju/source"
)

func Main(host *engine.Host) {
	println("Starting runtime")
	source.Main(host)
}
