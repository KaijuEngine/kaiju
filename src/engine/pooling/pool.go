/******************************************************************************/
/* pool.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pooling

const (
	ElementsInPool     = 256
	takenBitsPerUint64 = 64
	takenUint64Count   = (ElementsInPool + takenBitsPerUint64 - 1) / takenBitsPerUint64
)

type PoolIndex = uint8

type Pool[T any] struct {
	elements     [ElementsInPool]T
	taken        [takenUint64Count]uint64
	available    [ElementsInPool]PoolIndex
	takenLen     int
	availableLen int
}

func (p *Pool[T]) init() {
	for i, idx := ElementsInPool-1, 0; i >= 0; i-- {
		p.available[idx] = PoolIndex(i)
		idx++
	}
	p.taken = [takenUint64Count]uint64{}
	p.takenLen = 0
	p.availableLen = ElementsInPool
}
