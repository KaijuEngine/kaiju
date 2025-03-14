package collision_module

import (
	"kaiju/engine"
	"kaiju/engine/collision_system"
)

const (
	CollisionShapeEntityDataName = "CollisionShape"
)

func addShape(e *engine.Entity, host *engine.Host, shape collision_system.Shape, shapeData any) {
	man := host.CollisionManager()
	s := collision_system.RegisterCollisionShape(man, &e.Transform, shape, shapeData)
	e.AddNamedData(CollisionShapeEntityDataName, s)
	e.OnDestroy.Add(func() { man.Remove(s) })
}
