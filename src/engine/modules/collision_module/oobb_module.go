package collision_module

import (
	"kaiju/engine/collision"
	"kaiju/engine"
	"kaiju/engine/collision_system"
	"kaiju/matrix"
)

type OOBBModuleBinding struct {
	Center matrix.Vec3
	Extent matrix.Vec3
}

func (b *OOBBModuleBinding) Init(e *engine.Entity, host *engine.Host) {
	shapeData := collision.OOBB{
		Center:      b.Center,
		Extent:      b.Extent,
		Orientation: matrix.Mat3Identity(),
	}
	addShape(e, host, collision_system.ShapeOOBB, shapeData)
}
