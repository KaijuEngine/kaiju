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
	candidate := newGPUOcclusionCandidate(bounds, false, DefaultOcclusionTuning())
	if candidate.WorldMinPrevious != (matrix.Vec4{-1, -2, -3, 0}) {
		t.Fatalf("world min/previous = %v", candidate.WorldMinPrevious)
	}
	if candidate.WorldMaxPadding != (matrix.Vec4{4, 5, 6, DefaultOcclusionRectPadPx}) {
		t.Fatalf("world max/padding = %v", candidate.WorldMaxPadding)
	}
	candidate = newGPUOcclusionCandidate(bounds, true, DefaultOcclusionTuning())
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
	frame.resultsPending = true
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
	if frame.resultsPending {
		t.Fatalf("frame still has pending results after applying")
	}
}

func TestGPUOcclusionApplyResultsWithoutPendingResultsFailsOpen(t *testing.T) {
	hidden := NewShaderDataBase()
	hidden.VisibilityState().LastOcclusionVisible = false
	tester := GPUOcclusionTester{}
	frame := &tester.frames[0]
	frame.candidateCount = 1
	frame.targets = []*ShaderDataBase{&hidden}
	tester.applyResults(0)
	if !hidden.VisibilityState().LastOcclusionVisible {
		t.Fatalf("unsubmitted occlusion work should fail open")
	}
	if frame.candidateCount != 0 || len(frame.targets) != 0 || frame.resultsPending {
		t.Fatalf("frame was not reset after fail-open apply")
	}
}

func TestOcclusionCandidateBatchRequiresDepthSource(t *testing.T) {
	painter := GPUPainter{}
	painter.SetOcclusionRuntimeMode(OcclusionRuntimeConservative)
	camera := testOcclusionCameraContainer().Camera
	pass := &RenderPass{occlusionDepthCopyFrom: -1}
	painter.BeginOcclusionCandidateBatch(pass, camera)
	if painter.hasActiveOcclusionCandidateBatch() {
		t.Fatalf("pass without occlusion depth source should not open a candidate batch")
	}

	pass.occlusionDepthCopyFrom = 0
	pass.occlusionDepthCopy.RenderId.Image.handle = unsafe.Pointer(uintptr(1))
	painter.BeginOcclusionCandidateBatch(pass, camera)
	if !painter.hasActiveOcclusionCandidateBatch() {
		t.Fatalf("valid occlusion depth source should open a candidate batch")
	}
	container := testOcclusionCameraContainer()
	container.ChangeCamera(camera)
	if !painter.occlusionCandidateBatchAllows(&container) {
		t.Fatalf("matching camera container should be allowed into the active batch")
	}
	other := testOcclusionCameraContainer()
	if painter.occlusionCandidateBatchAllows(&other) {
		t.Fatalf("different camera container should not be allowed into the active batch")
	}
	painter.EndOcclusionCandidateBatch(pass, camera)
	if painter.hasActiveOcclusionCandidateBatch() {
		t.Fatalf("candidate batch should close after matching pass/camera flush")
	}
}
