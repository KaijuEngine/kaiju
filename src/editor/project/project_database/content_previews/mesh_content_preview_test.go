/******************************************************************************/
/* mesh_content_preview_test.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_previews

import (
	"testing"

	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
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

func TestContentPreviewerExpandsMultiMeshParentToSubmeshPreviewRefs(t *testing.T) {
	cache := content_database.New()
	cache.IndexCachedContent(content_database.CachedContent{
		Path: project_file_system.ContentPath("database/content/mesh/temple.glb").ToConfigPath().String(),
		Config: content_database.ContentConfig{
			Type: content_database.Mesh{}.TypeName(),
			Mesh: &content_database.MeshConfig{Submeshes: []content_database.MeshSubmeshConfig{
				{Key: "roof"},
				{Key: "wall"},
				{Key: "removed", Missing: true},
			}},
		},
	})
	previewer := ContentPreviewer{ed: previewTestEditor{cache: &cache}}

	got := previewer.expandPreviewIds([]string{"temple.glb"})
	want := []string{
		"temple.glb",
		kaiju_mesh.MeshRefString("temple.glb", "roof"),
		kaiju_mesh.MeshRefString("temple.glb", "wall"),
	}
	if len(got) != len(want) {
		t.Fatalf("expanded ids = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expanded ids = %#v, want %#v", got, want)
		}
	}
}

func absFloat(v matrix.Float) matrix.Float {
	if v < 0 {
		return -v
	}
	return v
}

type previewTestEditor struct {
	cache *content_database.Cache
}

func (e previewTestEditor) Host() *engine.Host { return nil }
func (e previewTestEditor) Events() *editor_events.EditorEvents {
	return nil
}
func (e previewTestEditor) ProjectFileSystem() *project_file_system.FileSystem {
	return nil
}
func (e previewTestEditor) Cache() *content_database.Cache { return e.cache }
