/******************************************************************************/
/* translation_tool_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package transform_tools

import (
	"testing"

	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/matrix"
)

func TestTranslationToolVisiblePlaneCount(t *testing.T) {
	tool := TranslationTool{}
	if got := tool.visibleArrowCount(); got != 3 {
		t.Fatalf("3D/default mode should show all arrow handles, got %d", got)
	}
	if got := tool.visiblePlaneCount(); got != 3 {
		t.Fatalf("3D/default mode should show all plane handles, got %d", got)
	}
	tool.cameraMode = editor_controls.EditorCameraMode2d
	if got := tool.visibleArrowCount(); got != 2 {
		t.Fatalf("2D mode should show the X/Y arrow handles only, got %d", got)
	}
	if got := tool.visiblePlaneCount(); got != 1 {
		t.Fatalf("2D mode should show the XY plane handle only, got %d", got)
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
