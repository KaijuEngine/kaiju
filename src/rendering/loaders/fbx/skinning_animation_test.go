/******************************************************************************/
/* skinning_animation_test.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/load_result"
)

func TestApplyFBXSkinningCapsFourInfluencesAndNormalizes(t *testing.T) {
	vertex := rendering.Vertex{}
	applyFBXSkinning(&vertex, []fbxVertexInfluence{
		{JointIndex: 1, Weight: 0.10},
		{JointIndex: 2, Weight: 0.40},
		{JointIndex: 3, Weight: 0.20},
		{JointIndex: 4, Weight: 0.30},
		{JointIndex: 5, Weight: 0.80},
	})
	if vertex.JointIds != (matrix.Vec4i{5, 2, 4, 3}) {
		t.Fatalf("JointIds = %#v, want top four sorted by weight", vertex.JointIds)
	}
	wantWeights := matrix.Vec4{
		matrix.Float(0.80 / 1.70),
		matrix.Float(0.40 / 1.70),
		matrix.Float(0.30 / 1.70),
		matrix.Float(0.20 / 1.70),
	}
	if !matrix.Vec4ApproxTo(vertex.JointWeights, wantWeights, 0.0001) {
		t.Fatalf("JointWeights = %#v, want %#v", vertex.JointWeights, wantWeights)
	}
}

func TestFBXClusterInverseBindFromTransformData(t *testing.T) {
	meshBind := matrix.Mat4Identity()
	meshBind.Translate(matrix.Vec3{10, 0, 0})
	jointBind := matrix.Mat4Identity()
	jointBind.Translate(matrix.Vec3{2, 0, 0})
	cluster := testFBXObjectWithChildren(30, "Cluster", "Deformer", "Cluster",
		testNodeWithProperty("Transform", mat4Float64s(meshBind)),
		testNodeWithProperty("TransformLink", mat4Float64s(jointBind)),
	)
	got := fbxClusterInverseBind(cluster, newFBXBasisConverter(DefaultGlobalSettings()), 1)
	expected := jointBind
	expected.Inverse()
	expected = matrix.Mat4Multiply(expected, meshBind)
	if !matrix.Mat4ApproxTo(got, expected, 0.0001) {
		t.Fatalf("inverse bind = %#v, want %#v", got, expected)
	}
}

func TestFBXAnimationKTimeConversionAndKeyMerging(t *testing.T) {
	model := testFBXObject(10, "Bone", "Model", nil)
	stack := testAnimationObject(100, "Take", "AnimationStack", "")
	layer := testAnimationObject(110, "BaseLayer", "AnimationLayer", "")
	curveNode := testAnimationObject(120, "T", "AnimationCurveNode", "")
	curveX := testAnimationCurveObject(130, "X", []int64{0, int64(fbxKTimeTicksPerSecond)}, []float64{1, 3}, []int32{4})
	curveY := testAnimationCurveObject(140, "Y", []int64{0, int64(fbxKTimeTicksPerSecond)}, []float64{2, 4}, []int32{4})
	index := testSceneIndex(model, stack, layer, curveNode, curveX, curveY)
	addTestConnection(&index, "OO", layer.ID, stack.ID, "")
	addTestConnection(&index, "OO", curveNode.ID, layer.ID, "")
	addTestConnection(&index, "OP", curveNode.ID, model.ID, "Lcl Translation")
	addTestConnection(&index, "OP", curveX.ID, curveNode.ID, "d|X")
	addTestConnection(&index, "OP", curveY.ID, curveNode.ID, "d|Y")

	anims := fbxAnimations(index, map[int64]int{model.ID: 0}, newFBXBasisConverter(DefaultGlobalSettings()), 1)
	if len(anims) != 1 {
		t.Fatalf("animation count = %d, want 1", len(anims))
	}
	if len(anims[0].Frames) != 2 {
		t.Fatalf("frame count = %d, want 2", len(anims[0].Frames))
	}
	if !matrix.Approx(anims[0].Frames[0].Time, 1) || !matrix.Approx(anims[0].Frames[1].Time, 0) {
		t.Fatalf("relative frame times = %v, %v; want 1, 0", anims[0].Frames[0].Time, anims[0].Frames[1].Time)
	}
	for i, want := range []matrix.Vec3{{1, 2, 0}, {3, 4, 0}} {
		if len(anims[0].Frames[i].Bones) != 1 {
			t.Fatalf("frame %d bone count = %d, want merged translation bone", i, len(anims[0].Frames[i].Bones))
		}
		got := matrix.Vec3FromSlice(anims[0].Frames[i].Bones[0].Data[:])
		if !matrix.Vec3ApproxTo(got, want, 0.0001) {
			t.Fatalf("frame %d translation = %#v, want %#v", i, got, want)
		}
	}
}

func TestFBXAnimationChannelMappingToTRS(t *testing.T) {
	model := testFBXObject(10, "Bone", "Model", map[string]matrix.Vec3{
		"Lcl Scaling": {1, 1, 1},
	})
	stack := testAnimationObject(100, "Take", "AnimationStack", "")
	layer := testAnimationObject(110, "BaseLayer", "AnimationLayer", "")
	tNode := testAnimationObject(120, "T", "AnimationCurveNode", "")
	rNode := testAnimationObject(130, "R", "AnimationCurveNode", "")
	sNode := testAnimationObject(140, "S", "AnimationCurveNode", "")
	tCurve := testAnimationCurveObject(150, "TX", []int64{0}, []float64{5}, []int32{2})
	rCurve := testAnimationCurveObject(160, "RX", []int64{0}, []float64{90}, []int32{4})
	sCurve := testAnimationCurveObject(170, "SX", []int64{0}, []float64{2}, []int32{4})
	index := testSceneIndex(model, stack, layer, tNode, rNode, sNode, tCurve, rCurve, sCurve)
	addTestConnection(&index, "OO", layer.ID, stack.ID, "")
	addAnimCurveNode(&index, tNode.ID, layer.ID, model.ID, "Lcl Translation", tCurve.ID, "d|X")
	addAnimCurveNode(&index, rNode.ID, layer.ID, model.ID, "Lcl Rotation", rCurve.ID, "d|X")
	addAnimCurveNode(&index, sNode.ID, layer.ID, model.ID, "Lcl Scaling", sCurve.ID, "d|X")

	anims := fbxAnimations(index, map[int64]int{model.ID: 0}, newFBXBasisConverter(DefaultGlobalSettings()), 1)
	if len(anims) != 1 || len(anims[0].Frames) != 1 {
		t.Fatalf("got animations %#v, want one single-frame animation", anims)
	}
	bones := anims[0].Frames[0].Bones
	if len(bones) != 3 {
		t.Fatalf("bone count = %d, want translation, rotation, and scale", len(bones))
	}
	byPath := map[load_result.AnimationPathType]load_result.AnimBone{}
	for _, bone := range bones {
		byPath[bone.PathType] = bone
	}
	translation := byPath[load_result.AnimPathTranslation].Data
	if got := matrix.Vec3FromSlice(translation[:]); !matrix.Vec3ApproxTo(got, matrix.Vec3{5, 0, 0}, 0.0001) {
		t.Fatalf("translation = %#v, want {5 0 0}", got)
	}
	wantRotation := matrix.QuaternionFromEuler(matrix.Vec3{90, 0, 0})
	if !quaternionApproxTo(matrix.Quaternion(byPath[load_result.AnimPathRotation].Data), wantRotation, 0.0001) {
		t.Fatalf("rotation = %#v, want %#v", byPath[load_result.AnimPathRotation].Data, wantRotation)
	}
	scale := byPath[load_result.AnimPathScale].Data
	if got := matrix.Vec3FromSlice(scale[:]); !matrix.Vec3ApproxTo(got, matrix.Vec3{2, 1, 1}, 0.0001) {
		t.Fatalf("scale = %#v, want {2 1 1}", got)
	}
	if byPath[load_result.AnimPathTranslation].Interpolation != load_result.AnimInterpolateStep {
		t.Fatalf("constant interpolation was not mapped to step")
	}
}

func TestToLoadResultGeneratedBinarySkinnedFBX(t *testing.T) {
	doc, err := Parse(testFBXFileWithNodes(7400,
		testNode{
			name: "Objects",
			children: []testNode{
				{
					name: "Geometry",
					properties: [][]byte{
						propInt64(100),
						propString("TriangleMesh\x00\x01Geometry"),
						propString("Mesh"),
					},
					children: []testNode{
						{name: "Vertices", properties: [][]byte{propArrayZlib('d', floats64Bytes(
							0, 0, 0,
							1, 0, 0,
							0, 1, 0,
						))}},
						{name: "PolygonVertexIndex", properties: [][]byte{propArrayZlib('i', int32sBytes(0, 1, -3))}},
					},
				},
				{
					name: "Model",
					properties: [][]byte{
						propInt64(200),
						propString("TriangleNode\x00\x01Model"),
						propString("Mesh"),
					},
				},
				{
					name: "Model",
					properties: [][]byte{
						propInt64(300),
						propString("Bone\x00\x01Model"),
						propString("LimbNode"),
					},
				},
				{
					name: "Deformer",
					properties: [][]byte{
						propInt64(400),
						propString("Skin\x00\x01Deformer"),
						propString("Skin"),
					},
				},
				{
					name: "Deformer",
					properties: [][]byte{
						propInt64(500),
						propString("Cluster\x00\x01Deformer"),
						propString("Cluster"),
					},
					children: []testNode{
						{name: "Indexes", properties: [][]byte{propArrayRaw('i', int32sBytes(0, 1, 2))}},
						{name: "Weights", properties: [][]byte{propArrayRaw('d', floats64Bytes(1, 1, 1))}},
						{name: "Transform", properties: [][]byte{propArrayRaw('d', floats64Bytes(mat4Float64s(matrix.Mat4Identity())...))}},
						{name: "TransformLink", properties: [][]byte{propArrayRaw('d', floats64Bytes(mat4Float64s(matrix.Mat4Identity())...))}},
					},
				},
				{
					name: "AnimationStack",
					properties: [][]byte{
						propInt64(600),
						propString("Take 001\x00\x01AnimationStack"),
						propString(""),
					},
				},
				{
					name: "AnimationLayer",
					properties: [][]byte{
						propInt64(610),
						propString("BaseLayer\x00\x01AnimationLayer"),
						propString(""),
					},
				},
				{
					name: "AnimationCurveNode",
					properties: [][]byte{
						propInt64(620),
						propString("T\x00\x01AnimationCurveNode"),
						propString(""),
					},
				},
				{
					name: "AnimationCurve",
					properties: [][]byte{
						propInt64(630),
						propString("TX\x00\x01AnimationCurve"),
						propString(""),
					},
					children: []testNode{
						{name: "KeyTime", properties: [][]byte{propArrayRaw('l', int64sBytes(0, int64(fbxKTimeTicksPerSecond)))}},
						{name: "KeyValueFloat", properties: [][]byte{propArrayRaw('f', floats32Bytes(0, 1))}},
						{name: "KeyAttrFlags", properties: [][]byte{propArrayRaw('i', int32sBytes(4))}},
					},
				},
			},
		},
		testNode{
			name: "Connections",
			children: []testNode{
				testConnectionNode("OO", 100, 200, ""),
				testConnectionNode("OO", 400, 100, ""),
				testConnectionNode("OO", 500, 400, ""),
				testConnectionNode("OO", 300, 500, ""),
				testConnectionNode("OO", 610, 600, ""),
				testConnectionNode("OO", 620, 610, ""),
				testConnectionNode("OP", 620, 300, "Lcl Translation"),
				testConnectionNode("OP", 630, 620, "d|X"),
			},
		},
	))
	if err != nil {
		t.Fatalf("Parse generated skinned FBX returned error: %v", err)
	}
	res, err := ToLoadResult(doc)
	if err != nil {
		t.Fatalf("ToLoadResult generated skinned FBX returned error: %v", err)
	}
	if len(res.Meshes) != 1 || len(res.Joints) != 1 || len(res.Animations) != 1 {
		t.Fatalf("counts meshes/joints/anims = %d/%d/%d, want 1/1/1", len(res.Meshes), len(res.Joints), len(res.Animations))
	}
	if res.Joints[0].Id != 1 {
		t.Fatalf("joint id = %d, want Bone node index 1", res.Joints[0].Id)
	}
	for i, vertex := range res.Meshes[0].Verts {
		if vertex.JointIds.X() != 0 || !matrix.Approx(vertex.JointWeights.X(), 1) {
			t.Fatalf("vertex %d skinning = ids %#v weights %#v, want joint 0 weight 1", i, vertex.JointIds, vertex.JointWeights)
		}
	}
	if len(res.Animations[0].Frames) != 2 || len(res.Animations[0].Frames[1].Bones) != 1 {
		t.Fatalf("animation frames = %#v, want two translated keyframes", res.Animations[0].Frames)
	}
	bone := res.Animations[0].Frames[1].Bones[0]
	data := bone.Data
	if bone.NodeIndex != 1 || bone.PathType != load_result.AnimPathTranslation ||
		!matrix.Vec3ApproxTo(matrix.Vec3FromSlice(data[:]), matrix.Vec3{1, 0, 0}, 0.0001) {
		t.Fatalf("animation bone = %#v, want Bone translation to {1 0 0}", bone)
	}
}

func mat4Float64s(m matrix.Mat4) []float64 {
	out := make([]float64, len(m))
	for i := range m {
		out[i] = float64(m[i])
	}
	return out
}

func testFBXObjectWithChildren(id int64, name string, class string, subClass string, children ...Node) *Object {
	obj := testFBXObject(id, name, class, nil)
	obj.SubClass = subClass
	obj.Node = &Node{Name: class, Children: children}
	return obj
}

func testAnimationObject(id int64, name string, class string, subClass string) *Object {
	return &Object{
		ID:         id,
		Name:       name,
		Class:      class,
		SubClass:   subClass,
		NodeClass:  class,
		Properties: PropertyTable{ByName: map[string]Property70{}},
		Node:       &Node{Name: class},
	}
}

func testAnimationCurveObject(id int64, name string, times []int64, values []float64, flags []int32) *Object {
	return &Object{
		ID:         id,
		Name:       name,
		Class:      "AnimationCurve",
		NodeClass:  "AnimationCurve",
		Properties: PropertyTable{ByName: map[string]Property70{}},
		Node: &Node{
			Name: "AnimationCurve",
			Children: []Node{
				testNodeWithProperty("KeyTime", times),
				testNodeWithProperty("KeyValueFloat", values),
				testNodeWithProperty("KeyAttrFlags", flags),
			},
		},
	}
}

func addAnimCurveNode(index *SceneIndex, curveNodeID, layerID, modelID int64, path string, curveID int64, axis string) {
	addTestConnection(index, "OO", curveNodeID, layerID, "")
	addTestConnection(index, "OP", curveNodeID, modelID, path)
	addTestConnection(index, "OP", curveID, curveNodeID, axis)
}

func addTestConnection(index *SceneIndex, typ string, child, parent int64, property string) {
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

func testConnectionNode(typ string, child, parent int64, property string) testNode {
	props := [][]byte{propString(typ), propInt64(child), propInt64(parent)}
	if property != "" {
		props = append(props, propString(property))
	}
	return testNode{name: "C", properties: props}
}
