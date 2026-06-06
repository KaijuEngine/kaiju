/******************************************************************************/
/* project_mesh_upgrade_test.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"
	"testing"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

func TestUpgradeMeshContentToGLBRewritesLegacyContentInPlace(t *testing.T) {
	pfs, err := project_file_system.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pfs.Close() })
	meshConfigDir := filepath.Join(project_file_system.ContentConfigFolder, project_file_system.ContentMeshFolder)
	meshContentDir := filepath.Join(project_file_system.ContentFolder, project_file_system.ContentMeshFolder)
	if err = pfs.MkdirAll(meshConfigDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err = pfs.MkdirAll(meshContentDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	const legacyID = "legacy.obj"
	cfgPath := filepath.Join(meshConfigDir, legacyID+".json")
	if err = content_database.WriteConfig(cfgPath, content_database.ContentConfig{
		Name:    "LegacyMesh",
		Type:    (content_database.Mesh{}).TypeName(),
		SrcName: "LegacyMesh",
	}, &pfs); err != nil {
		t.Fatal(err)
	}
	legacyMesh := kaiju_mesh.KaijuMesh{
		Name: "legacy",
		Verts: []rendering.Vertex{
			{Position: matrix.Vec3{0, 0, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
			{Position: matrix.Vec3{1, 0, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
			{Position: matrix.Vec3{0, 1, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
		},
		Indexes: []uint32{0, 1, 2},
	}
	legacyData := bytes.Buffer{}
	if err = gob.NewEncoder(&legacyData).Encode(legacyMesh); err != nil {
		t.Fatal(err)
	}
	contentPath := filepath.Join(meshContentDir, legacyID)
	if err = pfs.WriteFile(contentPath, legacyData.Bytes(), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	project := Project{
		fileSystem:    pfs,
		cacheDatabase: content_database.New(),
		Settings: Settings{
			EditorVersion: 0.0018,
		},
	}
	if err = project.cacheDatabase.Build(&project.fileSystem); err != nil {
		t.Fatal(err)
	}
	if err = project.upgradeMeshContentToGLB(); err != nil {
		t.Fatal(err)
	}
	if !project.fileSystem.Exists(contentPath) {
		t.Fatal("legacy mesh content path should still exist after upgrade")
	}
	if project.fileSystem.Exists(contentPath + ".glb") {
		t.Fatal("upgrade must not rename legacy mesh content")
	}
	upgraded, err := project.fileSystem.ReadFile(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	if !kaiju_mesh.IsGLB(upgraded) {
		t.Fatal("legacy mesh content was not rewritten as GLB bytes")
	}
	km, err := kaiju_mesh.Deserialize(upgraded)
	if err != nil {
		t.Fatal(err)
	}
	if km.BVH != nil {
		t.Fatal("expected upgraded GLB mesh to omit triangle BVH")
	}
	bvh := km.GenerateBVH(nil, nil, "hit")
	target, _, ok := bvh.RayIntersect(graviton.Ray{
		Origin:    matrix.Vec3{0.25, 0.25, 1},
		Direction: matrix.Vec3{0, 0, -1},
	}, 2)
	if !ok {
		t.Fatal("expected upgraded GLB mesh to generate runtime bounds BVH")
	}
	if target != "hit" {
		t.Fatalf("runtime bounds BVH target = %v, want hit", target)
	}
	if !project.fileSystem.Exists(cfgPath) {
		t.Fatal("mesh config path should not be renamed during upgrade")
	}
}
