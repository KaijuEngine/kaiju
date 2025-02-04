package pooling

const (
	ElementsInPool = 0xFF
)

type Pool[T any] struct {
	elements     [ElementsInPool]T
	taken        [ElementsInPool]uint8
	available    [ElementsInPool]uint8
	takenLen     int
	availableLen int
}

func (p *Pool[T]) init() {
	for i := range len(p.available) {
		p.available[i] = uint8(i)
	}
	p.availableLen = ElementsInPool
}
