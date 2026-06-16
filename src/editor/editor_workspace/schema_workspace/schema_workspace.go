/******************************************************************************/
/* schema_workspace.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	ID          = "schema"
	DisplayName = "Schema"

	uiFile = "editor/ui/workspace/schema_workspace.go.html"
)

const (
	schemaTypeString    = "string"
	schemaTypeInteger   = "integer"
	schemaTypeNumber    = "number"
	schemaTypeBoolean   = "boolean"
	schemaTypeObject    = "object"
	schemaTypeArray     = "array"
	schemaTypeReference = "reference"
	schemaTypeOneOf     = "oneOf"
	schemaTypeAnyOf     = "anyOf"
	schemaTypeAllOf     = "allOf"
	schemaTypeAny       = "any"
	schemaTypeNull      = "null"
)

var schemaTypeOrder = []schemaTypeOption{
	{Name: "String", Value: "00:string"},
	{Name: "Integer", Value: "01:integer"},
	{Name: "Number", Value: "02:number"},
	{Name: "Boolean", Value: "03:boolean"},
	{Name: "Object", Value: "04:object"},
	{Name: "Array", Value: "05:array"},
	{Name: "Reference", Value: "06:reference"},
	{Name: "One of", Value: "07:oneOf"},
	{Name: "Any of", Value: "08:anyOf"},
	{Name: "All of", Value: "09:allOf"},
	{Name: "Any", Value: "10:any"},
	{Name: "Null", Value: "11:null"},
}

func init() {
	editor_workspace_registry.Register(&SchemaWorkspace{})
}

type SchemaWorkspace struct {
	common_workspace.CommonWorkspace
	doc            schemaDocument
	isOpen         bool
	refreshQueued  bool
	selectedAnchor string
}

type schemaWorkspaceData struct {
	Document    *schemaDocument
	TypeOptions []schemaTypeOption
	SchemaJSON  template.HTML
	Status      string
	Selected    string
}

type schemaTypeOption struct {
	Name  string
	Value string
}

type schemaDocument struct {
	Name        string
	SchemaURI   string
	ID          string
	Title       string
	Description string
	Root        schemaNode
	Definitions []schemaDefinition
	Status      string
}

type schemaDefinition struct {
	Name string
	Node schemaNode
}

type schemaField struct {
	Name     string
	Required bool
	Node     schemaNode
}

type schemaNode struct {
	Type            string
	Nullable        bool
	OptionsOpen     bool
	Title           string
	Description     string
	DefaultJSON     string
	EnumValues      string
	Format          string
	MinLength       string
	MaxLength       string
	Pattern         string
	Minimum         string
	Maximum         string
	MultipleOf      string
	Reference       string
	CombinationRefs string
	ItemsType       string
	ItemsReference  string
	AllowAdditional bool
	Fields          []schemaField
}

func (w *SchemaWorkspace) ID() string          { return ID }
func (w *SchemaWorkspace) DisplayName() string { return DisplayName }
func (w *SchemaWorkspace) IsRequired() bool    { return false }

func (w *SchemaWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	defer tracing.NewRegion("SchemaWorkspace.Initialize").End()
	w.doc = defaultSchemaDocument()
	w.selectedAnchor = "root"
	return w.CommonWorkspace.InitializeWithUI(ed.Host(), uiFile, w.uiData(), w.funcMap())
}

func (w *SchemaWorkspace) Shutdown() {
	defer tracing.NewRegion("SchemaWorkspace.Shutdown").End()
	w.CommonShutdown()
}

func (w *SchemaWorkspace) Open() {
	defer tracing.NewRegion("SchemaWorkspace.Open").End()
	w.isOpen = true
	w.CommonOpen()
}

func (w *SchemaWorkspace) Close() {
	defer tracing.NewRegion("SchemaWorkspace.Close").End()
	w.isOpen = false
	w.CommonClose()
}

func (w *SchemaWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func defaultSchemaDocument() schemaDocument {
	return schemaDocument{
		Name:        "Draft schema",
		SchemaURI:   "https://json-schema.org/draft/2020-12/schema",
		Title:       "Draft schema",
		Description: "",
		Root: schemaNode{
			Type:            schemaTypeObject,
			AllowAdditional: true,
			Fields: []schemaField{
				{
					Name:     "id",
					Required: true,
					Node: schemaNode{
						Type: schemaTypeString,
					},
				},
				{
					Name:     "price",
					Required: true,
					Node: schemaNode{
						Type:        schemaTypeNumber,
						OptionsOpen: true,
					},
				},
			},
		},
		Definitions: []schemaDefinition{
			{
				Name: "address",
				Node: schemaNode{
					Type:            schemaTypeObject,
					AllowAdditional: false,
				},
			},
		},
	}
}

func defaultSchemaField(name string) schemaField {
	return schemaField{
		Name: name,
		Node: schemaNode{
			Type: schemaTypeString,
		},
	}
}

func defaultSchemaDefinition(index int) schemaDefinition {
	return schemaDefinition{
		Name: fmt.Sprintf("definition%d", index+1),
		Node: schemaNode{
			Type:            schemaTypeObject,
			AllowAdditional: false,
		},
	}
}

func (w *SchemaWorkspace) uiData() schemaWorkspaceData {
	return schemaWorkspaceData{
		Document:    &w.doc,
		TypeOptions: schemaTypeOrder,
		SchemaJSON:  template.HTML(template.HTMLEscapeString(w.schemaJSON())),
		Status:      w.doc.Status,
		Selected:    w.selectedAnchor,
	}
}

func (w *SchemaWorkspace) funcMap() map[string]func(*document.Element) {
	return map[string]func(*document.Element){
		"schemaValueChanged":       w.schemaValueChanged,
		"schemaToggleBool":         w.schemaToggleBool,
		"schemaAddRootField":       w.schemaAddRootField,
		"schemaAddDefinition":      w.schemaAddDefinition,
		"schemaAddDefinitionField": w.schemaAddDefinitionField,
		"schemaRemoveField":        w.schemaRemoveField,
		"schemaRemoveDefinition":   w.schemaRemoveDefinition,
		"schemaNewDocument":        w.schemaNewDocument,
		"schemaLoadDocument":       w.schemaLoadDocument,
		"schemaSaveDocument":       w.schemaSaveDocument,
		"schemaRefreshJSON":        w.schemaRefreshJSON,
		"schemaApplyJSON":          w.schemaApplyJSON,
		"schemaSelectAnchor":       w.schemaSelectAnchor,
	}
}

func (w *SchemaWorkspace) schemaValueChanged(e *document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaValueChanged").End()
	w.applyElementValue(e)
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaToggleBool(e *document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaToggleBool").End()
	w.captureCurrentInputs()
	if b := w.boolTarget(e); b != nil {
		*b = !*b
	}
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaAddRootField(*document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaAddRootField").End()
	w.captureCurrentInputs()
	w.doc.Root.Fields = append(w.doc.Root.Fields, defaultSchemaField(uniqueFieldName(w.doc.Root.Fields, "field")))
	w.selectedAnchor = "root"
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaAddDefinition(e *document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaAddDefinition").End()
	w.captureCurrentInputs()
	w.doc.Definitions = append(w.doc.Definitions, defaultSchemaDefinition(len(w.doc.Definitions)))
	w.selectedAnchor = fmt.Sprintf("definition-%d", len(w.doc.Definitions)-1)
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaAddDefinitionField(e *document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaAddDefinitionField").End()
	w.captureCurrentInputs()
	defIdx, ok := indexAttr(e, "data-def-index")
	if !ok || defIdx < 0 || defIdx >= len(w.doc.Definitions) {
		return
	}
	def := &w.doc.Definitions[defIdx]
	def.Node.Fields = append(def.Node.Fields, defaultSchemaField(uniqueFieldName(def.Node.Fields, "nestedField")))
	w.selectedAnchor = fmt.Sprintf("definition-%d", defIdx)
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaRemoveField(e *document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaRemoveField").End()
	w.captureCurrentInputs()
	target := e.Attribute("data-target")
	switch target {
	case "root-field":
		idx, ok := indexAttr(e, "data-field-index")
		if ok && idx >= 0 && idx < len(w.doc.Root.Fields) {
			w.doc.Root.Fields = append(w.doc.Root.Fields[:idx], w.doc.Root.Fields[idx+1:]...)
		}
	case "definition-field":
		defIdx, ok := indexAttr(e, "data-def-index")
		fieldIdx, ok2 := indexAttr(e, "data-field-index")
		if ok && ok2 && defIdx >= 0 && defIdx < len(w.doc.Definitions) {
			fields := &w.doc.Definitions[defIdx].Node.Fields
			if fieldIdx >= 0 && fieldIdx < len(*fields) {
				*fields = append((*fields)[:fieldIdx], (*fields)[fieldIdx+1:]...)
			}
			w.selectedAnchor = fmt.Sprintf("definition-%d", defIdx)
		}
	}
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaRemoveDefinition(e *document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaRemoveDefinition").End()
	w.captureCurrentInputs()
	idx, ok := indexAttr(e, "data-def-index")
	if ok && idx >= 0 && idx < len(w.doc.Definitions) {
		w.doc.Definitions = append(w.doc.Definitions[:idx], w.doc.Definitions[idx+1:]...)
	}
	w.selectedAnchor = "root"
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaNewDocument(*document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaNewDocument").End()
	w.doc = defaultSchemaDocument()
	w.doc.Status = "Started a new in-memory schema"
	w.selectedAnchor = "root"
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaLoadDocument(*document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaLoadDocument").End()
	w.captureCurrentInputs()
	w.doc.Status = "Load is not wired yet; editing the in-memory schema"
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaSaveDocument(*document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaSaveDocument").End()
	w.captureCurrentInputs()
	w.doc.Status = "Save is not wired yet; schema is kept in memory"
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaRefreshJSON(*document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaRefreshJSON").End()
	w.captureCurrentInputs()
	w.doc.Status = "Schema JSON refreshed"
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaApplyJSON(*document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaApplyJSON").End()
	if w.Doc == nil {
		return
	}
	elm, ok := w.Doc.GetElementById("schemaJSONEditor")
	if !ok || elm.UI == nil || !elm.UI.IsType(ui.ElementTypeTextArea) {
		return
	}
	text := elm.UI.ToTextArea().Text()
	if err := w.applySchemaJSON(text); err != nil {
		w.doc.Status = "JSON error: " + err.Error()
		slog.Warn("failed to apply schema JSON", "error", err)
	} else {
		w.doc.Status = "Applied JSON to in-memory schema"
		w.selectedAnchor = "root"
	}
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) schemaSelectAnchor(e *document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.schemaSelectAnchor").End()
	w.captureCurrentInputs()
	if anchor := e.Attribute("data-anchor"); anchor != "" {
		w.selectedAnchor = anchor
	}
	w.scheduleRefresh()
}

func (w *SchemaWorkspace) scheduleRefresh() {
	if w.Host == nil || w.refreshQueued {
		return
	}
	w.refreshQueued = true
	w.Host.RunNextFrame(func() {
		w.refreshQueued = false
		if err := w.ReloadUI(uiFile, w.uiData(), w.funcMap()); err != nil {
			slog.Error("failed to reload schema workspace UI", "error", err)
			return
		}
		if w.isOpen && w.Doc != nil {
			w.Doc.Activate()
			if !w.IsBlurred {
				w.UiMan.EnableUpdate()
			}
		}
	})
}

func (w *SchemaWorkspace) captureCurrentInputs() {
	if w.Doc == nil {
		return
	}
	for _, elm := range w.Doc.GetElementsByClass("schemaBinding") {
		w.applyElementValue(elm)
	}
}

func (w *SchemaWorkspace) applyElementValue(e *document.Element) {
	if e == nil || e.UI == nil {
		return
	}
	value := schemaElementValue(e)
	prop := e.Attribute("data-prop")
	switch e.Attribute("data-target") {
	case "document":
		switch prop {
		case "Name":
			if w.doc.Title == "" || w.doc.Title == w.doc.Name {
				w.doc.Title = value
			}
			w.doc.Name = value
		case "SchemaURI":
			w.doc.SchemaURI = value
		case "ID":
			w.doc.ID = value
		case "Title":
			w.doc.Title = value
		case "Description":
			w.doc.Description = value
		}
	case "root-field":
		if field := w.rootFieldTarget(e); field != nil {
			applySchemaFieldValue(field, prop, value)
		}
	case "definition":
		if def := w.definitionTarget(e); def != nil && prop == "Name" {
			def.Name = value
		}
	case "definition-root":
		if def := w.definitionTarget(e); def != nil {
			applySchemaNodeValue(&def.Node, prop, value)
		}
	case "definition-field":
		if field := w.definitionFieldTarget(e); field != nil {
			applySchemaFieldValue(field, prop, value)
		}
	}
}

func schemaElementValue(e *document.Element) string {
	switch e.UI.Type() {
	case ui.ElementTypeInput:
		return e.UI.ToInput().Text()
	case ui.ElementTypeSelect:
		return normalizeSchemaType(e.UI.ToSelect().Value())
	case ui.ElementTypeTextArea:
		return e.UI.ToTextArea().Text()
	default:
		return e.Attribute("value")
	}
}

func (w *SchemaWorkspace) rootFieldTarget(e *document.Element) *schemaField {
	idx, ok := indexAttr(e, "data-field-index")
	if !ok || idx < 0 || idx >= len(w.doc.Root.Fields) {
		return nil
	}
	return &w.doc.Root.Fields[idx]
}

func (w *SchemaWorkspace) definitionTarget(e *document.Element) *schemaDefinition {
	idx, ok := indexAttr(e, "data-def-index")
	if !ok || idx < 0 || idx >= len(w.doc.Definitions) {
		return nil
	}
	return &w.doc.Definitions[idx]
}

func (w *SchemaWorkspace) definitionFieldTarget(e *document.Element) *schemaField {
	def := w.definitionTarget(e)
	fieldIdx, ok := indexAttr(e, "data-field-index")
	if def == nil || !ok || fieldIdx < 0 || fieldIdx >= len(def.Node.Fields) {
		return nil
	}
	return &def.Node.Fields[fieldIdx]
}

func (w *SchemaWorkspace) boolTarget(e *document.Element) *bool {
	prop := e.Attribute("data-prop")
	switch e.Attribute("data-target") {
	case "root":
		if prop == "AllowAdditional" {
			return &w.doc.Root.AllowAdditional
		}
	case "root-field":
		if field := w.rootFieldTarget(e); field != nil {
			return schemaFieldBoolTarget(field, prop)
		}
	case "definition-root":
		if def := w.definitionTarget(e); def != nil {
			return schemaNodeBoolTarget(&def.Node, prop)
		}
	case "definition-field":
		if field := w.definitionFieldTarget(e); field != nil {
			return schemaFieldBoolTarget(field, prop)
		}
	}
	return nil
}

func schemaFieldBoolTarget(field *schemaField, prop string) *bool {
	switch prop {
	case "Required":
		return &field.Required
	case "Nullable", "OptionsOpen", "AllowAdditional":
		return schemaNodeBoolTarget(&field.Node, prop)
	default:
		return nil
	}
}

func schemaNodeBoolTarget(node *schemaNode, prop string) *bool {
	switch prop {
	case "Nullable":
		return &node.Nullable
	case "OptionsOpen":
		return &node.OptionsOpen
	case "AllowAdditional":
		return &node.AllowAdditional
	default:
		return nil
	}
}

func applySchemaFieldValue(field *schemaField, prop, value string) {
	switch prop {
	case "Name":
		field.Name = value
	case "Required":
		field.Required = value == "true"
	default:
		applySchemaNodeValue(&field.Node, prop, value)
	}
}

func applySchemaNodeValue(node *schemaNode, prop, value string) {
	switch prop {
	case "Type":
		node.Type = normalizeSchemaType(value)
		if node.ItemsType == "" {
			node.ItemsType = schemaTypeString
		}
	case "Title":
		node.Title = value
	case "Description":
		node.Description = value
	case "DefaultJSON":
		node.DefaultJSON = value
	case "EnumValues":
		node.EnumValues = value
	case "Format":
		node.Format = value
	case "MinLength":
		node.MinLength = value
	case "MaxLength":
		node.MaxLength = value
	case "Pattern":
		node.Pattern = value
	case "Minimum":
		node.Minimum = value
	case "Maximum":
		node.Maximum = value
	case "MultipleOf":
		node.MultipleOf = value
	case "Reference":
		node.Reference = value
	case "CombinationRefs":
		node.CombinationRefs = value
	case "ItemsType":
		node.ItemsType = normalizeSchemaType(value)
	case "ItemsReference":
		node.ItemsReference = value
	}
}

func (w *SchemaWorkspace) schemaJSON() string {
	data := w.doc.schemaMap()
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(out)
}

func (d schemaDocument) schemaMap() map[string]any {
	out := d.Root.toSchemaMap()
	out["$schema"] = firstNonEmpty(d.SchemaURI, "https://json-schema.org/draft/2020-12/schema")
	if d.ID != "" {
		out["$id"] = d.ID
	}
	if d.Title != "" {
		out["title"] = d.Title
	}
	if d.Description != "" {
		out["description"] = d.Description
	}
	if len(d.Definitions) > 0 {
		defs := map[string]any{}
		for i := range d.Definitions {
			name := strings.TrimSpace(d.Definitions[i].Name)
			if name == "" {
				continue
			}
			defs[name] = d.Definitions[i].Node.toSchemaMap()
		}
		if len(defs) > 0 {
			out["$defs"] = defs
		}
	}
	return out
}

func (n schemaNode) toSchemaMap() map[string]any {
	out := map[string]any{}
	switch n.Type {
	case schemaTypeReference:
		ref := strings.TrimSpace(n.Reference)
		if ref == "" {
			ref = "#/$defs/address"
		}
		out["$ref"] = ref
	case schemaTypeAny:
	case schemaTypeOneOf, schemaTypeAnyOf, schemaTypeAllOf:
		key := n.Type
		refs := strings.Fields(strings.ReplaceAll(n.CombinationRefs, ",", "\n"))
		items := make([]any, 0, len(refs))
		for _, ref := range refs {
			items = append(items, map[string]any{"$ref": ref})
		}
		if len(items) == 0 {
			items = append(items, map[string]any{})
		}
		out[key] = items
	case schemaTypeObject:
		out["type"] = schemaTypeObject
		if len(n.Fields) > 0 {
			props := map[string]any{}
			required := make([]string, 0)
			for i := range n.Fields {
				name := strings.TrimSpace(n.Fields[i].Name)
				if name == "" {
					continue
				}
				props[name] = n.Fields[i].Node.toSchemaMap()
				if n.Fields[i].Required {
					required = append(required, name)
				}
			}
			out["properties"] = props
			if len(required) > 0 {
				out["required"] = required
			}
		}
		out["additionalProperties"] = n.AllowAdditional
	case schemaTypeArray:
		out["type"] = schemaTypeArray
		item := schemaNode{Type: firstNonEmpty(n.ItemsType, schemaTypeString), Reference: n.ItemsReference}
		out["items"] = item.toSchemaMap()
	case schemaTypeString:
		out["type"] = schemaTypeString
		addStringIfSet(out, "format", n.Format)
		addIntIfSet(out, "minLength", n.MinLength)
		addIntIfSet(out, "maxLength", n.MaxLength)
		addStringIfSet(out, "pattern", n.Pattern)
	case schemaTypeInteger:
		out["type"] = schemaTypeInteger
		addNumberIfSet(out, "minimum", n.Minimum)
		addNumberIfSet(out, "maximum", n.Maximum)
		addNumberIfSet(out, "multipleOf", n.MultipleOf)
	case schemaTypeNumber:
		out["type"] = schemaTypeNumber
		addNumberIfSet(out, "minimum", n.Minimum)
		addNumberIfSet(out, "maximum", n.Maximum)
		addNumberIfSet(out, "multipleOf", n.MultipleOf)
	case schemaTypeBoolean:
		out["type"] = schemaTypeBoolean
	case schemaTypeNull:
		out["type"] = schemaTypeNull
	default:
		out["type"] = schemaTypeString
	}
	addStringIfSet(out, "title", n.Title)
	addStringIfSet(out, "description", n.Description)
	if v, ok := parseLooseJSON(n.DefaultJSON); ok {
		out["default"] = v
	}
	if enum := parseEnumValues(n.EnumValues); len(enum) > 0 {
		out["enum"] = enum
	}
	if n.Nullable {
		out = nullableSchema(out)
	}
	return out
}

func nullableSchema(schema map[string]any) map[string]any {
	if t, ok := schema["type"].(string); ok && t != schemaTypeNull {
		schema["type"] = []any{t, schemaTypeNull}
		return schema
	}
	if _, hasRef := schema["$ref"]; hasRef {
		return map[string]any{"anyOf": []any{schema, map[string]any{"type": schemaTypeNull}}}
	}
	return schema
}

func (w *SchemaWorkspace) applySchemaJSON(text string) error {
	var raw map[string]any
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return err
	}
	next := defaultSchemaDocument()
	next.SchemaURI = stringFromMap(raw, "$schema")
	next.ID = stringFromMap(raw, "$id")
	next.Title = stringFromMap(raw, "title")
	next.Description = stringFromMap(raw, "description")
	if next.Title != "" {
		next.Name = next.Title
	}
	next.Root = nodeFromSchemaMap(raw)
	next.Root.Type = schemaTypeObject
	if defsRaw, ok := raw["$defs"].(map[string]any); ok {
		keys := sortedMapKeys(defsRaw)
		next.Definitions = make([]schemaDefinition, 0, len(keys))
		for _, key := range keys {
			if defMap, ok := defsRaw[key].(map[string]any); ok {
				next.Definitions = append(next.Definitions, schemaDefinition{
					Name: key,
					Node: nodeFromSchemaMap(defMap),
				})
			}
		}
	}
	w.doc = next
	return nil
}

func nodeFromSchemaMap(raw map[string]any) schemaNode {
	node := schemaNode{
		Type:            schemaTypeString,
		AllowAdditional: true,
		Title:           stringFromMap(raw, "title"),
		Description:     stringFromMap(raw, "description"),
		Format:          stringFromMap(raw, "format"),
		Pattern:         stringFromMap(raw, "pattern"),
		Reference:       stringFromMap(raw, "$ref"),
	}
	if node.Reference != "" {
		node.Type = schemaTypeReference
	}
	if v, ok := raw["default"]; ok {
		if b, err := json.Marshal(v); err == nil {
			node.DefaultJSON = string(b)
		}
	}
	if enum, ok := raw["enum"].([]any); ok {
		parts := make([]string, 0, len(enum))
		for _, item := range enum {
			if b, err := json.Marshal(item); err == nil {
				parts = append(parts, string(b))
			}
		}
		node.EnumValues = strings.Join(parts, "\n")
	}
	if v, ok := raw["minLength"]; ok {
		node.MinLength = fmt.Sprint(v)
	}
	if v, ok := raw["maxLength"]; ok {
		node.MaxLength = fmt.Sprint(v)
	}
	if v, ok := raw["minimum"]; ok {
		node.Minimum = fmt.Sprint(v)
	}
	if v, ok := raw["maximum"]; ok {
		node.Maximum = fmt.Sprint(v)
	}
	if v, ok := raw["multipleOf"]; ok {
		node.MultipleOf = fmt.Sprint(v)
	}
	if t, nullable := schemaTypeFromRaw(raw["type"]); t != "" {
		node.Type = t
		node.Nullable = nullable
	}
	if refs, ok := combinationRefs(raw, schemaTypeOneOf); ok {
		node.Type = schemaTypeOneOf
		node.CombinationRefs = refs
	}
	if refs, ok := combinationRefs(raw, schemaTypeAnyOf); ok {
		node.Type = schemaTypeAnyOf
		node.CombinationRefs = refs
	}
	if refs, ok := combinationRefs(raw, schemaTypeAllOf); ok {
		node.Type = schemaTypeAllOf
		node.CombinationRefs = refs
	}
	if add, ok := raw["additionalProperties"].(bool); ok {
		node.AllowAdditional = add
	}
	if props, ok := raw["properties"].(map[string]any); ok {
		node.Type = schemaTypeObject
		required := stringSetFromRaw(raw["required"])
		keys := sortedMapKeys(props)
		node.Fields = make([]schemaField, 0, len(keys))
		for _, key := range keys {
			if propMap, ok := props[key].(map[string]any); ok {
				node.Fields = append(node.Fields, schemaField{
					Name:     key,
					Required: required[key],
					Node:     nodeFromSchemaMap(propMap),
				})
			}
		}
	}
	if items, ok := raw["items"].(map[string]any); ok {
		itemNode := nodeFromSchemaMap(items)
		node.ItemsType = itemNode.Type
		node.ItemsReference = itemNode.Reference
	}
	return node
}

func schemaTypeFromRaw(raw any) (string, bool) {
	switch v := raw.(type) {
	case string:
		return normalizeSchemaType(v), false
	case []any:
		typ := ""
		nullable := false
		for _, item := range v {
			if s, ok := item.(string); ok {
				if s == schemaTypeNull {
					nullable = true
				} else if typ == "" {
					typ = normalizeSchemaType(s)
				}
			}
		}
		return typ, nullable
	default:
		return "", false
	}
}

func combinationRefs(raw map[string]any, key string) (string, bool) {
	items, ok := raw[key].([]any)
	if !ok {
		return "", false
	}
	refs := make([]string, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			if ref := stringFromMap(m, "$ref"); ref != "" {
				refs = append(refs, ref)
			}
		}
	}
	return strings.Join(refs, "\n"), true
}

func stringSetFromRaw(raw any) map[string]bool {
	out := map[string]bool{}
	if list, ok := raw.([]any); ok {
		for _, item := range list {
			if s, ok := item.(string); ok {
				out[s] = true
			}
		}
	}
	return out
}

func stringFromMap(raw map[string]any, key string) string {
	if v, ok := raw[key].(string); ok {
		return v
	}
	return ""
}

func sortedMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func addStringIfSet(out map[string]any, key, value string) {
	if strings.TrimSpace(value) != "" {
		out[key] = strings.TrimSpace(value)
	}
}

func addIntIfSet(out map[string]any, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	if v, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
		out[key] = v
	}
}

func addNumberIfSet(out map[string]any, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	if v, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil {
		out[key] = v
	}
}

func parseLooseJSON(value string) (any, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, false
	}
	var out any
	if err := json.Unmarshal([]byte(value), &out); err == nil {
		return out, true
	}
	return value, true
}

func parseEnumValues(value string) []any {
	lines := strings.FieldsFunc(value, func(r rune) bool { return r == '\n' || r == ',' })
	out := make([]any, 0, len(lines))
	for _, line := range lines {
		if v, ok := parseLooseJSON(line); ok {
			out = append(out, v)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func uniqueFieldName(fields []schemaField, prefix string) string {
	used := map[string]bool{}
	for i := range fields {
		used[fields[i].Name] = true
	}
	for i := 1; ; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		if !used[name] {
			return name
		}
	}
}

func indexAttr(e *document.Element, key string) (int, bool) {
	idx, err := strconv.Atoi(e.Attribute(key))
	if err != nil {
		return 0, false
	}
	return idx, true
}

func normalizeSchemaType(value string) string {
	value = strings.TrimSpace(value)
	if idx := strings.Index(value, ":"); idx >= 0 {
		value = value[idx+1:]
	}
	switch value {
	case schemaTypeString, schemaTypeInteger, schemaTypeNumber, schemaTypeBoolean,
		schemaTypeObject, schemaTypeArray, schemaTypeReference, schemaTypeOneOf,
		schemaTypeAnyOf, schemaTypeAllOf, schemaTypeAny, schemaTypeNull:
		return value
	default:
		return schemaTypeString
	}
}

func schemaTypeSelectValue(value string) string {
	value = normalizeSchemaType(value)
	for i := range schemaTypeOrder {
		if strings.HasSuffix(schemaTypeOrder[i].Value, ":"+value) {
			return schemaTypeOrder[i].Value
		}
	}
	return schemaTypeOrder[0].Value
}

func (n schemaNode) TypeSelectValue() string { return schemaTypeSelectValue(n.Type) }

func (n schemaNode) ItemsTypeSelectValue() string {
	return schemaTypeSelectValue(firstNonEmpty(n.ItemsType, schemaTypeString))
}

func (n schemaNode) IsString() bool { return n.Type == schemaTypeString }
func (n schemaNode) IsNumberLike() bool {
	return n.Type == schemaTypeNumber || n.Type == schemaTypeInteger
}
func (n schemaNode) IsObject() bool    { return n.Type == schemaTypeObject }
func (n schemaNode) IsArray() bool     { return n.Type == schemaTypeArray }
func (n schemaNode) IsReference() bool { return n.Type == schemaTypeReference }
func (n schemaNode) IsCombinator() bool {
	return n.Type == schemaTypeOneOf || n.Type == schemaTypeAnyOf || n.Type == schemaTypeAllOf
}

func (n schemaNode) TypeLabel() string {
	for i := range schemaTypeOrder {
		if strings.HasSuffix(schemaTypeOrder[i].Value, ":"+normalizeSchemaType(n.Type)) {
			return schemaTypeOrder[i].Name
		}
	}
	return "String"
}

func (d schemaDocument) DefinitionsPanelHeight() int {
	height := 54
	for i := range d.Definitions {
		height += d.Definitions[i].PanelHeight()
	}
	return max(180, height)
}

func (d schemaDefinition) PanelHeight() int {
	return 84 + d.Node.CardHeight()
}

func (n schemaNode) PanelHeight() int {
	height := 112
	for i := range n.Fields {
		height += n.Fields[i].CardHeight()
	}
	return max(220, height)
}

func (f schemaField) CardHeight() int {
	return 72 + f.Node.OptionsHeight()
}

func (n schemaNode) CardHeight() int {
	return 72 + n.OptionsHeight()
}

func (n schemaNode) OptionsHeight() int {
	if !n.OptionsOpen {
		return 24
	}
	height := 166
	switch {
	case n.IsString():
		height += 54
	case n.IsNumberLike():
		height += 54
	case n.IsObject():
		height += 112
		for i := range n.Fields {
			height += n.Fields[i].CardHeight()
		}
	case n.IsArray(), n.IsReference(), n.IsCombinator():
		height += 62
	}
	return height
}
