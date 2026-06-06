/******************************************************************************/
/* pool_group.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pooling

import (
	"sync"

	"kaijuengine.com/build"
	"kaijuengine.com/platform/concurrent"
)

type PoolGroupId = int // This is actually just 3 bytes
const MaxPoolGroupId = 0x00FFFFFF

type PoolGroup[T any] struct {
	pools []*Pool[T]
	lock  sync.RWMutex
}

func (p *PoolGroup[T]) Count() int { return len(p.pools) }

func (p *PoolGroup[T]) ElementCount() int {
	count := 0
	for i := range p.pools {
		count += p.pools[i].takenLen
	}
	return count
}

func (p *PoolGroup[T]) selectPool() (*Pool[T], PoolGroupId) {
	for i := range p.pools {
		if p.pools[i].availableLen > 0 {
			return p.pools[i], i
		}
	}
	p.pools = append(p.pools, &Pool[T]{})
	last := len(p.pools) - 1
	p.pools[last].init()
	if build.Debug {
		if len(p.pools) > MaxPoolGroupId {
			panic("the pool amount has gone beyond the allowed limit")
		}
	}
	return p.pools[last], last
}

func (p *PoolGroup[T]) Clear() {
	for i := range p.pools {
		for j, idx := ElementsInPool-1, 0; j >= 0; j-- {
			p.pools[i].available[idx] = PoolIndex(j)
			idx++
		}
		p.pools[i].availableLen = ElementsInPool
		p.pools[i].takenLen = 0
	}
	// TODO:  Should the pools be cleared instead?
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
	if build.Debug {
		if len(p.pools) > MaxPoolGroupId {
			panic("the pool amount has gone beyond the allowed limit")
		}
	}
}

// Each will iterate through every element, both active and inactive element in
// the pool and supply it to the expression that was supplied to this function call
func (p *PoolGroup[T]) All(each func(elm *T)) {
	for i := range p.pools {
		for j := range p.pools[i].elements {
			each(&p.pools[i].elements[j])
		}
	}
}

// Each will iterate through each active element in the pool and supply it to
// the expression that was supplied to this function call
func (p *PoolGroup[T]) Each(each func(elm *T)) {
	for i := range p.pools {
		// Loop in reverse so that it's safe to remove elements during this call
		for j := p.pools[i].takenLen - 1; j >= 0; j-- {
			each(&p.pools[i].elements[p.pools[i].taken[j]])
		}
	}
}

func (p *PoolGroup[T]) EachParallel(workName string, workGroup *concurrent.WorkGroup, threads *concurrent.Threads, each func(elm *T)) {
	for i := range p.pools {
		for j := range p.pools[i].takenLen {
			workGroup.Add(workName, func() { each(&p.pools[i].elements[p.pools[i].taken[j]]) })
		}
	}
	workGroup.Execute(workName, threads)
}

// ConditionalEach iterates over each active element in the pool group, invoking the
// provided callback function `each`. If the callback returns false for any element,
// the iteration stops early. This allows callers to break out of the loop based on
// a condition while still processing elements in order of their allocation.
func (p *PoolGroup[T]) ConditionalEach(each func(elm *T) bool) {
outerLoop:
	for i := range p.pools {
		for j := p.pools[i].takenLen - 1; j >= 0; j-- {
			if !each(&p.pools[i].elements[p.pools[i].taken[j]]) {
				break outerLoop
			}
		}
	}
}
