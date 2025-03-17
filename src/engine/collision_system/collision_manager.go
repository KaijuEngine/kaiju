package collision_system

import (
	"kaiju/engine/pooling"
)

type Manager struct {
	pools    pooling.PoolGroup[CollisionShape]
	updateId int
}

func (m *Manager) Remove(shape *CollisionShape) {
	m.pools.Remove(shape.poolId, shape.elmId)
}

func (m *Manager) Update(deltaTime float64) {

}
