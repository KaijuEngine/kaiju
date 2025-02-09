package pooling

const (
	ElementsInPool = 0xFF
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
	for i := range len(p.available) {
		p.available[i] = PoolIndex(i)
	}
	p.availableLen = ElementsInPool
}
