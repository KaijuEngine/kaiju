//go:build debug

package ui

import "sync/atomic"

type DirtyStats struct {
	SetDirty           uint64
	BubbleSteps        uint64
	CleanScopes        uint64
	CleanNodes         uint64
	StylizerCalls      uint64
	RenderCalls        uint64
	FullCleanFallbacks uint64
	BatchFlushes       uint64
}

var dirtyStats struct {
	setDirty           atomic.Uint64
	bubbleSteps        atomic.Uint64
	cleanScopes        atomic.Uint64
	cleanNodes         atomic.Uint64
	stylizerCalls      atomic.Uint64
	renderCalls        atomic.Uint64
	fullCleanFallbacks atomic.Uint64
	batchFlushes       atomic.Uint64
}

func ResetDirtyStats() {
	dirtyStats.setDirty.Store(0)
	dirtyStats.bubbleSteps.Store(0)
	dirtyStats.cleanScopes.Store(0)
	dirtyStats.cleanNodes.Store(0)
	dirtyStats.stylizerCalls.Store(0)
	dirtyStats.renderCalls.Store(0)
	dirtyStats.fullCleanFallbacks.Store(0)
	dirtyStats.batchFlushes.Store(0)
}

func DirtyStatsSnapshot() DirtyStats {
	return DirtyStats{
		SetDirty:           dirtyStats.setDirty.Load(),
		BubbleSteps:        dirtyStats.bubbleSteps.Load(),
		CleanScopes:        dirtyStats.cleanScopes.Load(),
		CleanNodes:         dirtyStats.cleanNodes.Load(),
		StylizerCalls:      dirtyStats.stylizerCalls.Load(),
		RenderCalls:        dirtyStats.renderCalls.Load(),
		FullCleanFallbacks: dirtyStats.fullCleanFallbacks.Load(),
		BatchFlushes:       dirtyStats.batchFlushes.Load(),
	}
}

func recordDirtySet()               { dirtyStats.setDirty.Add(1) }
func recordDirtyBubbleStep()        { dirtyStats.bubbleSteps.Add(1) }
func recordDirtyCleanScope()        { dirtyStats.cleanScopes.Add(1) }
func recordDirtyCleanNode()         { dirtyStats.cleanNodes.Add(1) }
func recordDirtyStylizerCall()      { dirtyStats.stylizerCalls.Add(1) }
func recordDirtyRenderCall()        { dirtyStats.renderCalls.Add(1) }
func recordDirtyFullCleanFallback() { dirtyStats.fullCleanFallbacks.Add(1) }
func recordDirtyBatchFlush(rootCount int) {
	if rootCount > 0 {
		dirtyStats.batchFlushes.Add(1)
	}
}
