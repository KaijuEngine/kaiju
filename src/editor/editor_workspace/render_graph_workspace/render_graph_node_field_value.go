/******************************************************************************/
/* render_graph_node_field_value.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"slices"

	"kaijuengine.com/matrix"
)

type renderGraphNodeFieldValue struct {
	Text   string
	Parts  []string
	Bool   bool
	Color  matrix.Color
	Option string
}

func (v renderGraphNodeFieldValue) Clone() renderGraphNodeFieldValue {
	v.Parts = append([]string(nil), v.Parts...)
	return v
}

func (v renderGraphNodeFieldValue) Equals(other renderGraphNodeFieldValue) bool {
	return v.Text == other.Text &&
		slices.Equal(v.Parts, other.Parts) &&
		v.Bool == other.Bool &&
		matrix.Vec4Approx(matrix.Vec4(v.Color), matrix.Vec4(other.Color)) &&
		v.Option == other.Option
}

func renderGraphDefaultFieldValue(spec renderGraphNodeFieldSpec) renderGraphNodeFieldValue {
	switch spec.Type {
	case renderGraphNodeFieldBool:
		return renderGraphNodeFieldValue{Bool: spec.DefaultBool}
	case renderGraphNodeFieldColor:
		color := spec.DefaultColor
		if matrix.Vec4Approx(matrix.Vec4(color), matrix.Vec4(matrix.ColorClear())) {
			color = matrix.ColorWhite()
		}
		return renderGraphNodeFieldValue{Color: color}
	case renderGraphNodeFieldTexture:
		return renderGraphNodeFieldValue{Text: spec.Default}
	case renderGraphNodeFieldVector2:
		return renderGraphNodeFieldValue{Parts: renderGraphFieldParts(spec.DefaultValues, 2)}
	case renderGraphNodeFieldVector3:
		return renderGraphNodeFieldValue{Parts: renderGraphFieldParts(spec.DefaultValues, 3)}
	case renderGraphNodeFieldVector4:
		return renderGraphNodeFieldValue{Parts: renderGraphFieldParts(spec.DefaultValues, 4)}
	case renderGraphNodeFieldSelect:
		value := spec.Default
		if value == "" && len(spec.Options) > 0 {
			value = spec.Options[0].Value
		}
		return renderGraphNodeFieldValue{Option: value}
	default:
		return renderGraphNodeFieldValue{Text: spec.Default}
	}
}

func renderGraphFieldParts(values []string, count int) []string {
	parts := make([]string, count)
	for i := range parts {
		if i < len(values) {
			parts[i] = values[i]
		} else {
			parts[i] = "0"
		}
	}
	return parts
}
