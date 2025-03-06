package pooling

import "sync"

type PoolGroupId = int

type PoolGroup[T any] struct {
	pools []*Pool[T]
	lock  sync.RWMutex
}

func (p *PoolGroup[T]) Count() int { return len(p.pools) }

func (p *PoolGroup[T]) selectPool() (*Pool[T], PoolGroupId) {
	for i := range p.pools {
		if p.pools[i].availableLen > 0 {
			return p.pools[i], i
		}
	}
	p.pools = append(p.pools, &Pool[T]{})
	last := len(p.pools) - 1
	p.pools[last].init()
	return p.pools[last], last
}

func (p *PoolGroup[T]) Clear() {
	for i := range p.pools {
		for j, idx := ElementsInPool-1, 0; i >= 0; i-- {
			p.pools[i].available[idx] = PoolIndex(j)
			idx++
		}
		p.pools[i].availableLen = ElementsInPool
		p.pools[i].takenLen = 0
	}
	// TODO:  Should the pools be cleared [:0] instead?
}

func (p *PoolGroup[T]) Add() (elm *T, poolId PoolGroupId, elmId PoolIndex) {
	pool, poolId := p.selectPool()
	lastId := pool.available[pool.availableLen-1]
	pool.availableLen--
	pool.taken[pool.takenLen] = lastId
	pool.takenLen++
	return &pool.elements[lastId], poolId, lastId
}

func (p *PoolGroup[T]) Remove(poolIndex PoolGroupId, elementId PoolIndex) {
	if len(p.pools) <= poolIndex {
		return
	}
	pool := p.pools[poolIndex]
	for i := range pool.takenLen {
		if pool.taken[i] == elementId {
			// Swap this index with the last one
			pool.taken[i], pool.taken[pool.takenLen-1] =
				pool.taken[pool.takenLen-1], pool.taken[i]
			pool.takenLen--
			pool.available[pool.availableLen] = elementId
			pool.availableLen++
			break
		}
	}
}

func (p *PoolGroup[T]) Reserve(additionalElements int) {
	for i := range p.pools {
		additionalElements -= p.pools[i].availableLen
	}
	if additionalElements <= 0 {
		return
	}
	addPools := additionalElements / ElementsInPool
	for range addPools {
		p.pools = append(p.pools, &Pool[T]{})
		p.pools[len(p.pools)-1].init()
	}
}

func (p *PoolGroup[T]) Each(each func(elm *T)) {
	for i := range p.pools {
		for j := range p.pools[i].takenLen {
			each(&p.pools[i].elements[p.pools[i].taken[j]])
		}
	}
}
