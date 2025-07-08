package collision_system

import (
	"kaiju/engine/pooling"
	"kaiju/matrix"
)

type Shape = int

const (
	ShapeAABB = Shape(iota)
	ShapeOOBB
)

type CollisionShape struct {
	Transform *matrix.Transform
	ShapeData any
	Shape     Shape
	poolId    pooling.PoolGroupId
	elmId     pooling.PoolIndex
}

func RegisterCollisionShape(man *Manager, transform *matrix.Transform, shape Shape, shapeData any) *CollisionShape {
	s, pIdx, eIdx := man.pools.Add()
	*s = CollisionShape{
		Transform: transform,
		ShapeData: shapeData,
		Shape:     shape,
		poolId:    pIdx,
		elmId:     eIdx,
	}
	return s
}
