/******************************************************************************/
/* shader_graph_node_field_value.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shading_workspace

import "kaijuengine.com/matrix"

type shaderGraphNodeFieldValue struct {
	Text   string
	Parts  []string
	Bool   bool
	Color  matrix.Color
	Option string
}

func shaderGraphDefaultFieldValue(spec shaderGraphNodeFieldSpec) shaderGraphNodeFieldValue {
	switch spec.Type {
	case shaderGraphNodeFieldBool:
		return shaderGraphNodeFieldValue{Bool: spec.DefaultBool}
	case shaderGraphNodeFieldColor:
		color := spec.DefaultColor
		if matrix.Vec4Approx(matrix.Vec4(color), matrix.Vec4(matrix.ColorClear())) {
			color = matrix.ColorWhite()
		}
		return shaderGraphNodeFieldValue{Color: color}
	case shaderGraphNodeFieldVector3:
		return shaderGraphNodeFieldValue{Parts: shaderGraphFieldParts(spec.DefaultValues, 3)}
	case shaderGraphNodeFieldSelect:
		value := spec.Default
		if value == "" && len(spec.Options) > 0 {
			value = spec.Options[0].Value
		}
		return shaderGraphNodeFieldValue{Option: value}
	default:
		return shaderGraphNodeFieldValue{Text: spec.Default}
	}
}

func shaderGraphFieldParts(values []string, count int) []string {
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
