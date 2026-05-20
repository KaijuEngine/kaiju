/******************************************************************************/
/* occlusion_gpu_test.go                                                       */
/******************************************************************************/

package rendering

import (
	"math"
	"testing"
	"unsafe"

	"kaijuengine.com/engine/cameras"
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

func TestGPUOcclusionReferenceDoesNotOccludeOffsetSphere(t *testing.T) {
	cameraPosition := matrix.Vec3{-12.218892, 8.048108, -3.305541}
	camera := cameras.NewStandardCamera(1280, 720, 1280, 720, cameraPosition)
	camera.SetPositionAndLookAt(cameraPosition, matrix.Vec3Zero())
	camera.SetProperties(60, 0.01, 250, 1280, 720)

	sphereBounds := graviton.AABBFromWidth(matrix.Vec3{-0.029, 3.127, -0.17}, 1)
	cubeBounds := graviton.AABBFromMinMax(
		matrix.Vec3{-14.458 * 0.5, 4.095 - 10.464*0.5, -2.883 - 0.657*0.5},
		matrix.Vec3{14.458 * 0.5, 4.095 + 10.464*0.5, -2.883 + 0.657*0.5},
	)
	tuning := DefaultOcclusionTuning()
	visible := testGPUOcclusionReferenceVisible(sphereBounds, camera, tuning,
		func(uv matrix.Vec2) matrix.Float {
			depth, ok := testDepthFromAABBAtUV(cubeBounds, camera, 60, uv)
			if !ok {
				return tuning.MissingFar
			}
			return depth
		})
	if !visible {
		t.Fatalf("offset sphere should not be occluded by the cube")
	}
}

func TestGPUOcclusionReferenceOccludesCoveredSphere(t *testing.T) {
	cameraPosition := matrix.Vec3{0, 4.233, -9.681}
	camera := cameras.NewStandardCamera(1280, 720, 1280, 720, cameraPosition)
	camera.SetPositionAndLookAt(cameraPosition, matrix.Vec3{0, 4.233, 0})
	camera.SetProperties(60, 0.01, 500, 1280, 720)

	sphereBounds := graviton.AABBFromWidth(matrix.Vec3{-0.029, 3.127, -0.17}, 0.5)
	cubeBounds := graviton.AABBFromMinMax(
		matrix.Vec3{-14.458 * 0.5, 4.095 - 10.464*0.5, -2.883 - 0.657*0.5},
		matrix.Vec3{14.458 * 0.5, 4.095 + 10.464*0.5, -2.883 + 0.657*0.5},
	)
	tuning := DefaultOcclusionTuning()
	visible := testGPUOcclusionReferenceVisible(sphereBounds, camera, tuning,
		func(uv matrix.Vec2) matrix.Float {
			depth, ok := testDepthFromAABBAtUV(cubeBounds, camera, 60, uv)
			if !ok {
				return tuning.MissingFar
			}
			return depth
		})
	if visible {
		t.Fatalf("covered sphere should be occluded by the cube")
	}
}

func testGPUOcclusionReferenceVisible(bounds graviton.AABB, camera cameras.Camera, tuning OcclusionTuning, depthAt func(matrix.Vec2) matrix.Float) bool {
	rectMin := matrix.Vec2{1, 1}
	rectMax := matrix.Vec2{0, 0}
	nearestDepth := matrix.Float(1)
	for _, corner := range bounds.Corners() {
		screen, ok := testProjectOcclusionPoint(corner, camera)
		if !ok || screen.Z() < 0 || screen.Z() > 1 {
			return true
		}
		uv := matrix.Vec2{screen.X(), screen.Y()}
		rectMin = matrix.Vec2Min(rectMin, uv)
		rectMax = matrix.Vec2Max(rectMax, uv)
		nearestDepth = min(nearestDepth, screen.Z())
	}
	if rectMax.X() <= 0 || rectMax.Y() <= 0 || rectMin.X() >= 1 || rectMin.Y() >= 1 {
		return true
	}
	padUv := matrix.Vec2{
		tuning.RectPadPx / max(matrix.Float(camera.Width()), 1),
		tuning.RectPadPx / max(matrix.Float(camera.Height()), 1),
	}
	rectMin = matrix.Vec2Max(rectMin.Subtract(padUv), matrix.Vec2Zero())
	rectMax = matrix.Vec2Min(rectMax.Add(padUv), matrix.Vec2One())
	center := rectMin.Add(rectMax).Scale(0.5)
	sampledDepth := matrix.Float(0)
	for _, uv := range []matrix.Vec2{
		center,
		rectMin,
		{rectMax.X(), rectMin.Y()},
		{rectMin.X(), rectMax.Y()},
		rectMax,
	} {
		sampledDepth = max(sampledDepth, depthAt(uv))
	}
	if sampledDepth >= tuning.MissingFar || sampledDepth <= 0 ||
		math.IsNaN(float64(sampledDepth)) || math.IsInf(float64(sampledDepth), 0) {
		return true
	}
	return nearestDepth <= sampledDepth+tuning.DepthBias
}

func testDepthFromAABBAtUV(bounds graviton.AABB, camera cameras.Camera, fov float32, uv matrix.Vec2) (matrix.Float, bool) {
	aspect := matrix.Float(camera.Width() / camera.Height())
	tanHalfFOV := matrix.Tan(matrix.Deg2Rad(fov) * 0.5)
	ndcX := uv.X()*2 - 1
	ndcY := uv.Y()*2 - 1
	direction := camera.Forward().Normal().
		Add(camera.Right().Normal().Scale(ndcX * aspect * tanHalfFOV)).
		Subtract(camera.Up().Normal().Scale(ndcY * tanHalfFOV)).
		Normal()
	hit, ok := bounds.RayHit(graviton.Ray{Origin: camera.Position(), Direction: direction})
	if !ok {
		return 0, false
	}
	screen, ok := testProjectOcclusionPoint(hit, camera)
	if !ok || screen.Z() < 0 || screen.Z() > 1 {
		return 0, false
	}
	return screen.Z(), true
}

func testProjectOcclusionPoint(point matrix.Vec3, camera cameras.Camera) (matrix.Vec3, bool) {
	clip := matrix.Mat4MultiplyVec4(camera.Projection(),
		matrix.Mat4MultiplyVec4(camera.View(), point.AsVec4()))
	if testInvalidVec4(clip) || clip.W() <= 0.0001 {
		return matrix.Vec3{}, false
	}
	ndc := matrix.Vec3{clip.X(), clip.Y(), clip.Z()}.Scale(1.0 / clip.W())
	if ndc.IsNaN() || ndc.IsInf(0) {
		return matrix.Vec3{}, false
	}
	return matrix.Vec3{ndc.X()*0.5 + 0.5, ndc.Y()*0.5 + 0.5, ndc.Z()}, true
}

func testInvalidVec4(v matrix.Vec4) bool {
	return math.IsNaN(float64(v.X())) || math.IsNaN(float64(v.Y())) ||
		math.IsNaN(float64(v.Z())) || math.IsNaN(float64(v.W())) ||
		math.IsInf(float64(v.X()), 0) || math.IsInf(float64(v.Y()), 0) ||
		math.IsInf(float64(v.Z()), 0) || math.IsInf(float64(v.W()), 0)
}
