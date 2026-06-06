/******************************************************************************/
/* translation_tool_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package transform_tools

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestTranslationToolDefaultModeShowsAllHandles(t *testing.T) {
	tool := TranslationTool{}
	for axis := range 3 {
		if !tool.axisVisible(axis) {
			t.Fatalf("default mode axisVisible(%d) = false, want true", axis)
		}
	}
	if axis, ok := tool.planarTranslationPlaneAxis(); ok {
		t.Fatalf("default mode planarTranslationPlaneAxis = %d, true; want no planar override", axis)
	}
}

func TestTranslationToolPlaneDragNormals(t *testing.T) {
	tests := []struct {
		name string
		axis int
		want matrix.Vec3
	}{
		{name: "xy plane", axis: matrix.Vx, want: matrix.Vec3Forward()},
		{name: "yz plane", axis: matrix.Vy, want: matrix.Vec3Right()},
		{name: "xz plane", axis: matrix.Vz, want: matrix.Vec3Up()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := TranslationTool{}
			tool.currentAxis = tt.axis
			tool.currentType = TRANSLATION_TYPE_PLANE
			if got := tool.dragPlaneNormal(nil, matrix.Vec3Zero()); got != tt.want {
				t.Fatalf("dragPlaneNormal = %v; want %v", got, tt.want)
			}
		})
	}
}
