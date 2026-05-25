//go:build !debug

package ui

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

func ResetDirtyStats()                    {}
func DirtyStatsSnapshot() DirtyStats      { return DirtyStats{} }
func recordDirtySet()                     {}
func recordDirtyBubbleStep()              {}
func recordDirtyCleanScope()              {}
func recordDirtyCleanNode()               {}
func recordDirtyStylizerCall()            {}
func recordDirtyRenderCall()              {}
func recordDirtyFullCleanFallback()       {}
func recordDirtyBatchFlush(rootCount int) {}
