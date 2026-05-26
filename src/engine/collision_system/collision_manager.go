/******************************************************************************/
/* collision_manager.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package collision_system

import (
	"kaijuengine.com/engine/pooling"
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
