package pooling

type AtomicPoolGroup[T any] PoolGroup[T]

func (p *AtomicPoolGroup[T]) Count() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return len(p.pools)
}

func (p *AtomicPoolGroup[T]) Clear() {
	p.lock.Lock()
	defer p.lock.Unlock()
	(*PoolGroup[T])(p).Clear()
}

func (p *AtomicPoolGroup[T]) Add() (elm *T, poolId PoolGroupId, elmId PoolIndex) {
	p.lock.Lock()
	defer p.lock.Unlock()
	return (*PoolGroup[T])(p).Add()
}

func (p *AtomicPoolGroup[T]) Remove(poolIndex PoolGroupId, elementId PoolIndex) {
	p.lock.Lock()
	defer p.lock.Unlock()
	(*PoolGroup[T])(p).Remove(poolIndex, elementId)
}

func (p *AtomicPoolGroup[T]) Reserve(additionalElements int) {
	p.lock.Lock()
	defer p.lock.Unlock()
	(*PoolGroup[T])(p).Reserve(additionalElements)
}

func (p *AtomicPoolGroup[T]) Each(each func(elm *T)) {
	p.lock.Lock()
	defer p.lock.Unlock()
	(*PoolGroup[T])(p).Each(each)
}
