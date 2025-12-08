/******************************************************************************/
/* pool_group.go                                                              */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

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
		for j := range p.pools[i].takenLen {
			each(&p.pools[i].elements[p.pools[i].taken[j]])
		}
	}
}
