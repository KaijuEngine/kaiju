/******************************************************************************/
/* sweep_and_prune.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"sort"
	"sync"

	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

// Interval represents a projected AABB onto one axis
type Interval struct {
	Min  matrix.Float
	Max  matrix.Float
	Body *RigidBody
	id   int
}

// ActivePair represents a potential collision pair found by SAP
type ActivePair struct {
	BodyA *RigidBody
	BodyB *RigidBody
}

type BroadPhaseFilter func(a, b *RigidBody) bool

type SweepPrune struct {
	// One sorted list per axis
	intervals [3][]Interval
	// Active body proxies cached for the current simulation step.
	proxies []broadPhaseProxy
	// Reusable body collection buffer.
	bodies []*RigidBody
	// Reusable buffer to avoid allocations
	pairBuffer []ActivePair
	// Per-job pair buffers used by the parallel sweep.
	pairBuffers [][]ActivePair
	// Scratch space for estimating the cheapest sweep axis.
	activeMax []matrix.Float
}

type broadPhaseProxy struct {
	body   *RigidBody
	bounds [3]axisBounds
	id     int
}

type axisBounds struct {
	min matrix.Float
	max matrix.Float
}

func (s *SweepPrune) Initialize(initialBodyCount int) {
	for i := range s.intervals {
		s.intervals[i] = make([]Interval, 0, initialBodyCount)
	}
	s.proxies = make([]broadPhaseProxy, 0, initialBodyCount)
	s.bodies = make([]*RigidBody, 0, initialBodyCount)
	s.pairBuffer = make([]ActivePair, 0, initialBodyCount*2)
	s.activeMax = make([]matrix.Float, 0, initialBodyCount)
}

func (s *SweepPrune) Rebuild(bodies *pooling.PoolGroup[RigidBody]) {
	s.RebuildParallel(bodies, nil)
}

func (s *SweepPrune) RebuildParallel(bodies *pooling.PoolGroup[RigidBody], threads *concurrent.Threads) {
	for i := range s.intervals {
		s.intervals[i] = s.intervals[i][:0]
	}
	s.bodies = s.bodies[:0]
	bodies.Each(func(body *RigidBody) {
		if !body.Active {
			return
		}
		s.bodies = append(s.bodies, body)
	})
	count := len(s.bodies)
	if count == 0 {
		s.proxies = s.proxies[:0]
		return
	}
	s.proxies = growSlice(s.proxies, count)
	if workers := broadPhaseWorkerCount(threads, count, 128); workers > 1 {
		runBroadPhaseJobs(threads, workers, count, func(start, end, _ int) {
			for i := start; i < end; i++ {
				s.proxies[i] = newBroadPhaseProxy(s.bodies[i])
			}
		})
	} else {
		for i := range count {
			s.proxies[i] = newBroadPhaseProxy(s.bodies[i])
		}
	}
	for i := range s.proxies {
		p := &s.proxies[i]
		s.intervals[AxisX] = append(s.intervals[AxisX], Interval{
			Min: p.bounds[AxisX].min, Max: p.bounds[AxisX].max, Body: p.body, id: i,
		})
		s.intervals[AxisY] = append(s.intervals[AxisY], Interval{
			Min: p.bounds[AxisY].min, Max: p.bounds[AxisY].max, Body: p.body, id: i,
		})
		s.intervals[AxisZ] = append(s.intervals[AxisZ], Interval{
			Min: p.bounds[AxisZ].min, Max: p.bounds[AxisZ].max, Body: p.body, id: i,
		})
	}
	s.sortIntervals(threads)
}

func newBroadPhaseProxy(body *RigidBody) broadPhaseProxy {
	worldAABB := body.WorldAABB()
	min := worldAABB.Min()
	max := worldAABB.Max()
	return broadPhaseProxy{
		body: body,
		id:   body.poolLocation(),
		bounds: [3]axisBounds{
			AxisX: {min: min.X(), max: max.X()},
			AxisY: {min: min.Y(), max: max.Y()},
			AxisZ: {min: min.Z(), max: max.Z()},
		},
	}
}

func (s *SweepPrune) sortIntervals(threads *concurrent.Threads) {
	sortAxis := func(axis Axis) {
		intervals := s.intervals[axis]
		sort.Slice(intervals, func(a, b int) bool {
			if intervals[a].Min != intervals[b].Min {
				return intervals[a].Min < intervals[b].Min
			}
			if intervals[a].Max != intervals[b].Max {
				return intervals[a].Max < intervals[b].Max
			}
			return s.proxies[intervals[a].id].id < s.proxies[intervals[b].id].id
		})
	}
	if broadPhaseWorkerCount(threads, len(s.proxies), 128) > 1 {
		wg := sync.WaitGroup{}
		wg.Add(3)
		threads.AddWork([]func(int){
			func(int) { defer wg.Done(); sortAxis(AxisX) },
			func(int) { defer wg.Done(); sortAxis(AxisY) },
			func(int) { defer wg.Done(); sortAxis(AxisZ) },
		})
		wg.Wait()
		return
	}
	for i := range s.intervals {
		sortAxis(Axis(i))
	}
}

func (s *SweepPrune) Sweep() []ActivePair {
	return s.SweepParallel(nil, nil)
}

func (s *SweepPrune) SweepParallel(threads *concurrent.Threads, filter BroadPhaseFilter) []ActivePair {
	s.pairBuffer = s.pairBuffer[:0]
	if len(s.proxies) < 2 {
		return s.pairBuffer
	}
	axis := s.bestSweepAxis()
	count := len(s.intervals[axis])
	workers := broadPhaseWorkerCount(threads, count, 96)
	if workers == 1 {
		s.sweepRange(axis, 0, count, &s.pairBuffer, filter)
		return s.pairBuffer
	}
	s.ensurePairBuffers(workers)
	runBroadPhaseJobs(threads, workers, count, func(start, end, worker int) {
		pairs := s.pairBuffers[worker][:0]
		s.sweepRange(axis, start, end, &pairs, filter)
		s.pairBuffers[worker] = pairs
	})
	total := 0
	for i := range workers {
		total += len(s.pairBuffers[i])
	}
	if cap(s.pairBuffer) < total {
		s.pairBuffer = make([]ActivePair, 0, total)
	}
	for i := range workers {
		s.pairBuffer = append(s.pairBuffer, s.pairBuffers[i]...)
	}
	return s.pairBuffer
}

func (s *SweepPrune) bestSweepAxis() Axis {
	bestAxis := AxisX
	bestEstimate := int(^uint(0) >> 1)
	for axis := AxisX; axis <= AxisZ; axis++ {
		estimate := s.estimateAxisPairs(axis)
		if estimate < bestEstimate {
			bestEstimate = estimate
			bestAxis = axis
		}
	}
	return bestAxis
}

func (s *SweepPrune) estimateAxisPairs(axis Axis) int {
	activeMax := s.activeMax[:0]
	estimate := 0
	for _, interval := range s.intervals[axis] {
		valid := activeMax[:0]
		for _, max := range activeMax {
			if max >= interval.Min {
				valid = append(valid, max)
			}
		}
		activeMax = append(valid, interval.Max)
		estimate += len(valid)
	}
	s.activeMax = activeMax[:0]
	return estimate
}

func (s *SweepPrune) sweepRange(axis Axis, start, end int, pairs *[]ActivePair, filter BroadPhaseFilter) {
	intervals := s.intervals[axis]
	for i := start; i < end; i++ {
		aInterval := intervals[i]
		a := &s.proxies[aInterval.id]
		for j := i + 1; j < len(intervals); j++ {
			bInterval := intervals[j]
			if bInterval.Min > aInterval.Max {
				break
			}
			b := &s.proxies[bInterval.id]
			if a.body == b.body || !overlapsOnOtherAxes(a.bounds, b.bounds, axis) {
				continue
			}
			if filter != nil && !filter(a.body, b.body) {
				continue
			}
			*pairs = append(*pairs, ActivePair{BodyA: a.body, BodyB: b.body})
		}
	}
}

func overlapsOnOtherAxes(a, b [3]axisBounds, skip Axis) bool {
	for axis := AxisX; axis <= AxisZ; axis++ {
		if axis == skip {
			continue
		}
		if a[axis].max < b[axis].min || b[axis].max < a[axis].min {
			return false
		}
	}
	return true
}

func (s *SweepPrune) ensurePairBuffers(count int) {
	for len(s.pairBuffers) < count {
		s.pairBuffers = append(s.pairBuffers, make([]ActivePair, 0, 64))
	}
}

func broadPhaseWorkerCount(threads *concurrent.Threads, items, minItemsPerWorker int) int {
	if threads == nil || items < minItemsPerWorker {
		return 1
	}
	workers := threads.ThreadCount()
	if workers <= 1 {
		return 1
	}
	if workers > items {
		workers = items
	}
	if maxWorkers := items / minItemsPerWorker; maxWorkers > 0 && workers > maxWorkers {
		workers = maxWorkers
	}
	if workers < 1 {
		return 1
	}
	return workers
}

func runBroadPhaseJobs(threads *concurrent.Threads, workers, items int, work func(start, end, worker int)) {
	jobs := make([]func(int), workers)
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for worker := range workers {
		start := worker * items / workers
		end := (worker + 1) * items / workers
		worker := worker
		jobs[worker] = func(int) {
			defer wg.Done()
			work(start, end, worker)
		}
	}
	threads.AddWork(jobs)
	wg.Wait()
}

func growSlice[T any](items []T, count int) []T {
	if cap(items) < count {
		return make([]T, count)
	}
	return items[:count]
}
