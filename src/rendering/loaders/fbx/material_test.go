/******************************************************************************/
/* material_test.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"bytes"
	"testing"
)

func TestFBXMaterialExternalTexturePathResolution(t *testing.T) {
	geometry, model := testTexturedGeometryAndModel()
	material := testFBXObject(30, "Material", "Material", nil)
	texture := testFBXObject(40, "Diffuse", "Texture", nil)
	texture.Node.Children = append(texture.Node.Children,
		testNodeWithProperty("RelativeFilename", "textures\\diffuse.png"))

	index := testSceneIndex(geometry, model, material, texture)
	connect(index, "OO", geometry.ID, model.ID, "")
	connect(index, "OO", material.ID, model.ID, "")
	connect(index, "OP", texture.ID, material.ID, "DiffuseColor")

	res, err := sceneIndexToLoadResultWithPath(index, "models/ship.fbx")
	if err != nil {
		t.Fatalf("sceneIndexToLoadResult returned error: %v", err)
	}
	if got := res.Meshes[0].Textures["baseColor"]; got != "models/textures/diffuse.png" {
		t.Fatalf("baseColor texture = %q, want models/textures/diffuse.png", got)
	}
}

func TestFBXMaterialEmbeddedVideoContent(t *testing.T) {
	geometry, model := testTexturedGeometryAndModel()
	material := testFBXObject(30, "Material", "Material", nil)
	texture := testFBXObject(40, "MetallicRoughness", "Texture", nil)
	video := testFBXObject(50, "PackedTexture", "Video", nil)
	content := []byte{0xff, 0xd8, 0xff, 0xe0}
	video.Node.Children = append(video.Node.Children,
		testNodeWithProperty("Content", content))

	index := testSceneIndex(geometry, model, material, texture, video)
	connect(index, "OO", geometry.ID, model.ID, "")
	connect(index, "OO", material.ID, model.ID, "")
	connect(index, "OP", texture.ID, material.ID, "Maya|metallic")
	connect(index, "OO", video.ID, texture.ID, "")

	res, err := sceneIndexToLoadResultWithPath(index, "models/ship.fbx")
	if err != nil {
		t.Fatalf("sceneIndexToLoadResult returned error: %v", err)
	}
	key := res.Meshes[0].Textures["metallicRoughness"]
	if key != "embedded_50_metallicRoughness" {
		t.Fatalf("metallicRoughness texture = %q, want embedded_50_metallicRoughness", key)
	}
	if got := res.TextureBytes[key]; !bytes.Equal(got, content) {
		t.Fatalf("TextureBytes[%q] = %#v, want %#v", key, got, content)
	}
}

func TestFBXMaterialModelTakesPrecedenceOverGeometry(t *testing.T) {
	geometry, model := testTexturedGeometryAndModel()
	geoMaterial := testFBXObject(30, "GeoMaterial", "Material", nil)
	modelMaterial := testFBXObject(31, "ModelMaterial", "Material", nil)
	geoTexture := testFBXObject(40, "GeoTexture", "Texture", nil)
	modelTexture := testFBXObject(41, "ModelTexture", "Texture", nil)
	geoTexture.Node.Children = append(geoTexture.Node.Children,
		testNodeWithProperty("RelativeFilename", "geometry.png"))
	modelTexture.Node.Children = append(modelTexture.Node.Children,
		testNodeWithProperty("RelativeFilename", "model.png"))

	index := testSceneIndex(geometry, model, geoMaterial, modelMaterial, geoTexture, modelTexture)
	connect(index, "OO", geometry.ID, model.ID, "")
	connect(index, "OO", geoMaterial.ID, geometry.ID, "")
	connect(index, "OO", modelMaterial.ID, model.ID, "")
	connect(index, "OP", geoTexture.ID, geoMaterial.ID, "DiffuseColor")
	connect(index, "OP", modelTexture.ID, modelMaterial.ID, "DiffuseColor")

	res, err := sceneIndexToLoadResultWithPath(index, "models/ship.fbx")
	if err != nil {
		t.Fatalf("sceneIndexToLoadResult returned error: %v", err)
	}
	if got := res.Meshes[0].Textures["baseColor"]; got != "models/model.png" {
		t.Fatalf("baseColor texture = %q, want model material texture", got)
	}
}

func TestFBXMaterialCurrentDirectoryTexturePathIsIgnored(t *testing.T) {
	geometry, model := testTexturedGeometryAndModel()
	material := testFBXObject(30, "Material", "Material", nil)
	texture := testFBXObject(40, "Diffuse", "Texture", nil)
	texture.Node.Children = append(texture.Node.Children,
		testNodeWithProperty("RelativeFilename", "."))

	index := testSceneIndex(geometry, model, material, texture)
	connect(index, "OO", geometry.ID, model.ID, "")
	connect(index, "OO", material.ID, model.ID, "")
	connect(index, "OP", texture.ID, material.ID, "DiffuseColor")

	res, err := sceneIndexToLoadResultWithPath(index, "models/ship.fbx")
	if err != nil {
		t.Fatalf("sceneIndexToLoadResult returned error: %v", err)
	}
	if got := len(res.Meshes[0].Textures); got != 0 {
		t.Fatalf("texture count = %d, want current-directory texture path ignored", got)
	}
}

func testTexturedGeometryAndModel() (*Object, *Object) {
	geometry := testMeshGeometryObject(testMeshGeometryNode(
		testNodeWithProperty("Vertices", []float64{
			0, 0, 0,
			1, 0, 0,
			0, 1, 0,
		}),
		testNodeWithProperty("PolygonVertexIndex", []int32{0, 1, -3}),
	))
	geometry.ID = 10
	model := testFBXObject(20, "Model", "Model", nil)
	return geometry, model
}

func connect(index SceneIndex, typ string, child int64, parent int64, property string) {
	connection := Connection{
		Type:     typ,
		Child:    child,
		Parent:   parent,
		Property: property,
	}
	index.Connections.All = append(index.Connections.All, connection)
	index.Connections.ChildrenByParent[parent] = append(index.Connections.ChildrenByParent[parent], connection)
	index.Connections.ParentsByChild[child] = append(index.Connections.ParentsByChild[child], connection)
	if typ == "OP" {
		index.Connections.PropertiesByNode[parent] = append(index.Connections.PropertiesByNode[parent], connection)
	}
}
