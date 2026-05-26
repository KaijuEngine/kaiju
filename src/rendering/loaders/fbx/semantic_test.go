/******************************************************************************/
/* semantic_test.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"os"
	"reflect"
	"testing"
)

func TestSplitNameClass(t *testing.T) {
	name, class := SplitNameClass("Suzanne\x00\x01Geometry")
	if name != "Suzanne" || class != "Geometry" {
		t.Fatalf("SplitNameClass = %q, %q; want Suzanne, Geometry", name, class)
	}
	name, class = SplitNameClass("PlainName")
	if name != "PlainName" || class != "" {
		t.Fatalf("SplitNameClass without separator = %q, %q; want PlainName, empty class", name, class)
	}
	name, class = SplitNameClass(string([]byte{'B', 0xff, 0x00, 0x01, 'M', 0xfe}))
	if name != "B\uFFFD" || class != "M\uFFFD" {
		t.Fatalf("SplitNameClass invalid UTF-8 = %q, %q; want replacement runes", name, class)
	}
}

func TestParseProperties70TypedAccessors(t *testing.T) {
	node := Node{
		Children: []Node{{
			Name: "Properties70",
			Children: []Node{
				testProperty70("Number", "double", "", "A", float64(12.5)),
				testProperty70("Vector", "Vector3D", "Vector", "A", float64(1), float64(2), float64(3)),
				testProperty70("Text", "KString", "", "A", "hello"),
				testProperty70("Flag", "bool", "", "A", true),
				testProperty70("Enum", "enum", "", "A", int32(2)),
			},
		}},
	}
	props := ParseProperties70(&node)
	if got, ok := props.Number("Number"); !ok || got != 12.5 {
		t.Fatalf("Number = %v, %v; want 12.5, true", got, ok)
	}
	if got, ok := props.Vec3("Vector"); !ok || !reflect.DeepEqual(got, [3]float64{1, 2, 3}) {
		t.Fatalf("Vector = %#v, %v; want [1 2 3], true", got, ok)
	}
	if got, ok := props.String("Text"); !ok || got != "hello" {
		t.Fatalf("Text = %q, %v; want hello, true", got, ok)
	}
	if got, ok := props.Bool("Flag"); !ok || !got {
		t.Fatalf("Flag = %v, %v; want true, true", got, ok)
	}
	if got, ok := props.Enum("Enum"); !ok || got != 2 {
		t.Fatalf("Enum = %d, %v; want 2, true", got, ok)
	}
}

func TestBuildSceneIndexSyntheticGeometryToModel(t *testing.T) {
	geometryID := int64(100)
	modelID := int64(200)
	textureID := int64(300)
	doc, err := Parse(testFBXFileWithNodes(7400,
		testNode{
			name: "Objects",
			children: []testNode{
				{
					name: "Geometry",
					properties: [][]byte{
						propInt64(geometryID),
						propString("Suzanne\x00\x01Geometry"),
						propString("Mesh"),
					},
				},
				{
					name: "Model",
					properties: [][]byte{
						propInt64(modelID),
						propString("Suzanne\x00\x01Model"),
						propString("Mesh"),
					},
					children: []testNode{{
						name: "Properties70",
						children: []testNode{{
							name: "P",
							properties: [][]byte{
								propString("Lcl Translation"),
								propString("Lcl Translation"),
								propString(""),
								propString("A"),
								propFloat64(1),
								propFloat64(2),
								propFloat64(3),
							},
						}},
					}},
				},
			},
		},
		testNode{
			name: "Connections",
			children: []testNode{{
				name: "C",
				properties: [][]byte{
					propString("OO"),
					propInt64(geometryID),
					propInt64(modelID),
				},
			}, {
				name: "C",
				properties: [][]byte{
					propString("OP"),
					propInt64(textureID),
					propInt64(modelID),
					propString("DiffuseColor"),
				},
			}},
		},
	))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	index, err := BuildSceneIndex(doc)
	if err != nil {
		t.Fatalf("BuildSceneIndex returned error: %v", err)
	}
	if len(index.Geometry) != 1 || len(index.Model) != 1 {
		t.Fatalf("Geometry/Model counts = %d/%d; want 1/1", len(index.Geometry), len(index.Model))
	}
	if index.Geometry[geometryID].Name != "Suzanne" || index.Geometry[geometryID].Class != "Geometry" {
		t.Fatalf("Geometry object = %#v; want Suzanne Geometry", index.Geometry[geometryID])
	}
	if got := index.Connections.ChildrenByParent[modelID]; !testHasConnection(got, "OO", geometryID, modelID, "") {
		t.Fatalf("ChildrenByParent[%d] = %#v; want Geometry child", modelID, got)
	}
	if got := index.Connections.ParentsByChild[geometryID]; len(got) != 1 || got[0].Parent != modelID {
		t.Fatalf("ParentsByChild[%d] = %#v; want Model parent", geometryID, got)
	}
	if got := index.Connections.PropertiesByNode[modelID]; len(got) != 1 || got[0].Property != "DiffuseColor" {
		t.Fatalf("PropertiesByNode[%d] = %#v; want DiffuseColor OP connection", modelID, got)
	}
	if got, ok := index.Model[modelID].Properties.Vec3("Lcl Translation"); !ok || got != [3]float64{1, 2, 3} {
		t.Fatalf("Lcl Translation = %#v, %v; want [1 2 3], true", got, ok)
	}
}

func TestBuildSceneIndexGlobalSettings(t *testing.T) {
	doc, err := Parse(testFBXFileWithNodes(7400,
		testNode{
			name: "GlobalSettings",
			children: []testNode{{
				name: "Properties70",
				children: []testNode{
					testNode{name: "P", properties: [][]byte{propString("UpAxis"), propString("int"), propString("Integer"), propString(""), propInt32(1)}},
					testNode{name: "P", properties: [][]byte{propString("UpAxisSign"), propString("int"), propString("Integer"), propString(""), propInt32(1)}},
					testNode{name: "P", properties: [][]byte{propString("FrontAxis"), propString("int"), propString("Integer"), propString(""), propInt32(2)}},
					testNode{name: "P", properties: [][]byte{propString("FrontAxisSign"), propString("int"), propString("Integer"), propString(""), propInt32(1)}},
					testNode{name: "P", properties: [][]byte{propString("CoordAxis"), propString("int"), propString("Integer"), propString(""), propInt32(0)}},
					testNode{name: "P", properties: [][]byte{propString("CoordAxisSign"), propString("int"), propString("Integer"), propString(""), propInt32(1)}},
					testNode{name: "P", properties: [][]byte{propString("UnitScaleFactor"), propString("double"), propString("Number"), propString(""), propFloat64(1)}},
				},
			}},
		},
	))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	index, err := BuildSceneIndex(doc)
	if err != nil {
		t.Fatalf("BuildSceneIndex returned error: %v", err)
	}
	if !index.GlobalSettings.IsKaijuCompatible() {
		t.Fatalf("GlobalSettings = %#v; want Kaiju-compatible defaults", index.GlobalSettings)
	}
}

func TestBuildSceneIndexDefinitions(t *testing.T) {
	doc, err := Parse(testFBXFileWithNodes(7400,
		testNode{
			name: "Definitions",
			children: []testNode{{
				name:       "ObjectType",
				properties: [][]byte{propString("Model")},
				children: []testNode{
					{name: "Count", properties: [][]byte{propInt32(1)}},
					{
						name:       "PropertyTemplate",
						properties: [][]byte{propString("FbxNode")},
						children: []testNode{{
							name: "Properties70",
							children: []testNode{{
								name: "P",
								properties: [][]byte{
									propString("DefaultAttributeIndex"),
									propString("int"),
									propString("Integer"),
									propString(""),
									propInt32(-1),
								},
							}},
						}},
					},
				},
			}},
		},
	))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	index, err := BuildSceneIndex(doc)
	if err != nil {
		t.Fatalf("BuildSceneIndex returned error: %v", err)
	}
	def, ok := index.Definitions.ByObjectType["Model"]
	if !ok {
		t.Fatal("Model definition was not indexed")
	}
	if def.Count != 1 {
		t.Fatalf("Model definition count = %d, want 1", def.Count)
	}
	if got, ok := def.Properties.Int("DefaultAttributeIndex"); !ok || got != -1 {
		t.Fatalf("DefaultAttributeIndex = %d, %v; want -1, true", got, ok)
	}
}

func TestBuildSceneIndexMonkeyFixture(t *testing.T) {
	data, err := os.ReadFile("../../../editor/editor_embedded_content/editor_content/meshes/monkey.fbx")
	if err != nil {
		t.Skipf("monkey fixture not available: %v", err)
	}
	doc, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(monkey.fbx) returned error: %v", err)
	}
	index, err := BuildSceneIndex(doc)
	if err != nil {
		t.Fatalf("BuildSceneIndex(monkey.fbx) returned error: %v", err)
	}
	if index.Version != 7400 {
		t.Fatalf("Version = %d, want 7400", index.Version)
	}
	if len(index.Geometry) != 1 {
		t.Fatalf("Geometry count = %d, want 1", len(index.Geometry))
	}
	if len(index.Model) != 1 {
		t.Fatalf("Model count = %d, want 1", len(index.Model))
	}
	if len(index.Connections.All) != 2 {
		t.Fatalf("Connection count = %d, want 2", len(index.Connections.All))
	}
}

func testProperty70(name string, typ string, label string, flags string, values ...any) Node {
	node := Node{
		Name: "P",
		Properties: []Property{
			{Value: name},
			{Value: typ},
			{Value: label},
			{Value: flags},
		},
	}
	for _, value := range values {
		node.Properties = append(node.Properties, Property{Value: value})
	}
	return node
}

func testHasConnection(connections []Connection, typ string, child int64, parent int64, property string) bool {
	for _, connection := range connections {
		if connection.Type == typ && connection.Child == child && connection.Parent == parent && connection.Property == property {
			return true
		}
	}
	return false
}
