/******************************************************************************/
/* mesh.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/load_result"
)

type fbxPolygonCorner struct {
	ControlPoint  int
	PolygonVertex int
}

type fbxPolygon struct {
	Corners []fbxPolygonCorner
}

type fbxLayerElementVec2 struct {
	Mapping   string
	Reference string
	Values    []matrix.Vec2
	Indices   []int32
	Valid     bool
}

type fbxLayerElementVec3 struct {
	Mapping   string
	Reference string
	Values    []matrix.Vec3
	Indices   []int32
	Valid     bool
}

type fbxLayerElementColor struct {
	Mapping   string
	Reference string
	Values    []matrix.Color
	Indices   []int32
	Valid     bool
}

func sceneIndexToLoadResult(index SceneIndex) (load_result.Result, error) {
	return sceneIndexToLoadResultWithPath(index, "")
}

func sceneIndexToLoadResultWithPath(index SceneIndex, sourcePath string) (load_result.Result, error) {
	res := load_result.Result{
		TextureBytes: make(map[string][]byte),
	}
	converter := newFBXBasisConverter(index.GlobalSettings)
	unitScale := matrix.Float(index.GlobalSettings.UnitScaleFactor)
	if unitScale == 0 {
		unitScale = 1
	}
	nodeIndexByObjectID := make(map[int64]int, len(index.Model))
	for _, modelID := range sortedObjectIDs(index.Model) {
		model := index.Model[modelID]
		nodeIndexByObjectID[modelID] = len(res.Nodes)
		res.Nodes = append(res.Nodes, loadResultNodeFromObject(model, converter, unitScale))
	}
	for _, modelID := range sortedObjectIDs(index.Model) {
		childIndex := nodeIndexByObjectID[modelID]
		for _, connection := range index.Connections.ParentsByChild[modelID] {
			if connection.Type != "OO" {
				continue
			}
			if parentIndex, ok := nodeIndexByObjectID[connection.Parent]; ok {
				res.Nodes[childIndex].Parent = parentIndex
				break
			}
		}
	}
	bindings := geometryBindings(index)
	skinBindings := fbxSkinBindings(index, nodeIndexByObjectID, converter, unitScale, &res)
	morphTargets := fbxMorphTargets(index, converter, unitScale)
	materials := fbxMaterialResolver{
		index:        index,
		sourcePath:   sourcePath,
		textureBytes: res.TextureBytes,
	}
	for i := range bindings {
		object := bindings[i].nodeObject
		if _, ok := nodeIndexByObjectID[object.ID]; ok {
			continue
		}
		nodeIndexByObjectID[object.ID] = len(res.Nodes)
		res.Nodes = append(res.Nodes, loadResultNodeFromObject(object, converter, unitScale))
	}
	for i := range bindings {
		binding := bindings[i]
		options := fbxMeshOptions{
			Skin:        skinBindings[binding.geometry.ID],
			MorphTarget: morphTargets[binding.geometry.ID],
		}
		verts, indices, err := meshGeometryFromObjectWithOptions(binding.geometry, binding.modelObject, converter, unitScale, options)
		if err != nil {
			return res, err
		}
		nodeIndex := nodeIndexByObjectID[binding.nodeObject.ID]
		if options.Skin == nil && !fbxModelHasAnimation(index, binding.nodeObject.ID) {
			if rotation, ok := fbxModelImportCorrectionRotation(binding.modelObject, converter); ok {
				bakeRotationTransform(verts, rotation)
				res.Nodes[nodeIndex].Rotation = matrix.QuaternionIdentity()
				if fbxIsUnitCorrectionScale(res.Nodes[nodeIndex].Scale) {
					res.Nodes[nodeIndex].Scale = matrix.Vec3One()
				}
			}
		}
		name := res.Nodes[nodeIndex].Name
		meshName := binding.geometry.Name
		if meshName == "" {
			meshName = fmt.Sprintf("Geometry_%d", binding.geometry.ID)
		}
		key := fmt.Sprintf("%s/%d", meshName, binding.geometry.ID)
		res.Add(name, key, verts, indices, materials.TexturesForBinding(binding), &res.Nodes[nodeIndex])
	}
	res.Animations = fbxAnimations(index, nodeIndexByObjectID, converter, unitScale)
	for i := range res.Animations {
		for j := range res.Animations[i].Frames {
			for k := range res.Animations[i].Frames[j].Bones {
				nodeIndex := res.Animations[i].Frames[j].Bones[k].NodeIndex
				if nodeIndex >= 0 && nodeIndex < len(res.Nodes) {
					res.Nodes[nodeIndex].IsAnimated = true
					for parent := res.Nodes[nodeIndex].Parent; parent >= 0; parent = res.Nodes[parent].Parent {
						res.Nodes[parent].IsAnimated = true
					}
				}
			}
		}
	}
	return res, nil
}

type fbxGeometryBinding struct {
	geometry    *Object
	modelObject *Object
	nodeObject  *Object
}

func geometryBindings(index SceneIndex) []fbxGeometryBinding {
	geometryIDs := sortedObjectIDs(index.Geometry)
	bindings := make([]fbxGeometryBinding, 0, len(geometryIDs))
	for _, geometryID := range geometryIDs {
		geometry := index.Geometry[geometryID]
		if childNode(geometry.Node, "Vertices") == nil || childNode(geometry.Node, "PolygonVertexIndex") == nil {
			continue
		}
		var modelObject *Object
		nodeObject := geometry
		for _, connection := range index.Connections.ParentsByChild[geometryID] {
			if model := index.Model[connection.Parent]; connection.Type == "OO" && model != nil {
				modelObject = model
				nodeObject = model
				break
			}
		}
		bindings = append(bindings, fbxGeometryBinding{
			geometry:    geometry,
			modelObject: modelObject,
			nodeObject:  nodeObject,
		})
	}
	return bindings
}

func fbxModelHasAnimation(index SceneIndex, modelID int64) bool {
	for _, connection := range index.Connections.PropertiesByNode[modelID] {
		if connection.Type != "OP" {
			continue
		}
		if fbxObjectKind(index.Animation[connection.Child], "AnimationCurveNode", "AnimCurveNode") {
			return true
		}
	}
	return false
}

func fbxIsUnitCorrectionScale(scale matrix.Vec3) bool {
	return matrix.Vec3ApproxTo(scale, matrix.NewVec3XYZ(100), 0.0001)
}

func sortedObjectIDs(objects map[int64]*Object) []int64 {
	ids := make([]int64, 0, len(objects))
	for id := range objects {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func loadResultNodeFromObject(object *Object, converter fbxBasisConverter, unitScale matrix.Float) load_result.Node {
	node := load_result.Node{
		Id:         int32(object.ID),
		Name:       object.Name,
		Parent:     -1,
		Attributes: make(map[string]any),
	}
	if node.Name == "" {
		node.Name = fmt.Sprintf("%s_%d", object.Class, object.ID)
	}
	node.Position = fbxPropertyVec3(object.Properties, "Lcl Translation", matrix.Vec3Zero())
	node.Position = converter.ConvertPosition(node.Position.Scale(unitScale))
	rotation := fbxPropertyVec3(object.Properties, "Lcl Rotation", matrix.Vec3Zero())
	node.Rotation = converter.ConvertRotation(rotation)
	node.Scale = converter.ConvertScale(fbxPropertyVec3(object.Properties, "Lcl Scaling", matrix.Vec3One()))
	return node
}

func meshGeometryFromObject(geometry *Object) ([]rendering.Vertex, []uint32, error) {
	return meshGeometryFromObjectWithTransforms(geometry, nil, newFBXBasisConverter(DefaultGlobalSettings()), 1)
}

func meshGeometryFromObjectWithTransforms(geometry, model *Object, converter fbxBasisConverter, unitScale matrix.Float) ([]rendering.Vertex, []uint32, error) {
	return meshGeometryFromObjectWithOptions(geometry, model, converter, unitScale, fbxMeshOptions{})
}

func meshGeometryFromObjectWithOptions(geometry, model *Object, converter fbxBasisConverter, unitScale matrix.Float, options fbxMeshOptions) ([]rendering.Vertex, []uint32, error) {
	positions, err := readControlPointPositions(geometry.Node)
	if err != nil {
		return nil, nil, err
	}
	rawPolygonVertexIndex, ok := childInt32Array(geometry.Node, "PolygonVertexIndex")
	if !ok {
		return nil, nil, errors.New("fbx mesh is missing PolygonVertexIndex")
	}
	polygons, err := decodePolygonVertexIndex(rawPolygonVertexIndex)
	if err != nil {
		return nil, nil, err
	}
	normals, err := readLayerElementVec3(geometry.Node, "LayerElementNormal", "Normals", "NormalsIndex")
	if err != nil {
		return nil, nil, err
	}
	uvs, err := readLayerElementVec2(geometry.Node, "LayerElementUV", "UV", "UVIndex")
	if err != nil {
		return nil, nil, err
	}
	colors, err := readLayerElementColor(geometry.Node, "LayerElementColor", "Colors", "ColorIndex")
	if err != nil {
		return nil, nil, err
	}

	cornerCount := 0
	indexCount := 0
	for _, polygon := range polygons {
		cornerCount += len(polygon.Corners)
		indexCount += (len(polygon.Corners) - 2) * 3
	}
	verts := make([]rendering.Vertex, 0, cornerCount)
	indices := make([]uint32, 0, indexCount)
	missingNormals := !normals.Valid
	for _, polygon := range polygons {
		polygonVertexStart := uint32(len(verts))
		for _, corner := range polygon.Corners {
			if corner.ControlPoint < 0 || corner.ControlPoint >= len(positions) {
				return nil, nil, fmt.Errorf("fbx polygon references missing control point %d", corner.ControlPoint)
			}
			position := converter.ConvertPosition(positions[corner.ControlPoint].Scale(unitScale))
			vert := rendering.Vertex{
				Position:     position,
				Tangent:      matrix.Vec4Zero(),
				UV0:          matrix.Vec2Zero(),
				Color:        matrix.ColorWhite(),
				JointIds:     matrix.Vec4i{},
				JointWeights: matrix.Vec4Zero(),
				MorphTarget:  position,
			}
			if options.MorphTarget != nil {
				if target, ok := options.MorphTarget[corner.ControlPoint]; ok {
					vert.MorphTarget = target
				}
			}
			if options.Skin != nil {
				applyFBXSkinning(&vert, options.Skin.Influences[corner.ControlPoint])
			}
			if normal, ok := normals.Value(corner.PolygonVertex, corner.ControlPoint); ok {
				vert.Normal = converter.ConvertDirection(normal)
			}
			if uv, ok := uvs.Value(corner.PolygonVertex, corner.ControlPoint); ok {
				vert.UV0 = convertFBXUV(uv)
			}
			if color, ok := colors.Value(corner.PolygonVertex, corner.ControlPoint); ok {
				vert.Color = color
			}
			verts = append(verts, vert)
		}
		if missingNormals {
			normal := faceNormalForPolygon(verts[polygonVertexStart : polygonVertexStart+uint32(len(polygon.Corners))])
			for i := range polygon.Corners {
				verts[int(polygonVertexStart)+i].Normal = normal
			}
		}
		for _, index := range triangleFanIndices(len(polygon.Corners)) {
			indices = append(indices, polygonVertexStart+index)
		}
	}
	if model != nil {
		bakeGeometricTransform(verts, model, converter, unitScale)
	}
	return verts, indices, nil
}

func convertFBXUV(uv matrix.Vec2) matrix.Vec2 {
	// FBX V coordinates are opposite the engine texture sampling convention.
	return matrix.NewVec2(uv.X(), 1-uv.Y())
}

func readControlPointPositions(node *Node) ([]matrix.Vec3, error) {
	raw, ok := childFloat64Array(node, "Vertices")
	if !ok {
		return nil, errors.New("fbx mesh is missing Vertices")
	}
	if len(raw)%3 != 0 {
		return nil, errors.New("fbx Vertices array length is not divisible by 3")
	}
	positions := make([]matrix.Vec3, len(raw)/3)
	for i := range positions {
		positions[i] = matrix.NewVec3(
			matrix.Float(raw[i*3+0]),
			matrix.Float(raw[i*3+1]),
			matrix.Float(raw[i*3+2]),
		)
	}
	return positions, nil
}

func decodePolygonVertexIndex(raw []int32) ([]fbxPolygon, error) {
	polygons := make([]fbxPolygon, 0)
	current := fbxPolygon{}
	polygonVertex := 0
	for _, value := range raw {
		v := int64(value)
		end := v < 0
		controlPoint := v
		if end {
			controlPoint = -v - 1
		}
		if controlPoint < 0 || controlPoint > int64(int(^uint(0)>>1)) {
			return nil, fmt.Errorf("fbx polygon control point index %d is out of range", controlPoint)
		}
		current.Corners = append(current.Corners, fbxPolygonCorner{
			ControlPoint:  int(controlPoint),
			PolygonVertex: polygonVertex,
		})
		polygonVertex++
		if end {
			if len(current.Corners) < 3 {
				return nil, errors.New("fbx polygon has fewer than 3 vertices")
			}
			polygons = append(polygons, current)
			current = fbxPolygon{}
		}
	}
	if len(current.Corners) > 0 {
		return nil, errors.New("fbx PolygonVertexIndex ended without a polygon-end marker")
	}
	return polygons, nil
}

func triangleFanIndices(cornerCount int) []uint32 {
	if cornerCount < 3 {
		return nil
	}
	indices := make([]uint32, 0, (cornerCount-2)*3)
	for i := 2; i < cornerCount; i++ {
		indices = append(indices, 0, uint32(i-1), uint32(i))
	}
	return indices
}

func faceNormalForPolygon(verts []rendering.Vertex) matrix.Vec3 {
	for i := 2; i < len(verts); i++ {
		normal := rendering.VertexFaceNormal([3]rendering.Vertex{verts[0], verts[i-1], verts[i]})
		if !normal.IsZero() {
			return normal
		}
	}
	return matrix.Vec3Zero()
}

func readLayerElementVec2(node *Node, elementName, valueName, indexName string) (fbxLayerElementVec2, error) {
	layerNode := childNode(node, elementName)
	if layerNode == nil {
		return fbxLayerElementVec2{}, nil
	}
	raw, ok := childFloat64Array(layerNode, valueName)
	if !ok {
		return fbxLayerElementVec2{}, fmt.Errorf("fbx %s is missing %s", elementName, valueName)
	}
	if len(raw)%2 != 0 {
		return fbxLayerElementVec2{}, fmt.Errorf("fbx %s array length is not divisible by 2", valueName)
	}
	values := make([]matrix.Vec2, len(raw)/2)
	for i := range values {
		values[i] = matrix.NewVec2(matrix.Float(raw[i*2+0]), matrix.Float(raw[i*2+1]))
	}
	indices, _ := childInt32Array(layerNode, indexName)
	mapping := layerString(layerNode, "MappingInformationType", "ByPolygonVertex")
	reference := layerString(layerNode, "ReferenceInformationType", "Direct")
	if err := validateLayerElementMode(elementName, mapping, reference); err != nil {
		return fbxLayerElementVec2{}, err
	}
	return fbxLayerElementVec2{
		Mapping:   mapping,
		Reference: reference,
		Values:    values,
		Indices:   indices,
		Valid:     true,
	}, nil
}

func readLayerElementVec3(node *Node, elementName, valueName, indexName string) (fbxLayerElementVec3, error) {
	layerNode := childNode(node, elementName)
	if layerNode == nil {
		return fbxLayerElementVec3{}, nil
	}
	raw, ok := childFloat64Array(layerNode, valueName)
	if !ok {
		return fbxLayerElementVec3{}, fmt.Errorf("fbx %s is missing %s", elementName, valueName)
	}
	if len(raw)%3 != 0 {
		return fbxLayerElementVec3{}, fmt.Errorf("fbx %s array length is not divisible by 3", valueName)
	}
	values := make([]matrix.Vec3, len(raw)/3)
	for i := range values {
		values[i] = matrix.NewVec3(
			matrix.Float(raw[i*3+0]),
			matrix.Float(raw[i*3+1]),
			matrix.Float(raw[i*3+2]),
		)
	}
	indices, _ := childInt32Array(layerNode, indexName)
	mapping := layerString(layerNode, "MappingInformationType", "ByPolygonVertex")
	reference := layerString(layerNode, "ReferenceInformationType", "Direct")
	if err := validateLayerElementMode(elementName, mapping, reference); err != nil {
		return fbxLayerElementVec3{}, err
	}
	return fbxLayerElementVec3{
		Mapping:   mapping,
		Reference: reference,
		Values:    values,
		Indices:   indices,
		Valid:     true,
	}, nil
}

func readLayerElementColor(node *Node, elementName, valueName, indexName string) (fbxLayerElementColor, error) {
	layerNode := childNode(node, elementName)
	if layerNode == nil {
		return fbxLayerElementColor{}, nil
	}
	raw, ok := childFloat64Array(layerNode, valueName)
	if !ok {
		return fbxLayerElementColor{}, fmt.Errorf("fbx %s is missing %s", elementName, valueName)
	}
	if len(raw)%4 != 0 {
		return fbxLayerElementColor{}, fmt.Errorf("fbx %s array length is not divisible by 4", valueName)
	}
	values := make([]matrix.Color, len(raw)/4)
	for i := range values {
		values[i] = matrix.NewColor(
			matrix.Float(raw[i*4+0]),
			matrix.Float(raw[i*4+1]),
			matrix.Float(raw[i*4+2]),
			matrix.Float(raw[i*4+3]),
		)
	}
	indices, _ := childInt32Array(layerNode, indexName)
	mapping := layerString(layerNode, "MappingInformationType", "ByPolygonVertex")
	reference := layerString(layerNode, "ReferenceInformationType", "Direct")
	if err := validateLayerElementMode(elementName, mapping, reference); err != nil {
		return fbxLayerElementColor{}, err
	}
	return fbxLayerElementColor{
		Mapping:   mapping,
		Reference: reference,
		Values:    values,
		Indices:   indices,
		Valid:     true,
	}, nil
}

func validateLayerElementMode(elementName, mapping, reference string) error {
	switch mapping {
	case "ByPolygonVertex", "ByVertice", "ByVertex", "ByControlPoint":
	default:
		slog.Warn("unsupported FBX layer mapping mode",
			"element", elementName,
			"mapping", mapping)
		return fmt.Errorf("fbx %s uses unsupported mapping mode %q", elementName, mapping)
	}
	switch reference {
	case "", "Direct", "IndexToDirect":
	default:
		slog.Warn("unsupported FBX layer reference mode",
			"element", elementName,
			"reference", reference)
		return fmt.Errorf("fbx %s uses unsupported reference mode %q", elementName, reference)
	}
	return nil
}

func (l fbxLayerElementVec2) Value(polygonVertex, controlPoint int) (matrix.Vec2, bool) {
	if !l.Valid {
		return matrix.Vec2Zero(), false
	}
	index, ok := l.directIndex(polygonVertex, controlPoint)
	if !ok || index < 0 || index >= len(l.Values) {
		return matrix.Vec2Zero(), false
	}
	return l.Values[index], true
}

func (l fbxLayerElementVec3) Value(polygonVertex, controlPoint int) (matrix.Vec3, bool) {
	if !l.Valid {
		return matrix.Vec3Zero(), false
	}
	index, ok := l.directIndex(polygonVertex, controlPoint)
	if !ok || index < 0 || index >= len(l.Values) {
		return matrix.Vec3Zero(), false
	}
	return l.Values[index], true
}

func (l fbxLayerElementColor) Value(polygonVertex, controlPoint int) (matrix.Color, bool) {
	if !l.Valid {
		return matrix.ColorWhite(), false
	}
	index, ok := l.directIndex(polygonVertex, controlPoint)
	if !ok || index < 0 || index >= len(l.Values) {
		return matrix.ColorWhite(), false
	}
	return l.Values[index], true
}

func (l fbxLayerElementVec2) directIndex(polygonVertex, controlPoint int) (int, bool) {
	return layerElementDirectIndex(l.Mapping, l.Reference, l.Indices, polygonVertex, controlPoint)
}

func (l fbxLayerElementVec3) directIndex(polygonVertex, controlPoint int) (int, bool) {
	return layerElementDirectIndex(l.Mapping, l.Reference, l.Indices, polygonVertex, controlPoint)
}

func (l fbxLayerElementColor) directIndex(polygonVertex, controlPoint int) (int, bool) {
	return layerElementDirectIndex(l.Mapping, l.Reference, l.Indices, polygonVertex, controlPoint)
}

func layerElementDirectIndex(mapping, reference string, indices []int32, polygonVertex, controlPoint int) (int, bool) {
	elementIndex := 0
	switch mapping {
	case "ByPolygonVertex":
		elementIndex = polygonVertex
	case "ByVertice", "ByVertex", "ByControlPoint":
		elementIndex = controlPoint
	default:
		return 0, false
	}
	switch reference {
	case "", "Direct":
		return elementIndex, true
	case "IndexToDirect":
		if elementIndex < 0 || elementIndex >= len(indices) {
			return 0, false
		}
		return int(indices[elementIndex]), true
	default:
		return 0, false
	}
}

func layerString(node *Node, name, fallback string) string {
	child := childNode(node, name)
	if child == nil || len(child.Properties) == 0 {
		return fallback
	}
	if value, ok := child.Properties[0].Value.(string); ok && value != "" {
		return value
	}
	return fallback
}

func childFloat64Array(node *Node, name string) ([]float64, bool) {
	child := childNode(node, name)
	if child == nil || len(child.Properties) == 0 {
		return nil, false
	}
	switch value := child.Properties[0].Value.(type) {
	case []float64:
		return value, true
	case []float32:
		out := make([]float64, len(value))
		for i := range value {
			out[i] = float64(value[i])
		}
		return out, true
	default:
		return nil, false
	}
}

func childInt32Array(node *Node, name string) ([]int32, bool) {
	child := childNode(node, name)
	if child == nil || len(child.Properties) == 0 {
		return nil, false
	}
	switch value := child.Properties[0].Value.(type) {
	case []int32:
		return value, true
	case []int64:
		out := make([]int32, len(value))
		for i := range value {
			if value[i] < -2147483648 || value[i] > 2147483647 {
				return nil, false
			}
			out[i] = int32(value[i])
		}
		return out, true
	default:
		return nil, false
	}
}
