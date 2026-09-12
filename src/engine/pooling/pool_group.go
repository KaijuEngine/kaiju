/******************************************************************************/
/* pool_group.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pooling

import (
	"math/bits"
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
		p.pools[i].taken = [takenUint64Count]uint64{}
		p.pools[i].availableLen = ElementsInPool
		p.pools[i].takenLen = 0
	}
	// TODO:  Should the pools be cleared instead?
}

func (p *PoolGroup[T]) Add() (elm *T, poolId PoolGroupId, elmId PoolIndex) {
	pool, poolId := p.selectPool()
	lastId := pool.available[pool.availableLen-1]
	pool.availableLen--
	wordIndex := int(lastId) / takenBitsPerUint64
	bitIndex := uint(lastId) % takenBitsPerUint64
	pool.taken[wordIndex] |= uint64(1) << bitIndex
	pool.takenLen++
	return &pool.elements[lastId], poolId, lastId
}

func (p *PoolGroup[T]) Remove(poolIndex PoolGroupId, elementId PoolIndex) {
	if poolIndex < 0 || len(p.pools) <= poolIndex {
		return
	}
	pool := p.pools[poolIndex]
	wordIndex := int(elementId) / takenBitsPerUint64
	bitIndex := uint(elementId) % takenBitsPerUint64
	mask := uint64(1) << bitIndex
	// Element isn't currently taken.
	if pool.taken[wordIndex]&mask == 0 {
		return
	}
	pool.taken[wordIndex] &^= mask
	pool.takenLen--
	pool.available[pool.availableLen] = elementId
	pool.availableLen++
}

func (p *PoolGroup[T]) Reserve(additionalElements int) {
	for i := range p.pools {
		additionalElements -= p.pools[i].availableLen
	}
	if additionalElements <= 0 {
		return
	}
	addPools := (additionalElements + ElementsInPool - 1) / ElementsInPool
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
		// Snapshot so elements added during iteration aren't unexpectedly
		// included in the current iteration.
		taken := p.pools[i].taken
		for wordIndex := range taken {
			word := taken[wordIndex]
			for word != 0 {
				bitIndex := bits.TrailingZeros64(word)
				mask := uint64(1) << bitIndex
				// Remove this bit from our local iteration mask.
				word &^= mask
				// It may have been removed by a previous callback.
				if p.pools[i].taken[wordIndex]&mask == 0 {
					continue
				}
				elementIndex := wordIndex*takenBitsPerUint64 + bitIndex
				each(&p.pools[i].elements[elementIndex])
			}
		}
	}
}

func (p *PoolGroup[T]) EachParallel(workName string, workGroup *concurrent.WorkGroup, threads *concurrent.Threads, each func(elm *T)) {
	for i := range p.pools {
		taken := p.pools[i].taken
		for wordIndex := range taken {
			word := taken[wordIndex]
			for word != 0 {
				bitIndex := bits.TrailingZeros64(word)
				word &= word - 1
				elementIndex := wordIndex*takenBitsPerUint64 + bitIndex
				elm := &p.pools[i].elements[elementIndex]
				workGroup.Add(workName, func() {
					each(elm)
				})
			}
		}
	}
	workGroup.Execute(workName, threads)
}

// ConditionalEach iterates over each active element in the pool group, invoking the
// provided callback function `each`. If the callback returns false for any element,
// the iteration stops early. This allows callers to break out of the loop based on
// a condition while still processing elements in order of their allocation.
func (p *PoolGroup[T]) ConditionalEach(each func(elm *T) bool) {
	for i := range p.pools {
		taken := p.pools[i].taken
		for wordIndex := range taken {
			word := taken[wordIndex]
			for word != 0 {
				bitIndex := bits.TrailingZeros64(word)
				mask := uint64(1) << bitIndex
				word &^= mask
				// It may have been removed during an earlier callback.
				if p.pools[i].taken[wordIndex]&mask == 0 {
					continue
				}
				elementIndex := wordIndex*takenBitsPerUint64 + bitIndex
				if !each(&p.pools[i].elements[elementIndex]) {
					return
				}
			}
		}
	}
}
