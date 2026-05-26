/******************************************************************************/
/* mesh_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"os"
	"reflect"
	"testing"

	"kaijuengine.com/matrix"
)

func TestDecodePolygonVertexIndex(t *testing.T) {
	polygons, err := decodePolygonVertexIndex([]int32{
		0, 1, -3,
		0, 1, 2, -4,
		0, 1, 2, 3, -5,
	})
	if err != nil {
		t.Fatalf("decodePolygonVertexIndex returned error: %v", err)
	}
	got := make([][]int, len(polygons))
	for i := range polygons {
		for _, corner := range polygons[i].Corners {
			got[i] = append(got[i], corner.ControlPoint)
		}
	}
	want := [][]int{{0, 1, 2}, {0, 1, 2, 3}, {0, 1, 2, 3, 4}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("decodePolygonVertexIndex = %#v, want %#v", got, want)
	}
	if polygons[2].Corners[4].PolygonVertex != 11 {
		t.Fatalf("last polygon vertex = %d, want 11", polygons[2].Corners[4].PolygonVertex)
	}
}

func TestTriangleFanIndices(t *testing.T) {
	cases := []struct {
		name        string
		cornerCount int
		want        []uint32
	}{
		{name: "triangle", cornerCount: 3, want: []uint32{0, 2, 1}},
		{name: "quad", cornerCount: 4, want: []uint32{0, 2, 1, 0, 3, 2}},
		{name: "ngon", cornerCount: 5, want: []uint32{0, 2, 1, 0, 3, 2, 0, 4, 3}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := triangleFanIndices(c.cornerCount); !reflect.DeepEqual(got, c.want) {
				t.Fatalf("triangleFanIndices(%d) = %#v, want %#v", c.cornerCount, got, c.want)
			}
		})
	}
}

func TestMeshGeometryByPolygonVertexIndexToDirectNormalsAndUVs(t *testing.T) {
	geometry := testMeshGeometryObject(testMeshGeometryNode(
		testNodeWithProperty("Vertices", []float64{
			0, 0, 0,
			1, 0, 0,
			0, 1, 0,
		}),
		testNodeWithProperty("PolygonVertexIndex", []int32{0, 1, -3}),
		testLayerElement("LayerElementNormal", "Normals", "NormalsIndex",
			[]float64{
				0, 0, 1,
				0, 1, 0,
			},
			[]int32{1, 0, 1},
		),
		testLayerElement("LayerElementUV", "UV", "UVIndex",
			[]float64{
				0.25, 0.5,
				0.75, 1,
			},
			[]int32{0, 1, 0},
		),
	))
	verts, indices, err := meshGeometryFromObject(geometry)
	if err != nil {
		t.Fatalf("meshGeometryFromObject returned error: %v", err)
	}
	if !reflect.DeepEqual(indices, []uint32{0, 2, 1}) {
		t.Fatalf("indices = %#v, want [0 2 1]", indices)
	}
	wantNormals := []matrix.Vec3{{0, 1, 0}, {0, 0, 1}, {0, 1, 0}}
	wantUVs := []matrix.Vec2{{0.25, 0.5}, {0.75, 1}, {0.25, 0.5}}
	for i := range verts {
		if verts[i].Normal != wantNormals[i] {
			t.Fatalf("verts[%d].Normal = %#v, want %#v", i, verts[i].Normal, wantNormals[i])
		}
		if verts[i].UV0 != wantUVs[i] {
			t.Fatalf("verts[%d].UV0 = %#v, want %#v", i, verts[i].UV0, wantUVs[i])
		}
	}
}

func TestMeshGeometryMissingNormalsGeneratesFaceNormals(t *testing.T) {
	geometry := testMeshGeometryObject(testMeshGeometryNode(
		testNodeWithProperty("Vertices", []float64{
			0, 0, 0,
			1, 0, 0,
			0, 1, 0,
		}),
		testNodeWithProperty("PolygonVertexIndex", []int32{0, 1, -3}),
	))
	verts, _, err := meshGeometryFromObject(geometry)
	if err != nil {
		t.Fatalf("meshGeometryFromObject returned error: %v", err)
	}
	for i := range verts {
		if verts[i].Normal.IsZero() {
			t.Fatalf("verts[%d].Normal = zero, want generated face normal", i)
		}
	}
}

func TestSceneIndexToLoadResultModelHierarchy(t *testing.T) {
	parent := testFBXObject(10, "Parent", "Model", nil)
	child := testFBXObject(20, "Child", "Model", nil)
	index := testSceneIndex(parent, child)
	index.Connections.ParentsByChild[child.ID] = []Connection{{
		Type:   "OO",
		Child:  child.ID,
		Parent: parent.ID,
	}}
	res, err := sceneIndexToLoadResult(index)
	if err != nil {
		t.Fatalf("sceneIndexToLoadResult returned error: %v", err)
	}
	parentNode := res.NodeByName("Parent")
	childNode := res.NodeByName("Child")
	if parentNode == nil || childNode == nil {
		t.Fatalf("nodes = %#v; want Parent and Child", res.Nodes)
	}
	if childNode.Parent != 0 {
		t.Fatalf("Child parent = %d, want Parent node index 0", childNode.Parent)
	}
	if parentNode.Parent != -1 {
		t.Fatalf("Parent parent = %d, want -1", parentNode.Parent)
	}
}

func TestLoadResultNodeLocalTRSAndUnitScale(t *testing.T) {
	model := testFBXObject(10, "Node", "Model", map[string]matrix.Vec3{
		"Lcl Translation": {1, 2, 3},
		"Lcl Rotation":    {0, 90, 0},
		"Lcl Scaling":     {2, 3, 4},
	})
	index := testSceneIndex(model)
	index.GlobalSettings.UnitScaleFactor = 10
	res, err := sceneIndexToLoadResult(index)
	if err != nil {
		t.Fatalf("sceneIndexToLoadResult returned error: %v", err)
	}
	if len(res.Nodes) != 1 {
		t.Fatalf("node count = %d, want 1", len(res.Nodes))
	}
	node := res.Nodes[0]
	if !matrix.Vec3ApproxTo(node.Position, matrix.Vec3{10, 20, 30}, 0.0001) {
		t.Fatalf("Position = %#v, want {10 20 30}", node.Position)
	}
	if node.Scale != (matrix.Vec3{2, 3, 4}) {
		t.Fatalf("Scale = %#v, want {2 3 4}", node.Scale)
	}
	wantRotation := matrix.QuaternionFromEuler(matrix.Vec3{0, 90, 0})
	if !quaternionApproxTo(node.Rotation, wantRotation, 0.0001) {
		t.Fatalf("Rotation = %#v, want %#v", node.Rotation, wantRotation)
	}
}

func TestGeometricTransformBakesMeshWithoutAffectingChildNode(t *testing.T) {
	parent := testFBXObject(10, "Parent", "Model", map[string]matrix.Vec3{
		"GeometricTranslation": {5, 0, 0},
	})
	child := testFBXObject(20, "Child", "Model", map[string]matrix.Vec3{
		"Lcl Translation": {1, 0, 0},
	})
	geometry := testMeshGeometryObject(testMeshGeometryNode(
		testNodeWithProperty("Vertices", []float64{
			0, 0, 0,
			1, 0, 0,
			0, 1, 0,
		}),
		testNodeWithProperty("PolygonVertexIndex", []int32{0, 1, -3}),
	))
	geometry.ID = 30
	index := testSceneIndex(parent, child, geometry)
	index.Connections.ParentsByChild[child.ID] = []Connection{{
		Type:   "OO",
		Child:  child.ID,
		Parent: parent.ID,
	}}
	index.Connections.ParentsByChild[geometry.ID] = []Connection{{
		Type:   "OO",
		Child:  geometry.ID,
		Parent: parent.ID,
	}}
	res, err := sceneIndexToLoadResult(index)
	if err != nil {
		t.Fatalf("sceneIndexToLoadResult returned error: %v", err)
	}
	if len(res.Meshes) != 1 {
		t.Fatalf("mesh count = %d, want 1", len(res.Meshes))
	}
	if got := res.Meshes[0].Verts[0].Position; got != (matrix.Vec3{5, 0, 0}) {
		t.Fatalf("baked vertex position = %#v, want {5 0 0}", got)
	}
	childNode := res.NodeByName("Child")
	if childNode == nil {
		t.Fatal("Child node missing")
	}
	if childNode.Position != (matrix.Vec3{1, 0, 0}) {
		t.Fatalf("Child position = %#v, want local transform unaffected by geometric bake", childNode.Position)
	}
}

func TestToLoadResultMonkeyFixtureGeometry(t *testing.T) {
	data, err := os.ReadFile("../../../editor/editor_embedded_content/editor_content/meshes/monkey.fbx")
	if err != nil {
		t.Skipf("monkey fixture not available: %v", err)
	}
	doc, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(monkey.fbx) returned error: %v", err)
	}
	res, err := ToLoadResult(doc)
	if err != nil {
		t.Fatalf("ToLoadResult(monkey.fbx) returned error: %v", err)
	}
	if len(res.Meshes) != 1 {
		t.Fatalf("mesh count = %d, want 1", len(res.Meshes))
	}
	if got := len(res.Meshes[0].Verts); got != 1968 {
		t.Fatalf("corner vertex count = %d, want 1968", got)
	}
	if got := len(res.Meshes[0].Indexes); got != 2904 {
		t.Fatalf("index count = %d, want 2904", got)
	}
}

func testMeshGeometryObject(node Node) *Object {
	return &Object{
		ID:    1,
		Name:  "Mesh",
		Class: "Geometry",
		Node:  &node,
	}
}

func testFBXObject(id int64, name string, class string, vec3Props map[string]matrix.Vec3) *Object {
	obj := &Object{
		ID:         id,
		Name:       name,
		Class:      class,
		Properties: PropertyTable{ByName: map[string]Property70{}},
		Node:       &Node{Name: class},
	}
	for propName, value := range vec3Props {
		prop := Property70{
			Name: propName,
			Values: []any{
				float64(value.X()),
				float64(value.Y()),
				float64(value.Z()),
			},
		}
		obj.Properties.List = append(obj.Properties.List, prop)
		obj.Properties.ByName[propName] = prop
	}
	return obj
}

func testSceneIndex(objects ...*Object) SceneIndex {
	index := SceneIndex{
		Objects:  map[int64]*Object{},
		Geometry: map[int64]*Object{},
		Model:    map[int64]*Object{},
		Material: map[int64]*Object{},
		Texture:  map[int64]*Object{},
		Video:    map[int64]*Object{},
		Connections: ConnectionIndex{
			ChildrenByParent: map[int64][]Connection{},
			ParentsByChild:   map[int64][]Connection{},
			PropertiesByNode: map[int64][]Connection{},
		},
		GlobalSettings: DefaultGlobalSettings(),
	}
	for _, obj := range objects {
		index.Objects[obj.ID] = obj
		switch obj.Class {
		case "Geometry":
			index.Geometry[obj.ID] = obj
		case "Model":
			index.Model[obj.ID] = obj
		case "Material":
			index.Material[obj.ID] = obj
		case "Texture":
			index.Texture[obj.ID] = obj
		case "Video":
			index.Video[obj.ID] = obj
		}
	}
	return index
}

func quaternionApproxTo(a, b matrix.Quaternion, delta matrix.Float) bool {
	return matrix.Abs(a.W()-b.W()) < delta &&
		matrix.Abs(a.X()-b.X()) < delta &&
		matrix.Abs(a.Y()-b.Y()) < delta &&
		matrix.Abs(a.Z()-b.Z()) < delta ||
		matrix.Abs(a.W()+b.W()) < delta &&
			matrix.Abs(a.X()+b.X()) < delta &&
			matrix.Abs(a.Y()+b.Y()) < delta &&
			matrix.Abs(a.Z()+b.Z()) < delta
}

func testMeshGeometryNode(children ...Node) Node {
	return Node{
		Name:     "Geometry",
		Children: children,
	}
}

func testNodeWithProperty(name string, value any) Node {
	return Node{
		Name:       name,
		Properties: []Property{{Value: value}},
	}
}

func testLayerElement(name, valueName, indexName string, values []float64, indices []int32) Node {
	return Node{
		Name: name,
		Children: []Node{
			testNodeWithProperty("MappingInformationType", "ByPolygonVertex"),
			testNodeWithProperty("ReferenceInformationType", "IndexToDirect"),
			testNodeWithProperty(valueName, values),
			testNodeWithProperty(indexName, indices),
		},
	}
}
