/******************************************************************************/
/* semantic.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"kaijuengine.com/rendering/loaders/load_result"
)

const nameClassSeparator = "\x00\x01"

type SceneIndex struct {
	Version        uint32
	Objects        map[int64]*Object
	ByClass        map[string]map[int64]*Object
	Geometry       map[int64]*Object
	Model          map[int64]*Object
	Material       map[int64]*Object
	Texture        map[int64]*Object
	Video          map[int64]*Object
	Deformer       map[int64]*Object
	Animation      map[int64]*Object
	Connections    ConnectionIndex
	Definitions    DefinitionsIndex
	GlobalSettings GlobalSettings
}

type Object struct {
	ID         int64
	Name       string
	Class      string
	SubClass   string
	NodeClass  string
	Properties PropertyTable
	Node       *Node
}

type Connection struct {
	Type     string
	Child    int64
	Parent   int64
	Property string
	Node     *Node
}

type ConnectionIndex struct {
	All              []Connection
	ChildrenByParent map[int64][]Connection
	ParentsByChild   map[int64][]Connection
	PropertiesByNode map[int64][]Connection
}

type Definition struct {
	ObjectType string
	Count      int64
	Properties PropertyTable
	Node       *Node
}

type DefinitionsIndex struct {
	ByObjectType map[string]*Definition
}

type GlobalSettings struct {
	UpAxis                  int
	UpAxisSign              int
	FrontAxis               int
	FrontAxisSign           int
	CoordAxis               int
	CoordAxisSign           int
	UnitScaleFactor         float64
	OriginalUnitScaleFactor float64
}

func DefaultGlobalSettings() GlobalSettings {
	return GlobalSettings{
		UpAxis:                  1,
		UpAxisSign:              1,
		FrontAxis:               2,
		FrontAxisSign:           1,
		CoordAxis:               0,
		CoordAxisSign:           1,
		UnitScaleFactor:         1,
		OriginalUnitScaleFactor: 1,
	}
}

func (s GlobalSettings) IsKaijuCompatible() bool {
	return s.UpAxis == 1 && s.UpAxisSign == 1 &&
		s.FrontAxis == 2 && s.FrontAxisSign == 1 &&
		s.CoordAxis == 0 && s.CoordAxisSign == 1 &&
		s.UnitScaleFactor == 1
}

type PropertyTable struct {
	ByName map[string]Property70
	List   []Property70
}

type Property70 struct {
	Name   string
	Type   string
	Label  string
	Flags  string
	Values []any
	Node   *Node
}

func BuildSceneIndex(doc Document) (SceneIndex, error) {
	index := SceneIndex{
		Version:   doc.Version,
		Objects:   make(map[int64]*Object),
		ByClass:   make(map[string]map[int64]*Object),
		Geometry:  make(map[int64]*Object),
		Model:     make(map[int64]*Object),
		Material:  make(map[int64]*Object),
		Texture:   make(map[int64]*Object),
		Video:     make(map[int64]*Object),
		Deformer:  make(map[int64]*Object),
		Animation: make(map[int64]*Object),
		Connections: ConnectionIndex{
			ChildrenByParent: make(map[int64][]Connection),
			ParentsByChild:   make(map[int64][]Connection),
			PropertiesByNode: make(map[int64][]Connection),
		},
		Definitions: DefinitionsIndex{
			ByObjectType: make(map[string]*Definition),
		},
		GlobalSettings: DefaultGlobalSettings(),
	}
	if objects := topNode(doc.Nodes, "Objects"); objects != nil {
		if err := index.indexObjects(objects); err != nil {
			return index, err
		}
	}
	if connections := topNode(doc.Nodes, "Connections"); connections != nil {
		if err := index.indexConnections(connections); err != nil {
			return index, err
		}
	}
	if definitions := topNode(doc.Nodes, "Definitions"); definitions != nil {
		index.indexDefinitions(definitions)
	}
	if globalSettings := topNode(doc.Nodes, "GlobalSettings"); globalSettings != nil {
		index.GlobalSettings = indexGlobalSettings(globalSettings)
	}
	return index, nil
}

func SplitNameClass(value string) (string, string) {
	name, class, ok := strings.Cut(value, nameClassSeparator)
	name = strings.ToValidUTF8(name, "\uFFFD")
	if !ok {
		return name, ""
	}
	return name, strings.ToValidUTF8(class, "\uFFFD")
}

func (i *SceneIndex) indexObjects(objects *Node) error {
	for n := range objects.Children {
		node := &objects.Children[n]
		if len(node.Properties) == 0 {
			continue
		}
		id, ok := asInt64(node.Properties[0].Value)
		if !ok {
			return fmt.Errorf("fbx object %q has non-numeric id", node.Name)
		}
		name := ""
		class := node.Name
		if len(node.Properties) > 1 {
			if rawName, ok := node.Properties[1].Value.(string); ok {
				name, class = SplitNameClass(rawName)
				if class == "" {
					class = node.Name
				}
			}
		}
		subClass := ""
		if len(node.Properties) > 2 {
			if v, ok := node.Properties[2].Value.(string); ok {
				subClass = strings.ToValidUTF8(v, "\uFFFD")
			}
		}
		obj := &Object{
			ID:         id,
			Name:       name,
			Class:      class,
			SubClass:   subClass,
			NodeClass:  node.Name,
			Properties: ParseProperties70(node),
			Node:       node,
		}
		i.Objects[id] = obj
		if i.ByClass[obj.Class] == nil {
			i.ByClass[obj.Class] = make(map[int64]*Object)
		}
		i.ByClass[obj.Class][id] = obj
		switch obj.Class {
		case "Geometry":
			i.Geometry[id] = obj
		case "Model":
			i.Model[id] = obj
		case "Material":
			i.Material[id] = obj
		case "Texture":
			i.Texture[id] = obj
		case "Video":
			i.Video[id] = obj
		case "Deformer":
			i.Deformer[id] = obj
		default:
			if isAnimationClass(obj.Class) || isAnimationClass(obj.NodeClass) {
				i.Animation[id] = obj
			}
		}
	}
	return nil
}

func (i *SceneIndex) indexConnections(connections *Node) error {
	for n := range connections.Children {
		node := &connections.Children[n]
		if node.Name != "C" || len(node.Properties) < 3 {
			continue
		}
		connectionType, ok := node.Properties[0].Value.(string)
		if !ok {
			return errors.New("fbx connection type is not a string")
		}
		if connectionType != "OO" && connectionType != "OP" {
			continue
		}
		child, ok := asInt64(node.Properties[1].Value)
		if !ok {
			return fmt.Errorf("fbx %s connection child id is not numeric", connectionType)
		}
		parent, ok := asInt64(node.Properties[2].Value)
		if !ok {
			return fmt.Errorf("fbx %s connection parent id is not numeric", connectionType)
		}
		connection := Connection{
			Type:   connectionType,
			Child:  child,
			Parent: parent,
			Node:   node,
		}
		if len(node.Properties) > 3 {
			connection.Property, _ = node.Properties[3].Value.(string)
			connection.Property = strings.ToValidUTF8(connection.Property, "\uFFFD")
		}
		i.Connections.All = append(i.Connections.All, connection)
		i.Connections.ChildrenByParent[parent] = append(i.Connections.ChildrenByParent[parent], connection)
		i.Connections.ParentsByChild[child] = append(i.Connections.ParentsByChild[child], connection)
		if connection.Type == "OP" {
			i.Connections.PropertiesByNode[parent] = append(i.Connections.PropertiesByNode[parent], connection)
		}
	}
	return nil
}

func (i *SceneIndex) indexDefinitions(definitions *Node) {
	for n := range definitions.Children {
		node := &definitions.Children[n]
		if node.Name != "ObjectType" || len(node.Properties) == 0 {
			continue
		}
		objectType, ok := node.Properties[0].Value.(string)
		if !ok {
			continue
		}
		properties := ParseProperties70(node)
		if template := childNode(node, "PropertyTemplate"); template != nil {
			properties = ParseProperties70(template)
		}
		def := &Definition{
			ObjectType: strings.ToValidUTF8(objectType, "\uFFFD"),
			Properties: properties,
			Node:       node,
		}
		if count := childNode(node, "Count"); count != nil && len(count.Properties) > 0 {
			def.Count, _ = asInt64(count.Properties[0].Value)
		}
		i.Definitions.ByObjectType[def.ObjectType] = def
	}
}

func indexGlobalSettings(node *Node) GlobalSettings {
	settings := DefaultGlobalSettings()
	props := ParseProperties70(node)
	if v, ok := props.Int("UpAxis"); ok {
		settings.UpAxis = int(v)
	}
	if v, ok := props.Int("UpAxisSign"); ok {
		settings.UpAxisSign = int(v)
	}
	if v, ok := props.Int("FrontAxis"); ok {
		settings.FrontAxis = int(v)
	}
	if v, ok := props.Int("FrontAxisSign"); ok {
		settings.FrontAxisSign = int(v)
	}
	if v, ok := props.Int("CoordAxis"); ok {
		settings.CoordAxis = int(v)
	}
	if v, ok := props.Int("CoordAxisSign"); ok {
		settings.CoordAxisSign = int(v)
	}
	if v, ok := props.Number("UnitScaleFactor"); ok {
		settings.UnitScaleFactor = v
	}
	if v, ok := props.Number("OriginalUnitScaleFactor"); ok {
		settings.OriginalUnitScaleFactor = v
	}
	return settings
}

func ParseProperties70(node *Node) PropertyTable {
	table := PropertyTable{ByName: make(map[string]Property70)}
	propsNode := childNode(node, "Properties70")
	if propsNode == nil {
		return table
	}
	for p := range propsNode.Children {
		node := &propsNode.Children[p]
		if node.Name != "P" || len(node.Properties) == 0 {
			continue
		}
		property := Property70{Node: node}
		if v, ok := node.Properties[0].Value.(string); ok {
			property.Name = strings.ToValidUTF8(v, "\uFFFD")
		}
		if len(node.Properties) > 1 {
			property.Type, _ = node.Properties[1].Value.(string)
			property.Type = strings.ToValidUTF8(property.Type, "\uFFFD")
		}
		if len(node.Properties) > 2 {
			property.Label, _ = node.Properties[2].Value.(string)
			property.Label = strings.ToValidUTF8(property.Label, "\uFFFD")
		}
		if len(node.Properties) > 3 {
			property.Flags, _ = node.Properties[3].Value.(string)
			property.Flags = strings.ToValidUTF8(property.Flags, "\uFFFD")
		}
		if len(node.Properties) > 4 {
			property.Values = make([]any, len(node.Properties)-4)
			for i := 4; i < len(node.Properties); i++ {
				property.Values[i-4] = node.Properties[i].Value
			}
		}
		table.List = append(table.List, property)
		table.ByName[property.Name] = property
	}
	return table
}

func (t PropertyTable) Get(name string) (Property70, bool) {
	p, ok := t.ByName[name]
	return p, ok
}

func (t PropertyTable) Number(name string) (float64, bool) {
	if p, ok := t.Get(name); ok {
		return p.Number()
	}
	return 0, false
}

func (t PropertyTable) Int(name string) (int64, bool) {
	if p, ok := t.Get(name); ok {
		return p.Int()
	}
	return 0, false
}

func (t PropertyTable) Vec2(name string) ([2]float64, bool) {
	if p, ok := t.Get(name); ok {
		return p.Vec2()
	}
	return [2]float64{}, false
}

func (t PropertyTable) Vec3(name string) ([3]float64, bool) {
	if p, ok := t.Get(name); ok {
		return p.Vec3()
	}
	return [3]float64{}, false
}

func (t PropertyTable) Vec4(name string) ([4]float64, bool) {
	if p, ok := t.Get(name); ok {
		return p.Vec4()
	}
	return [4]float64{}, false
}

func (t PropertyTable) String(name string) (string, bool) {
	if p, ok := t.Get(name); ok {
		return p.String()
	}
	return "", false
}

func (t PropertyTable) Bool(name string) (bool, bool) {
	if p, ok := t.Get(name); ok {
		return p.Bool()
	}
	return false, false
}

func (t PropertyTable) Enum(name string) (int64, bool) {
	if p, ok := t.Get(name); ok {
		return p.Enum()
	}
	return 0, false
}

func (p Property70) Number() (float64, bool) {
	if len(p.Values) == 0 {
		return 0, false
	}
	return asFloat64(p.Values[0])
}

func (p Property70) Int() (int64, bool) {
	if len(p.Values) == 0 {
		return 0, false
	}
	return asInt64(p.Values[0])
}

func (p Property70) Vec2() ([2]float64, bool) {
	var out [2]float64
	if len(p.Values) < len(out) {
		return out, false
	}
	for i := range out {
		v, ok := asFloat64(p.Values[i])
		if !ok {
			return out, false
		}
		out[i] = v
	}
	return out, true
}

func (p Property70) Vec3() ([3]float64, bool) {
	var out [3]float64
	if len(p.Values) < len(out) {
		return out, false
	}
	for i := range out {
		v, ok := asFloat64(p.Values[i])
		if !ok {
			return out, false
		}
		out[i] = v
	}
	return out, true
}

func (p Property70) Vec4() ([4]float64, bool) {
	var out [4]float64
	if len(p.Values) < len(out) {
		return out, false
	}
	for i := range out {
		v, ok := asFloat64(p.Values[i])
		if !ok {
			return out, false
		}
		out[i] = v
	}
	return out, true
}

func (p Property70) String() (string, bool) {
	if len(p.Values) == 0 {
		return "", false
	}
	v, ok := p.Values[0].(string)
	if !ok {
		return "", false
	}
	return strings.ToValidUTF8(v, "\uFFFD"), true
}

func (p Property70) Bool() (bool, bool) {
	if len(p.Values) == 0 {
		return false, false
	}
	switch v := p.Values[0].(type) {
	case bool:
		return v, true
	case int16:
		return v != 0, true
	case int32:
		return v != 0, true
	case int64:
		return v != 0, true
	case float32:
		return v != 0, true
	case float64:
		return v != 0, true
	default:
		return false, false
	}
}

func (p Property70) Enum() (int64, bool) {
	return p.Int()
}

func topNode(nodes []Node, name string) *Node {
	for n := range nodes {
		if nodes[n].Name == name {
			return &nodes[n]
		}
	}
	return nil
}

func childNode(node *Node, name string) *Node {
	for c := range node.Children {
		if node.Children[c].Name == name {
			return &node.Children[c]
		}
	}
	return nil
}

func isAnimationClass(class string) bool {
	return strings.HasPrefix(class, "Animation") || strings.HasPrefix(class, "Anim")
}

func asFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func asInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case float32:
		if math.Trunc(float64(v)) == float64(v) {
			return int64(v), true
		}
	case float64:
		if math.Trunc(v) == v {
			return int64(v), true
		}
	}
	return 0, false
}

func ToLoadResult(doc Document) (load_result.Result, error) {
	if _, err := BuildSceneIndex(doc); err != nil {
		return load_result.Result{}, err
	}
	return load_result.Result{}, nil
}
