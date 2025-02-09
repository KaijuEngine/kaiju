package pooling

const (
	ElementsInPool = 256
)

type PoolIndex = uint8

type Pool[T any] struct {
	elements     [ElementsInPool]T
	taken        [ElementsInPool]PoolIndex
	available    [ElementsInPool]PoolIndex
	takenLen     int
	availableLen int
}

func (p *Pool[T]) init() {
	for i, idx := ElementsInPool-1, 0; i >= 0; i-- {
		p.available[idx] = PoolIndex(i)
		idx++
	}
	p.availableLen = ElementsInPool
}
