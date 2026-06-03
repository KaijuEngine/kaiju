/******************************************************************************/
/* mesh_content_preview_test.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_previews

import (
	"testing"

	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

func TestAdjustMeshSetColorAndLocationFramesAllSubmeshesTogether(t *testing.T) {
	cam := cameras.NewStandardCamera(256, 256, 256, 256, matrix.NewVec3(0, 0, 5))
	cam.SetLookAt(matrix.Vec3Zero())
	set := kaiju_mesh.KaijuMeshSet{
		Meshes: []kaiju_mesh.KaijuMesh{
			{Verts: []rendering.Vertex{{Position: matrix.NewVec3(-10, 0, 0)}}},
			{Verts: []rendering.Vertex{{Position: matrix.NewVec3(10, 0, 0)}}},
		},
	}

	adjustMeshSetColorAndLocation(cam, &set)

	left := set.Meshes[0].Verts[0]
	right := set.Meshes[1].Verts[0]
	if left.Position.X() >= right.Position.X() {
		t.Fatalf("expected submesh relative positions to be preserved, got left=%v right=%v", left.Position, right.Position)
	}
	midpoint := left.Position.Add(right.Position).Scale(0.5)
	expectedCenter := cam.Position()
	expectedCenter.AddAssign(cam.Forward().Scale(right.Position.Subtract(left.Position).Length() * 1.35))
	if absFloat(midpoint.X()-expectedCenter.X()) > 0.001 ||
		absFloat(midpoint.Y()-expectedCenter.Y()) > 0.001 ||
		absFloat(midpoint.Z()-expectedCenter.Z()) > 0.001 {
		t.Fatalf("preview midpoint = %v, want %v", midpoint, expectedCenter)
	}
	if left.Color != matrix.ColorSlateGrey() || right.Color != matrix.ColorSlateGrey() {
		t.Fatalf("expected preview colors to be slate grey, got left=%v right=%v", left.Color, right.Color)
	}
}

func absFloat(v matrix.Float) matrix.Float {
	if v < 0 {
		return -v
	}
	return v
}
