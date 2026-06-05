/******************************************************************************/
/* render_graph_document.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"encoding/json"
	"fmt"
	"strings"

	"kaijuengine.com/matrix"
)

const renderGraphDocumentVersion = 1

type RenderGraphDocument struct {
	Version     int                     `json:"version"`
	Pan         matrix.Vec2             `json:"pan,omitempty"`
	Nodes       []RenderGraphNode       `json:"nodes"`
	Connections []RenderGraphConnection `json:"connections,omitempty"`
}

type RenderGraphNode struct {
	ID       string                           `json:"id"`
	Type     string                           `json:"type"`
	Position matrix.Vec2                      `json:"position"`
	Values   map[string]RenderGraphFieldValue `json:"values,omitempty"`
}

type RenderGraphFieldValue struct {
	Text   string        `json:"text,omitempty"`
	Parts  []string      `json:"parts,omitempty"`
	Bool   *bool         `json:"bool,omitempty"`
	Color  *matrix.Color `json:"color,omitempty"`
	Option string        `json:"option,omitempty"`
}

type RenderGraphConnection struct {
	Output RenderGraphPortRef `json:"output"`
	Input  RenderGraphPortRef `json:"input"`
}

type RenderGraphPortRef struct {
	Node string `json:"node"`
	Port int    `json:"port"`
}

func SerializeRenderGraphDocument(document RenderGraphDocument) ([]byte, error) {
	if document.Version == 0 {
		document.Version = renderGraphDocumentVersion
	}
	if err := validateRenderGraphDocument(document); err != nil {
		return nil, err
	}
	return json.MarshalIndent(document, "", "\t")
}

func DeserializeRenderGraphDocument(data []byte) (RenderGraphDocument, error) {
	document := RenderGraphDocument{}
	if err := json.Unmarshal(data, &document); err != nil {
		return document, err
	}
	if err := validateRenderGraphDocument(document); err != nil {
		return document, err
	}
	if document.Version == 0 {
		document.Version = renderGraphDocumentVersion
	}
	return document, nil
}

func validateRenderGraphDocument(document RenderGraphDocument) error {
	if document.Version != 0 && document.Version != renderGraphDocumentVersion {
		return fmt.Errorf("unsupported render graph version %d", document.Version)
	}
	ids := make(map[string]struct{}, len(document.Nodes))
	for i := range document.Nodes {
		node := document.Nodes[i]
		if strings.TrimSpace(node.ID) == "" {
			return fmt.Errorf("render graph node %d has an empty id", i)
		}
		if _, exists := ids[node.ID]; exists {
			return fmt.Errorf("duplicate render graph node id %q", node.ID)
		}
		if strings.TrimSpace(node.Type) == "" {
			return fmt.Errorf("render graph node %q has an empty type", node.ID)
		}
		if _, ok := shaderGraphNodeCatalogSpec(node.Type); !ok {
			return fmt.Errorf("render graph node %q has unknown type %q", node.ID, node.Type)
		}
		ids[node.ID] = struct{}{}
	}
	for i := range document.Connections {
		connection := document.Connections[i]
		if _, ok := ids[connection.Output.Node]; !ok {
			return fmt.Errorf("render graph connection %d references unknown output node %q", i, connection.Output.Node)
		}
		if _, ok := ids[connection.Input.Node]; !ok {
			return fmt.Errorf("render graph connection %d references unknown input node %q", i, connection.Input.Node)
		}
		if connection.Output.Port < 0 || connection.Input.Port < 0 {
			return fmt.Errorf("render graph connection %d has a negative port index", i)
		}
	}
	return nil
}

func (g *shaderGraph) Serialize() ([]byte, error) {
	return SerializeRenderGraphDocument(g.Document())
}

func (g *shaderGraph) Deserialize(data []byte) error {
	document, err := DeserializeRenderGraphDocument(data)
	if err != nil {
		return err
	}
	return g.LoadDocument(document)
}

func (g *shaderGraph) Document() RenderGraphDocument {
	document := RenderGraphDocument{
		Version: renderGraphDocumentVersion,
		Pan:     g.pan,
		Nodes:   make([]RenderGraphNode, 0, len(g.nodes)),
	}
	nodeIDs := make(map[*shaderGraphNode]string, len(g.nodes))
	for i := range g.nodes {
		node := g.nodes[i]
		id := node.id
		if strings.TrimSpace(id) == "" {
			id = fmt.Sprintf("node-%d", i+1)
		}
		nodeIDs[node] = id
		document.Nodes = append(document.Nodes, RenderGraphNode{
			ID:       id,
			Type:     node.typeID,
			Position: node.position,
			Values:   renderGraphFieldValuesFromNode(node),
		})
	}
	for i := range g.connections {
		connection := g.connections[i]
		if connection.output == nil || connection.input == nil ||
			connection.output.node == nil || connection.input.node == nil {
			continue
		}
		outputNode, outputOK := nodeIDs[connection.output.node]
		inputNode, inputOK := nodeIDs[connection.input.node]
		if !outputOK || !inputOK {
			continue
		}
		document.Connections = append(document.Connections, RenderGraphConnection{
			Output: RenderGraphPortRef{Node: outputNode, Port: connection.output.index},
			Input:  RenderGraphPortRef{Node: inputNode, Port: connection.input.index},
		})
	}
	return document
}

func (g *shaderGraph) LoadDocument(document RenderGraphDocument) error {
	if err := validateRenderGraphDocument(document); err != nil {
		return err
	}
	if g.root == nil {
		return fmt.Errorf("render graph is not initialized")
	}
	g.clear()
	g.pan = document.Pan
	nodes := make(map[string]*shaderGraphNode, len(document.Nodes))
	for i := range document.Nodes {
		src := document.Nodes[i]
		node, ok := g.CreateCatalogNode(src.Type, src.Position)
		if !ok {
			return fmt.Errorf("failed to create render graph node %q of type %q", src.ID, src.Type)
		}
		node.id = src.ID
		for key, value := range renderGraphFieldValuesToNode(src.Values) {
			node.values[key] = value
		}
		node.applyFieldValues()
		nodes[src.ID] = node
	}
	for i := range document.Connections {
		connection := document.Connections[i]
		outputNode := nodes[connection.Output.Node]
		inputNode := nodes[connection.Input.Node]
		if g.CreateConnection(outputNode.Output(connection.Output.Port), inputNode.Input(connection.Input.Port)) == nil {
			return fmt.Errorf("failed to create render graph connection %d", i)
		}
	}
	g.applyViewOffsets()
	return nil
}

func renderGraphFieldValuesFromNode(node *shaderGraphNode) map[string]RenderGraphFieldValue {
	if node == nil || len(node.fields) == 0 {
		return nil
	}
	out := make(map[string]RenderGraphFieldValue, len(node.fields))
	for i := range node.fields {
		field := node.fields[i]
		value := node.FieldValue(field.spec.ID)
		switch field.spec.Type {
		case shaderGraphNodeFieldBool:
			boolValue := value.Bool
			out[field.spec.ID] = RenderGraphFieldValue{Bool: &boolValue}
		case shaderGraphNodeFieldColor:
			color := value.Color
			out[field.spec.ID] = RenderGraphFieldValue{Color: &color}
		case shaderGraphNodeFieldVector3:
			out[field.spec.ID] = RenderGraphFieldValue{Parts: append([]string(nil), value.Parts...)}
		case shaderGraphNodeFieldSelect:
			out[field.spec.ID] = RenderGraphFieldValue{Option: value.Option}
		default:
			out[field.spec.ID] = RenderGraphFieldValue{Text: value.Text}
		}
	}
	return out
}

func renderGraphFieldValuesToNode(values map[string]RenderGraphFieldValue) map[string]shaderGraphNodeFieldValue {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]shaderGraphNodeFieldValue, len(values))
	for key, value := range values {
		fieldValue := shaderGraphNodeFieldValue{
			Text:   value.Text,
			Parts:  append([]string(nil), value.Parts...),
			Option: value.Option,
		}
		if value.Bool != nil {
			fieldValue.Bool = *value.Bool
		}
		if value.Color != nil {
			fieldValue.Color = *value.Color
		}
		out[key] = fieldValue
	}
	return out
}
