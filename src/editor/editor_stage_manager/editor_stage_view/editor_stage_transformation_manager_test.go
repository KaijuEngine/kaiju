/******************************************************************************/
/* editor_stage_transformation_manager_test.go                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"reflect"
	"testing"
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

func TestTransformationManagerRefreshToolVisibilityKeepsOnlyCurrentToolVisible(t *testing.T) {
	tm := TransformationManager{currentTool: ToolStateMove}
	seedTransformationManagerDrawInstances(&tm)
	setToolVisible(&tm.translateTool, true)
	setToolVisible(&tm.rotationTool, true)
	setToolVisible(&tm.scalingTool, true)
	activateTransformationManagerDrawInstances(&tm)

	tm.refreshToolVisibilityAt(true, matrix.Vec3Zero())

	if !translationArrowDraw(&tm, 0).IsInView() {
		t.Fatal("expected translation gizmo to stay visible")
	}
	if rotationCircleDraw(&tm, 0).IsInView() {
		t.Fatal("expected stale rotation gizmo to be hidden")
	}
	if scalingShaftDraw(&tm, 0).IsInView() {
		t.Fatal("expected stale scale gizmo to be hidden")
	}
}

func seedTransformationManagerDrawInstances(tm *TransformationManager) {
	for i := range 3 {
		setTranslationArrowDraw(tm, i, newTransformVisibilityTestDraw())
		setTranslationPlaneDraw(tm, i, newTransformVisibilityTestDraw())
		setRotationCircleDraw(tm, i, newTransformVisibilityTestDraw())
		setScalingShaftDraw(tm, i, newTransformVisibilityTestDraw())
		setScalingBoxDraw(tm, i, newTransformVisibilityTestDraw())
	}
}

func activateTransformationManagerDrawInstances(tm *TransformationManager) {
	for i := range 3 {
		translationArrowDraw(tm, i).Activate()
		translationPlaneDraw(tm, i).Activate()
		rotationCircleDraw(tm, i).Activate()
		scalingShaftDraw(tm, i).Activate()
		scalingBoxDraw(tm, i).Activate()
	}
}

func newTransformVisibilityTestDraw() rendering.DrawInstance {
	return shader_data_registry.Create("unlit")
}

func setToolVisible(tool any, visible bool) {
	setUnexportedField(reflect.ValueOf(tool).Elem().
		FieldByName("TransformGizmo").FieldByName("visible"), visible)
}

func setTranslationArrowDraw(tm *TransformationManager, index int, draw rendering.DrawInstance) {
	setUnexportedField(translationToolArrayField(tm, "arrows").Index(index).FieldByName("shaderData"), draw)
}

func setTranslationPlaneDraw(tm *TransformationManager, index int, draw rendering.DrawInstance) {
	setUnexportedField(translationToolArrayField(tm, "planes").Index(index).FieldByName("shaderData"), draw)
}

func setRotationCircleDraw(tm *TransformationManager, index int, draw rendering.DrawInstance) {
	setUnexportedField(rotationToolArrayField(tm, "circles").Index(index).FieldByName("shaderData"), draw)
}

func setScalingShaftDraw(tm *TransformationManager, index int, draw rendering.DrawInstance) {
	setUnexportedField(scalingToolArrayField(tm, "boxes").Index(index).FieldByName("shaftShaderData"), draw)
}

func setScalingBoxDraw(tm *TransformationManager, index int, draw rendering.DrawInstance) {
	setUnexportedField(scalingToolArrayField(tm, "boxes").Index(index).FieldByName("boxShaderData"), draw)
}

func translationArrowDraw(tm *TransformationManager, index int) rendering.DrawInstance {
	return drawInstanceFromField(translationToolArrayField(tm, "arrows").Index(index).FieldByName("shaderData"))
}

func translationPlaneDraw(tm *TransformationManager, index int) rendering.DrawInstance {
	return drawInstanceFromField(translationToolArrayField(tm, "planes").Index(index).FieldByName("shaderData"))
}

func rotationCircleDraw(tm *TransformationManager, index int) rendering.DrawInstance {
	return drawInstanceFromField(rotationToolArrayField(tm, "circles").Index(index).FieldByName("shaderData"))
}

func scalingShaftDraw(tm *TransformationManager, index int) rendering.DrawInstance {
	return drawInstanceFromField(scalingToolArrayField(tm, "boxes").Index(index).FieldByName("shaftShaderData"))
}

func scalingBoxDraw(tm *TransformationManager, index int) rendering.DrawInstance {
	return drawInstanceFromField(scalingToolArrayField(tm, "boxes").Index(index).FieldByName("boxShaderData"))
}

func translationToolArrayField(tm *TransformationManager, name string) reflect.Value {
	return reflect.ValueOf(&tm.translateTool).Elem().FieldByName(name)
}

func rotationToolArrayField(tm *TransformationManager, name string) reflect.Value {
	return reflect.ValueOf(&tm.rotationTool).Elem().FieldByName(name)
}

func scalingToolArrayField(tm *TransformationManager, name string) reflect.Value {
	return reflect.ValueOf(&tm.scalingTool).Elem().FieldByName(name)
}

func setUnexportedField(field reflect.Value, value any) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().Set(reflect.ValueOf(value))
}

func drawInstanceFromField(field reflect.Value) rendering.DrawInstance {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().Interface().(rendering.DrawInstance)
}
