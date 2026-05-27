/******************************************************************************/
/* skinning_animation.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"sort"
	"strings"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/load_result"
)

const fbxKTimeTicksPerSecond = 46186158000.0

type fbxMeshOptions struct {
	Skin        *fbxSkinBinding
	MorphTarget map[int]matrix.Vec3
}

type fbxSkinBinding struct {
	Influences map[int][]fbxVertexInfluence
}

type fbxVertexInfluence struct {
	JointIndex int32
	Weight     matrix.Float
}

func fbxSkinBindings(index SceneIndex, nodeIndexByObjectID map[int64]int, converter fbxBasisConverter, unitScale matrix.Float, res *load_result.Result) map[int64]*fbxSkinBinding {
	bindings := make(map[int64]*fbxSkinBinding)
	jointIndexByNodeIndex := make(map[int]int)
	for _, deformerID := range sortedObjectIDs(index.Deformer) {
		skin := index.Deformer[deformerID]
		if !fbxObjectKind(skin, "Skin") {
			continue
		}
		geometryID, ok := fbxConnectedParentOfClass(index, skin.ID, "Geometry")
		if !ok {
			continue
		}
		binding := bindings[geometryID]
		if binding == nil {
			binding = &fbxSkinBinding{Influences: make(map[int][]fbxVertexInfluence)}
			bindings[geometryID] = binding
		}
		for _, connection := range index.Connections.ChildrenByParent[skin.ID] {
			if connection.Type != "OO" {
				continue
			}
			cluster := index.Deformer[connection.Child]
			if cluster == nil || !fbxObjectKind(cluster, "Cluster") {
				continue
			}
			jointNodeIndex, ok := fbxClusterJointNodeIndex(index, cluster.ID, nodeIndexByObjectID)
			if !ok {
				continue
			}
			jointIndex, ok := jointIndexByNodeIndex[jointNodeIndex]
			if !ok {
				jointIndex = len(res.Joints)
				jointIndexByNodeIndex[jointNodeIndex] = jointIndex
				res.Joints = append(res.Joints, load_result.Joint{
					Id:   int32(jointNodeIndex),
					Skin: fbxClusterInverseBind(cluster, converter, unitScale),
				})
			}
			indexes, weights := fbxClusterWeights(cluster)
			for i := 0; i < len(indexes) && i < len(weights); i++ {
				controlPoint := int(indexes[i])
				if controlPoint < 0 {
					continue
				}
				binding.Influences[controlPoint] = append(binding.Influences[controlPoint], fbxVertexInfluence{
					JointIndex: int32(jointIndex),
					Weight:     matrix.Float(weights[i]),
				})
			}
		}
	}
	return bindings
}

func applyFBXSkinning(vertex *rendering.Vertex, influences []fbxVertexInfluence) {
	if len(influences) == 0 {
		return
	}
	sort.SliceStable(influences, func(i, j int) bool {
		return influences[i].Weight > influences[j].Weight
	})
	if len(influences) > 4 {
		influences = influences[:4]
	}
	var sum matrix.Float
	for i := range influences {
		if influences[i].Weight > 0 {
			sum += influences[i].Weight
		}
	}
	if sum <= 0 {
		return
	}
	for i := range influences {
		weight := influences[i].Weight
		if weight < 0 {
			weight = 0
		}
		vertex.JointIds[i] = influences[i].JointIndex
		vertex.JointWeights[i] = weight / sum
	}
}

func fbxClusterWeights(cluster *Object) ([]int32, []float64) {
	indexes, _ := childInt32Array(cluster.Node, "Indexes")
	weights, _ := childFloat64Array(cluster.Node, "Weights")
	return indexes, weights
}

func fbxClusterJointNodeIndex(index SceneIndex, clusterID int64, nodeIndexByObjectID map[int64]int) (int, bool) {
	for _, connection := range index.Connections.ChildrenByParent[clusterID] {
		if connection.Type != "OO" {
			continue
		}
		if _, ok := index.Model[connection.Child]; !ok {
			continue
		}
		nodeIndex, ok := nodeIndexByObjectID[connection.Child]
		return nodeIndex, ok
	}
	return 0, false
}

func fbxClusterInverseBind(cluster *Object, converter fbxBasisConverter, unitScale matrix.Float) matrix.Mat4 {
	transform := fbxChildMat4(cluster.Node, "Transform", converter, unitScale)
	link := fbxChildMat4(cluster.Node, "TransformLink", converter, unitScale)
	link.Inverse()
	return matrix.Mat4Multiply(link, transform)
}

func fbxMorphTargets(index SceneIndex, converter fbxBasisConverter, unitScale matrix.Float) map[int64]map[int]matrix.Vec3 {
	targets := make(map[int64]map[int]matrix.Vec3)
	for _, deformerID := range sortedObjectIDs(index.Deformer) {
		blendShape := index.Deformer[deformerID]
		if !fbxObjectKind(blendShape, "BlendShape") {
			continue
		}
		geometryID, ok := fbxConnectedParentOfClass(index, blendShape.ID, "Geometry")
		if !ok {
			continue
		}
		if targets[geometryID] != nil {
			continue
		}
		for _, channelConnection := range index.Connections.ChildrenByParent[blendShape.ID] {
			if targets[geometryID] != nil {
				break
			}
			if channelConnection.Type != "OO" {
				continue
			}
			channel := index.Deformer[channelConnection.Child]
			if channel == nil || !fbxObjectKind(channel, "BlendShapeChannel") {
				continue
			}
			for _, shapeConnection := range index.Connections.ChildrenByParent[channel.ID] {
				if shapeConnection.Type != "OO" {
					continue
				}
				shape := index.Geometry[shapeConnection.Child]
				if shape == nil {
					continue
				}
				indexes, _ := childInt32Array(shape.Node, "Indexes")
				positions, err := readControlPointPositions(shape.Node)
				if err != nil {
					continue
				}
				if targets[geometryID] == nil {
					targets[geometryID] = make(map[int]matrix.Vec3)
				}
				for i := 0; i < len(indexes) && i < len(positions); i++ {
					controlPoint := int(indexes[i])
					if controlPoint < 0 {
						continue
					}
					targets[geometryID][controlPoint] = converter.ConvertPosition(positions[i].Scale(unitScale))
				}
				break
			}
		}
	}
	return targets
}

type fbxAnimationChannel struct {
	NodeIndex int
	PathType  load_result.AnimationPathType
	Axis      int
	Curve     *Object
}

type fbxAnimationSample struct {
	Values        [3]matrix.Float
	Set           [3]bool
	Interpolation load_result.AnimationInterpolation
}

func fbxAnimations(index SceneIndex, nodeIndexByObjectID map[int64]int, converter fbxBasisConverter, unitScale matrix.Float) []load_result.Animation {
	stackIDs := fbxObjectIDsByKind(index, "AnimationStack", "AnimStack")
	if len(stackIDs) == 0 {
		stackIDs = []int64{0}
	}
	out := make([]load_result.Animation, 0, len(stackIDs))
	for _, stackID := range stackIDs {
		stack := index.Animation[stackID]
		name := "FBX Animation"
		if stack != nil && stack.Name != "" {
			name = stack.Name
		}
		channels := fbxAnimationChannelsForStack(index, stackID, nodeIndexByObjectID)
		anim := fbxBuildAnimation(name, channels, index, nodeIndexByObjectID, converter, unitScale)
		if len(anim.Frames) > 0 {
			out = append(out, anim)
		}
	}
	return out
}

func fbxAnimationChannelsForStack(index SceneIndex, stackID int64, nodeIndexByObjectID map[int64]int) []fbxAnimationChannel {
	layerIDs := fbxAnimationLayerIDs(index, stackID)
	if len(layerIDs) == 0 {
		layerIDs = fbxObjectIDsByKind(index, "AnimationLayer", "AnimLayer")
	}
	layerSet := make(map[int64]bool, len(layerIDs))
	for _, id := range layerIDs {
		layerSet[id] = true
	}
	var channels []fbxAnimationChannel
	for _, curveNodeID := range sortedObjectIDs(index.Animation) {
		curveNode := index.Animation[curveNodeID]
		if !fbxObjectKind(curveNode, "AnimationCurveNode", "AnimCurveNode") {
			continue
		}
		if len(layerSet) > 0 && !fbxCurveNodeInLayers(index, curveNode.ID, layerSet) {
			continue
		}
		nodeIndex, pathType, ok := fbxCurveNodeTarget(index, curveNode.ID, nodeIndexByObjectID)
		if !ok {
			continue
		}
		for _, connection := range index.Connections.ChildrenByParent[curveNode.ID] {
			if connection.Type != "OP" {
				continue
			}
			curve := index.Animation[connection.Child]
			if curve == nil || !fbxObjectKind(curve, "AnimationCurve", "AnimCurve") {
				continue
			}
			axis, ok := fbxAnimationAxis(connection.Property)
			if !ok {
				continue
			}
			channels = append(channels, fbxAnimationChannel{
				NodeIndex: nodeIndex,
				PathType:  pathType,
				Axis:      axis,
				Curve:     curve,
			})
		}
	}
	return channels
}

func fbxBuildAnimation(name string, channels []fbxAnimationChannel, index SceneIndex, nodeIndexByObjectID map[int64]int, converter fbxBasisConverter, unitScale matrix.Float) load_result.Animation {
	type channelKey struct {
		NodeIndex int
		PathType  load_result.AnimationPathType
	}
	grouped := make(map[channelKey]map[float32]*fbxAnimationSample)
	for _, channel := range channels {
		times, values := fbxAnimationCurveKeys(channel.Curve)
		interpolation := fbxAnimationCurveInterpolation(channel.Curve, len(times))
		key := channelKey{NodeIndex: channel.NodeIndex, PathType: channel.PathType}
		if grouped[key] == nil {
			grouped[key] = make(map[float32]*fbxAnimationSample)
		}
		defaults := fbxAnimationDefaultValues(index, nodeIndexByObjectID, channel.NodeIndex, channel.PathType, unitScale)
		for i := 0; i < len(times) && i < len(values); i++ {
			seconds := fbxKTimeToSeconds(times[i])
			sample := grouped[key][seconds]
			if sample == nil {
				sample = &fbxAnimationSample{
					Values:        defaults,
					Interpolation: interpolation,
				}
				grouped[key][seconds] = sample
			}
			value := matrix.Float(values[i])
			if channel.PathType == load_result.AnimPathTranslation {
				value *= unitScale
			}
			sample.Values[channel.Axis] = value
			sample.Set[channel.Axis] = true
			if sample.Interpolation != load_result.AnimInterpolateCubicSpline {
				sample.Interpolation = interpolation
			}
		}
	}
	framesByTime := make(map[float32]*load_result.AnimKeyFrame)
	for key, samples := range grouped {
		for seconds, sample := range samples {
			frame := framesByTime[seconds]
			if frame == nil {
				frame = &load_result.AnimKeyFrame{
					Time:  seconds,
					Bones: make([]load_result.AnimBone, 0),
				}
				framesByTime[seconds] = frame
			}
			bone := load_result.AnimBone{
				NodeIndex:     key.NodeIndex,
				PathType:      key.PathType,
				Interpolation: sample.Interpolation,
			}
			values := matrix.NewVec3(sample.Values[0], sample.Values[1], sample.Values[2])
			switch key.PathType {
			case load_result.AnimPathTranslation:
				bone.Data = converter.ConvertPosition(values).AsAligned16()
			case load_result.AnimPathRotation:
				bone.Data = converter.ConvertRotation(values)
			case load_result.AnimPathScale:
				bone.Data = converter.ConvertScale(values).AsAligned16()
			default:
				continue
			}
			frame.Bones = append(frame.Bones, bone)
		}
	}
	frames := make([]load_result.AnimKeyFrame, 0, len(framesByTime))
	for _, frame := range framesByTime {
		sort.Slice(frame.Bones, func(i, j int) bool {
			if frame.Bones[i].NodeIndex == frame.Bones[j].NodeIndex {
				return frame.Bones[i].PathType < frame.Bones[j].PathType
			}
			return frame.Bones[i].NodeIndex < frame.Bones[j].NodeIndex
		})
		frames = append(frames, *frame)
	}
	sort.Slice(frames, func(i, j int) bool { return frames[i].Time < frames[j].Time })
	for i := 0; i < len(frames)-1; i++ {
		frames[i].Time = frames[i+1].Time - frames[i].Time
	}
	if len(frames) > 0 {
		frames[len(frames)-1].Time = 0
	}
	return load_result.Animation{Name: name, Frames: frames}
}

func fbxAnimationLayerIDs(index SceneIndex, stackID int64) []int64 {
	if stackID == 0 {
		return nil
	}
	var ids []int64
	for _, connection := range index.Connections.ChildrenByParent[stackID] {
		if connection.Type != "OO" {
			continue
		}
		layer := index.Animation[connection.Child]
		if layer != nil && fbxObjectKind(layer, "AnimationLayer", "AnimLayer") {
			ids = append(ids, layer.ID)
		}
	}
	return ids
}

func fbxCurveNodeInLayers(index SceneIndex, curveNodeID int64, layers map[int64]bool) bool {
	for _, connection := range index.Connections.ParentsByChild[curveNodeID] {
		if connection.Type == "OO" && layers[connection.Parent] {
			return true
		}
	}
	return false
}

func fbxCurveNodeTarget(index SceneIndex, curveNodeID int64, nodeIndexByObjectID map[int64]int) (int, load_result.AnimationPathType, bool) {
	for _, connection := range index.Connections.ParentsByChild[curveNodeID] {
		if connection.Type != "OP" {
			continue
		}
		nodeIndex, ok := nodeIndexByObjectID[connection.Parent]
		if !ok {
			continue
		}
		pathType, ok := fbxAnimationPath(connection.Property)
		if ok {
			return nodeIndex, pathType, true
		}
	}
	return 0, load_result.AnimPathInvalid, false
}

func fbxAnimationCurveKeys(curve *Object) ([]int64, []float64) {
	times, _ := childInt64Array(curve.Node, "KeyTime")
	values, _ := childFloat64Array(curve.Node, "KeyValueFloat")
	return times, values
}

func fbxAnimationCurveInterpolation(curve *Object, keyCount int) load_result.AnimationInterpolation {
	flags, _ := childInt32Array(curve.Node, "KeyAttrFlags")
	for _, flag := range flags {
		if flag&8 != 0 {
			if data, ok := childFloat64Array(curve.Node, "KeyAttrDataFloat"); ok && len(data) >= keyCount*4 {
				return load_result.AnimInterpolateCubicSpline
			}
			return load_result.AnimInterpolateLinear
		}
		if flag&2 != 0 {
			return load_result.AnimInterpolateStep
		}
		if flag&4 != 0 {
			return load_result.AnimInterpolateLinear
		}
	}
	return load_result.AnimInterpolateLinear
}

func fbxAnimationDefaultValues(index SceneIndex, nodeIndexByObjectID map[int64]int, nodeIndex int, pathType load_result.AnimationPathType, unitScale matrix.Float) [3]matrix.Float {
	var object *Object
	for objectID, candidate := range index.Model {
		if candidate == nil {
			continue
		}
		if nodeIndexByID, ok := nodeIndexByObjectID[objectID]; ok && nodeIndexByID == nodeIndex {
			object = candidate
			break
		}
	}
	if object == nil {
		switch pathType {
		case load_result.AnimPathScale:
			return [3]matrix.Float{1, 1, 1}
		default:
			return [3]matrix.Float{}
		}
	}
	property := "Lcl Translation"
	if pathType == load_result.AnimPathRotation {
		property = "Lcl Rotation"
	} else if pathType == load_result.AnimPathScale {
		property = "Lcl Scaling"
	}
	fallback := matrix.Vec3Zero()
	if pathType == load_result.AnimPathScale {
		fallback = matrix.Vec3One()
	}
	value := fbxPropertyVec3(object.Properties, property, fallback)
	if pathType == load_result.AnimPathTranslation {
		value = value.Scale(unitScale)
	}
	return [3]matrix.Float{value.X(), value.Y(), value.Z()}
}

func fbxKTimeToSeconds(value int64) float32 {
	return float32(float64(value) / fbxKTimeTicksPerSecond)
}

func fbxAnimationPath(property string) (load_result.AnimationPathType, bool) {
	switch strings.TrimSpace(property) {
	case "Lcl Translation", "T", "Translation":
		return load_result.AnimPathTranslation, true
	case "Lcl Rotation", "R", "Rotation":
		return load_result.AnimPathRotation, true
	case "Lcl Scaling", "S", "Scaling", "Scale":
		return load_result.AnimPathScale, true
	default:
		return load_result.AnimPathInvalid, false
	}
}

func fbxAnimationAxis(property string) (int, bool) {
	property = strings.TrimSpace(property)
	if i := strings.LastIndex(property, "|"); i >= 0 {
		property = property[i+1:]
	}
	switch strings.ToUpper(property) {
	case "X":
		return 0, true
	case "Y":
		return 1, true
	case "Z":
		return 2, true
	default:
		return 0, false
	}
}

func fbxConnectedParentOfClass(index SceneIndex, childID int64, class string) (int64, bool) {
	for _, connection := range index.Connections.ParentsByChild[childID] {
		if connection.Type != "OO" {
			continue
		}
		parent := index.Objects[connection.Parent]
		if parent != nil && (parent.Class == class || parent.NodeClass == class) {
			return parent.ID, true
		}
	}
	return 0, false
}

func fbxObjectKind(object *Object, kinds ...string) bool {
	if object == nil {
		return false
	}
	for _, kind := range kinds {
		if object.Class == kind || object.SubClass == kind || object.NodeClass == kind {
			return true
		}
	}
	return false
}

func fbxObjectIDsByKind(index SceneIndex, kinds ...string) []int64 {
	ids := make([]int64, 0)
	for id, object := range index.Animation {
		if fbxObjectKind(object, kinds...) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func fbxChildMat4(node *Node, name string, converter fbxBasisConverter, unitScale matrix.Float) matrix.Mat4 {
	raw, ok := childFloat64Array(node, name)
	if !ok || len(raw) < 16 {
		return matrix.Mat4Identity()
	}
	values := make([]matrix.Float, 16)
	for i := 0; i < 16; i++ {
		values[i] = matrix.Float(raw[i])
	}
	m := matrix.Mat4FromSlice(values)
	if converter.settings.IsKaijuCompatible() && unitScale == 1 {
		return m
	}
	position := converter.ConvertPosition(m.ExtractPosition().Scale(unitScale))
	scale := converter.ConvertScale(m.ExtractScale())
	rotation := converter.ConvertRotation(m.ExtractRotation().ToEuler())
	out := matrix.Mat4Identity()
	out.Scale(scale)
	out.MultiplyAssign(rotation.ToMat4())
	out.Translate(position)
	return out
}

func childInt64Array(node *Node, name string) ([]int64, bool) {
	child := childNode(node, name)
	if child == nil || len(child.Properties) == 0 {
		return nil, false
	}
	switch value := child.Properties[0].Value.(type) {
	case []int64:
		return value, true
	case []int32:
		out := make([]int64, len(value))
		for i := range value {
			out[i] = int64(value[i])
		}
		return out, true
	default:
		return nil, false
	}
}
