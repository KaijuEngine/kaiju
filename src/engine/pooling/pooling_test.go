package pooling

import (
	"sync"
	"testing"

	"kaijuengine.com/build"
	"kaijuengine.com/platform/concurrent"
)

// Pooling test coverage rationale (MC/DC):
//
// selectPool      availableLen > 0                  T: TestPoolGroupEach*, TestPoolGroupAdd* (pool has space)
//                                                  F: TestPoolGroupAddSpillsToNewPoolThenReuses (pool full)
//                 build.Debug                       T: -tags=debug run (zdebug.go), F: default run
//                 len(pools) > MaxPoolGroupId       T: TestPoolGroupPoolLimitPanicOnAdd/Reserve (debug only)
//                                                  F: every normal test
// Remove          poolIndex < 0                     T: TestPoolGroupRemoveInvalidPoolIndex (Remove(-1, ...))
//                 len(pools) <= poolIndex           T: TestPoolGroupRemoveInvalidPoolIndex (Remove(1, ...))
//                                                  F: TestPoolGroupRemoveFreesElementForReuse (Remove(0, ...))
//                 taken[wordIndex]&mask == 0        T: TestPoolGroupRemoveNotTakenIsNoOp
//                                                  F: TestPoolGroupRemoveFreesElementForReuse
// Reserve         additionalElements <= 0           T: TestPoolGroupReserveNoOpWhenEnoughAvailable
//                                                  F: TestPoolGroupReserveRoundsUpToWholePools
//                 build.Debug                       T: debug run, F: default run
//                 len(pools) > MaxPoolGroupId       T: TestPoolGroupPoolLimitPanicOnReserve (debug only)
//                                                  F: normal tests
// Each            word != 0                         T: tests with elements / F: TestPoolGroupEachEmpty
//                 taken[wordIndex]&mask == 0        T: TestPoolGroupEachSkipsRemovedDuringIteration
//                                                  F: TestPoolGroupEachVisitsActiveInAllocationOrder
// ConditionalEach word != 0                         T: tests with elements / F: TestPoolGroupConditionalEachEmpty
//                 taken[wordIndex]&mask == 0        T: TestPoolGroupConditionalEachSkipsRemoved
//                                                  F: TestPoolGroupConditionalEachStopsEarly
//                 !each(...)                        T: TestPoolGroupConditionalEachStopsEarly (returns false)
//                                                  F: TestPoolGroupConditionalEachCompletes
// EachParallel    word != 0                         T: TestPoolGroupEachParallelVisitsActive
//                                                  F: TestPoolGroupEachParallelEmpty
// Pool.init       j >= 0 loop                       T/F: any first Add (fills all 256 free slots)
// AtomicPoolGroup locks                            TestAtomicPoolGroupWraps / TestAtomicPoolGroupConcurrent

func TestPoolGroupAddAllocatesElementsInOrder(t *testing.T) {
	var g PoolGroup[int]
	ids := make([]PoolIndex, 0, 5)
	for i := 0; i < 5; i++ {
		elm, pid, id := g.Add()
		if pid != 0 {
			t.Fatalf("expected pool id 0, got %d", pid)
		}
		if int(id) != i {
			t.Fatalf("expected element id %d, got %d", i, id)
		}
		if elm != &g.pools[0].elements[i] {
			t.Fatalf("expected pointer to elements[%d], got %p", i, elm)
		}
		*elm = 100 + i
		ids = append(ids, id)
	}
	if g.Count() != 1 {
		t.Fatalf("expected 1 pool, got %d", g.Count())
	}
	if g.ElementCount() != 5 {
		t.Fatalf("expected 5 active elements, got %d", g.ElementCount())
	}
	// All ids must be distinct.
	seen := make(map[PoolIndex]bool)
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("duplicate id %d", id)
		}
		seen[id] = true
	}
}

func TestPoolGroupAddSpillsToNewPoolThenReuses(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < ElementsInPool; i++ {
		g.Add()
	}
	if g.Count() != 1 {
		t.Fatalf("expected 1 pool after filling, got %d", g.Count())
	}
	// Pool 0 is full: this must allocate pool 1.
	_, pid, id := g.Add()
	if pid != 1 {
		t.Fatalf("expected spill to pool 1, got %d", pid)
	}
	if id != 0 {
		t.Fatalf("expected first id in new pool, got %d", id)
	}
	if g.Count() != 2 {
		t.Fatalf("expected 2 pools, got %d", g.Count())
	}
	if g.ElementCount() != ElementsInPool+1 {
		t.Fatalf("expected %d active elements, got %d", ElementsInPool+1, g.ElementCount())
	}
	// Freeing a slot in pool 0 means the next Add must reuse pool 0.
	g.Remove(0, 0)
	_, pid2, id2 := g.Add()
	if pid2 != 0 {
		t.Fatalf("expected reuse of pool 0, got pool %d", pid2)
	}
	if id2 != 0 {
		t.Fatalf("expected reuse of element id 0, got %d", id2)
	}
}

func TestPoolGroupRemoveFreesElementForReuse(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < 3; i++ {
		e, _, _ := g.Add()
		*e = i
	}
	g.Remove(0, 1)
	if g.ElementCount() != 2 {
		t.Fatalf("expected 2 active elements after removal, got %d", g.ElementCount())
	}
	// Freed slots are reused LIFO.
	elm, pid, id := g.Add()
	if pid != 0 {
		t.Fatalf("expected pool 0, got %d", pid)
	}
	if id != 1 {
		t.Fatalf("expected LIFO reuse of id 1, got %d", id)
	}
	*elm = 42
	// Removing the element again is a real removal of the re-added element.
	g.Remove(0, 1)
	if g.ElementCount() != 2 {
		t.Fatalf("expected 2 active elements after removing re-added element, got %d", g.ElementCount())
	}
	// The element was reused (same storage), so its value was 42 before removal.
	if got := g.pools[0].elements[1]; got != 42 {
		t.Fatalf("expected reused element storage to hold 42, got %d", got)
	}
}

func TestPoolGroupRemoveNotTakenIsNoOp(t *testing.T) {
	var g PoolGroup[int]
	g.Add() // takes id 0 only
	before := g.ElementCount()
	// Element 5 has never been taken in this pool.
	g.Remove(0, 5)
	if g.ElementCount() != before {
		t.Fatalf("expected element count unchanged, got %d", g.ElementCount())
	}
	// Removing a taken element still works afterwards.
	g.Remove(0, 0)
	if g.ElementCount() != 0 {
		t.Fatalf("expected 0 active elements, got %d", g.ElementCount())
	}
	// Removing it again is a no-op.
	g.Remove(0, 0)
	if g.ElementCount() != 0 {
		t.Fatalf("expected 0 active elements after double remove, got %d", g.ElementCount())
	}
}

func TestPoolGroupRemoveInvalidPoolIndexNoOp(t *testing.T) {
	var g PoolGroup[int]
	g.Add()
	// Negative pool index.
	g.Remove(-1, 0)
	// Out of range pool index.
	g.Remove(1, 0)
	// Valid index still works afterwards.
	g.Remove(0, 0)
	if g.ElementCount() != 0 {
		t.Fatalf("expected 0 active elements, got %d", g.ElementCount())
	}
}

func TestPoolGroupElementCountAcrossPools(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < ElementsInPool; i++ {
		g.Add()
	}
	g.Add() // pool 1
	g.Add() // pool 1
	if g.ElementCount() != ElementsInPool+2 {
		t.Fatalf("expected %d active elements, got %d", ElementsInPool+2, g.ElementCount())
	}
}

func TestPoolGroupEachVisitsActiveInAllocationOrder(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < 6; i++ {
		e, _, _ := g.Add()
		*e = i
	}
	var visited []int
	g.Each(func(elm *int) {
		visited = append(visited, *elm)
	})
	expected := []int{0, 1, 2, 3, 4, 5}
	if len(visited) != len(expected) {
		t.Fatalf("expected %d visits, got %d (%v)", len(expected), len(visited), visited)
	}
	for i := range expected {
		if visited[i] != expected[i] {
			t.Fatalf("expected visit order %v, got %v", expected, visited)
		}
	}
}

func TestPoolGroupEachSkipsRemovedDuringIteration(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < 4; i++ {
		e, _, _ := g.Add()
		*e = i
	}
	var visited []int
	g.Each(func(elm *int) {
		if *elm == 0 {
			// Removing an element that has not been visited yet.
			g.Remove(0, 1)
		}
		visited = append(visited, *elm)
	})
	expected := []int{0, 2, 3}
	if len(visited) != len(expected) {
		t.Fatalf("expected %v visits, got %v", expected, visited)
	}
	for i := range expected {
		if visited[i] != expected[i] {
			t.Fatalf("expected visits %v, got %v", expected, visited)
		}
	}
	if g.ElementCount() != 3 {
		t.Fatalf("expected 3 active elements, got %d", g.ElementCount())
	}
}

func TestPoolGroupEachSnapshotIgnoresAddedElements(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < 2; i++ {
		e, _, _ := g.Add()
		*e = i
	}
	var visited []int
	g.Each(func(elm *int) {
		if *elm == 0 {
			// Elements added during iteration must not be visited.
			e, _, _ := g.Add()
			*e = 99
		}
		visited = append(visited, *elm)
	})
	if len(visited) != 2 || visited[0] != 0 || visited[1] != 1 {
		t.Fatalf("expected visits [0 1], got %v", visited)
	}
}

func TestPoolGroupEachEmpty(t *testing.T) {
	var g PoolGroup[int]
	count := 0
	g.Each(func(elm *int) { count++ })
	if count != 0 {
		t.Fatalf("expected no visits, got %d", count)
	}
}

func TestPoolGroupConditionalEachStopsEarly(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < 5; i++ {
		e, _, _ := g.Add()
		*e = i
	}
	var visited []int
	g.ConditionalEach(func(elm *int) bool {
		visited = append(visited, *elm)
		return *elm < 2
	})
	expected := []int{0, 1, 2} // id 2 returns false, iteration stops
	if len(visited) != len(expected) {
		t.Fatalf("expected %v visits, got %v", expected, visited)
	}
	for i := range expected {
		if visited[i] != expected[i] {
			t.Fatalf("expected visits %v, got %v", expected, visited)
		}
	}
}

func TestPoolGroupConditionalEachCompletes(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < 3; i++ {
		e, _, _ := g.Add()
		*e = i
	}
	var visited []int
	g.ConditionalEach(func(elm *int) bool {
		visited = append(visited, *elm)
		return true
	})
	if len(visited) != 3 {
		t.Fatalf("expected all 3 elements visited, got %v", visited)
	}
}

func TestPoolGroupConditionalEachSkipsRemoved(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < 4; i++ {
		e, _, _ := g.Add()
		*e = i
	}
	var visited []int
	g.ConditionalEach(func(elm *int) bool {
		if *elm == 0 {
			g.Remove(0, 1)
		}
		visited = append(visited, *elm)
		return true
	})
	expected := []int{0, 2, 3}
	if len(visited) != len(expected) {
		t.Fatalf("expected %v visits, got %v", expected, visited)
	}
	for i := range expected {
		if visited[i] != expected[i] {
			t.Fatalf("expected visits %v, got %v", expected, visited)
		}
	}
}

func TestPoolGroupConditionalEachEmpty(t *testing.T) {
	var g PoolGroup[int]
	count := 0
	g.ConditionalEach(func(elm *int) bool {
		count++
		return true
	})
	if count != 0 {
		t.Fatalf("expected no visits, got %d", count)
	}
}

func TestPoolGroupAllVisitsActiveAndInactive(t *testing.T) {
	var g PoolGroup[int]
	e, _, _ := g.Add()
	*e = 10
	g.Add()
	count := 0
	g.All(func(elm *int) {
		*elm = 1
		count++
	})
	if count != ElementsInPool {
		t.Fatalf("expected all %d elements visited, got %d", ElementsInPool, count)
	}
	if *e != 1 {
		t.Fatalf("expected inactive elements to be visited too, got %d", *e)
	}
	// Second pool to cover the outer pool loop.
	for i := 0; i < ElementsInPool; i++ {
		g.Add()
	}
	count = 0
	g.All(func(elm *int) { count++ })
	if count != ElementsInPool*2 {
		t.Fatalf("expected %d elements visited, got %d", ElementsInPool*2, count)
	}
}

func TestPoolGroupClearResetsElements(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < 10; i++ {
		e, _, _ := g.Add()
		*e = i
	}
	g.Clear()
	if g.ElementCount() != 0 {
		t.Fatalf("expected 0 active elements after Clear, got %d", g.ElementCount())
	}
	if g.Count() != 1 {
		t.Fatalf("expected pools to be retained, got %d", g.Count())
	}
	visited := 0
	g.Each(func(elm *int) { visited++ })
	if visited != 0 {
		t.Fatalf("expected no active elements after Clear, got %d", visited)
	}
	// A cleared pool hands out ids from the start again.
	_, pid, id := g.Add()
	if pid != 0 || id != 0 {
		t.Fatalf("expected reuse of pool 0 id 0, got pool %d id %d", pid, id)
	}
}

func TestPoolGroupReserveNoOpWhenEnoughAvailable(t *testing.T) {
	var g PoolGroup[int]
	g.Add() // pool 0 has 255 free slots
	before := g.Count()
	g.Reserve(10)
	g.Reserve(0)
	if g.Count() != before {
		t.Fatalf("expected no new pools, got %d", g.Count())
	}
	if g.ElementCount() != 1 {
		t.Fatalf("expected 1 active element, got %d", g.ElementCount())
	}
	// More than available adds exactly the needed extra pools.
	g.Reserve(300)
	if g.Count() != before+1 {
		t.Fatalf("expected %d pools, got %d", before+1, g.Count())
	}
}

func TestPoolGroupReserveRoundsUpToWholePools(t *testing.T) {
	var g PoolGroup[int]
	// 255 requested: exactly one new pool (previous code allocated zero).
	g.Reserve(255)
	if g.Count() != 1 {
		t.Fatalf("expected 1 pool for 255 requested, got %d", g.Count())
	}
	var g2 PoolGroup[int]
	g2.Reserve(513) // ceil(513/256) = 3 pools
	if g2.Count() != 3 {
		t.Fatalf("expected 3 pools for 513 requested, got %d", g2.Count())
	}
	// Filling pool 0 completely then reserving 1 must add one pool.
	var g3 PoolGroup[int]
	for i := 0; i < ElementsInPool; i++ {
		g3.Add()
	}
	g3.Reserve(1)
	if g3.Count() != 2 {
		t.Fatalf("expected 2 pools after reserving 1 with full pool, got %d", g3.Count())
	}
}

func TestPoolGroupEachParallelVisitsActive(t *testing.T) {
	var g PoolGroup[int]
	for i := 0; i < 8; i++ {
		e, _, _ := g.Add()
		*e = i
	}
	g.Remove(0, 3)
	g.Remove(0, 5)

	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()

	var mu sync.Mutex
	visited := make(map[int]int)
	g.EachParallel("test", &workGroup, &threads, func(elm *int) {
		mu.Lock()
		visited[*elm]++
		mu.Unlock()
	})
	if len(visited) != 6 {
		t.Fatalf("expected 6 active elements visited, got %d (%v)", len(visited), visited)
	}
	for i := 0; i < 8; i++ {
		if i == 3 || i == 5 {
			if visited[i] != 0 {
				t.Fatalf("expected removed element %d not visited, got %d", i, visited[i])
			}
			continue
		}
		if visited[i] != 1 {
			t.Fatalf("expected element %d visited once, got %d", i, visited[i])
		}
	}
}

func TestPoolGroupEachParallelEmpty(t *testing.T) {
	var g PoolGroup[int]
	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()
	// No active elements: no work is queued and Execute returns immediately.
	g.EachParallel("empty", &workGroup, &threads, func(elm *int) {
		t.Fatal("callback should not run on empty pool")
	})
}

func TestAtomicPoolGroupWraps(t *testing.T) {
	var g AtomicPoolGroup[int]
	if g.Count() != 0 {
		t.Fatalf("expected 0 pools, got %d", g.Count())
	}
	elm, pid, id := g.Add()
	if pid != 0 || id != 0 {
		t.Fatalf("expected pool 0 id 0, got pool %d id %d", pid, id)
	}
	*elm = 7
	if (*PoolGroup[int])(&g).ElementCount() != 1 {
		t.Fatalf("expected 1 active element, got %d", (*PoolGroup[int])(&g).ElementCount())
	}
	g.Each(func(e *int) {
		if *e != 7 {
			t.Fatalf("expected value 7, got %d", *e)
		}
	})
	g.Remove(pid, id)
	if (*PoolGroup[int])(&g).ElementCount() != 0 {
		t.Fatalf("expected 0 active elements after Remove, got %d", (*PoolGroup[int])(&g).ElementCount())
	}
	g.Remove(0, 0) // not taken -> no-op
	g.Reserve(1)
	if g.Count() != 1 {
		t.Fatalf("expected 1 pool after Reserve(1), got %d", g.Count())
	}
	g.Each(func(e *int) {
		t.Fatal("expected no active elements")
	})
	g.Clear()
	if (*PoolGroup[int])(&g).ElementCount() != 0 {
		t.Fatalf("expected 0 active elements after Clear, got %d", (*PoolGroup[int])(&g).ElementCount())
	}
}

func TestAtomicPoolGroupConcurrentAddRemove(t *testing.T) {
	var g AtomicPoolGroup[int]
	const goroutines = 4
	const perGoroutine = 50
	var wg sync.WaitGroup
	type slot struct {
		pool PoolGroupId
		id   PoolIndex
	}
	for w := 0; w < goroutines; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var mine []slot
			for k := 0; k < perGoroutine; k++ {
				e, p, i := g.Add()
				*e = 1
				mine = append(mine, slot{p, i})
			}
			for _, s := range mine {
				g.Remove(s.pool, s.id)
			}
		}()
	}
	wg.Wait()
	if (*PoolGroup[int])(&g).ElementCount() != 0 {
		t.Fatalf("expected 0 active elements after concurrent add/remove, got %d", (*PoolGroup[int])(&g).ElementCount())
	}
}

// poolGroupAtLimitFull returns a PoolGroup whose pools slice length already
// equals MaxPoolGroupId + 1 so the next append triggers the debug sanity
// panic. Every existing pool shares one full (availableLen == 0) pool.
// Only meaningful in debug builds; skipped otherwise.
func poolGroupAtLimitFull() *PoolGroup[int] {
	full := &Pool[int]{}
	var g PoolGroup[int]
	g.pools = make([]*Pool[int], MaxPoolGroupId+1)
	for i := range g.pools {
		g.pools[i] = full
	}
	return &g
}

func TestPoolGroupPoolLimitPanicOnAdd(t *testing.T) {
	if !build.Debug {
		t.Skip("pool limit guard only active in debug builds")
	}
	if testing.Short() {
		t.Skip("large allocation not run in short mode")
	}
	g := poolGroupAtLimitFull()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when pool limit exceeded")
		}
	}()
	g.Add()
}

func TestPoolGroupPoolLimitPanicOnReserve(t *testing.T) {
	if !build.Debug {
		t.Skip("pool limit guard only active in debug builds")
	}
	if testing.Short() {
		t.Skip("large allocation not run in short mode")
	}
	g := poolGroupAtLimitFull()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when pool limit exceeded")
		}
	}()
	g.Reserve(1)
}
