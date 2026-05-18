/******************************************************************************/
/* occlusion_gpu_test.go                                                       */
/******************************************************************************/

package rendering

import (
	"testing"
	"unsafe"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

func TestGPUOcclusionCandidatePacking(t *testing.T) {
	bounds := graviton.AABBFromMinMax(matrix.Vec3{-1, -2, -3}, matrix.Vec3{4, 5, 6})
	candidate := newGPUOcclusionCandidate(bounds, false)
	if candidate.WorldMinPrevious != (matrix.Vec4{-1, -2, -3, 0}) {
		t.Fatalf("world min/previous = %v", candidate.WorldMinPrevious)
	}
	if candidate.WorldMaxPadding != (matrix.Vec4{4, 5, 6, DefaultOcclusionRectPadPx}) {
		t.Fatalf("world max/padding = %v", candidate.WorldMaxPadding)
	}
	candidate = newGPUOcclusionCandidate(bounds, true)
	if candidate.WorldMinPrevious.W() != 1 {
		t.Fatalf("previous visible flag = %f, want 1", candidate.WorldMinPrevious.W())
	}
	if candidate.DepthParams.X() != DefaultOcclusionDepthBias ||
		candidate.DepthParams.Y() != DefaultOcclusionMinRectPx {
		t.Fatalf("depth params = %v", candidate.DepthParams)
	}
}

func TestOcclusionDispatchGroups(t *testing.T) {
	cases := map[int][3]uint32{
		0:   {0, 1, 1},
		1:   {1, 1, 1},
		64:  {1, 1, 1},
		65:  {2, 1, 1},
		129: {3, 1, 1},
	}
	for count, want := range cases {
		if got := occlusionDispatchGroups(count); got != want {
			t.Fatalf("dispatch groups for %d = %v, want %v", count, got, want)
		}
	}
}

func TestGPUOcclusionApplyResults(t *testing.T) {
	visible := NewShaderDataBase()
	hidden := NewShaderDataBase()
	results := []uint32{1, 0}
	tester := GPUOcclusionTester{}
	frame := &tester.frames[0]
	frame.capacity = len(results)
	frame.candidateCount = len(results)
	frame.resultMapping = unsafe.Pointer(&results[0])
	frame.targets = []*ShaderDataBase{&visible, &hidden}
	tester.applyResults(0)
	if !visible.VisibilityState().LastOcclusionVisible {
		t.Fatalf("visible result was not applied")
	}
	if hidden.VisibilityState().LastOcclusionVisible {
		t.Fatalf("hidden result was not applied")
	}
	if frame.candidateCount != 0 || len(frame.targets) != 0 {
		t.Fatalf("frame was not reset after applying results")
	}
}
