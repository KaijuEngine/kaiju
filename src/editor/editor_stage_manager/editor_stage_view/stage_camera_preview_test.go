/******************************************************************************/
/* stage_camera_preview_test.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"testing"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine_entity_data/engine_entity_data_camera"
	"kaijuengine.com/matrix"
)

func TestSelectedCameraPreviewBindingUsesLastSelectedCamera(t *testing.T) {
	t.Parallel()

	plain := &editor_stage_manager.StageEntity{}
	camera := &editor_stage_manager.StageEntity{}
	data := cameraPreviewTestBinding()
	camera.AddDataBinding(data)
	view := StageView{}
	view.manager.SelectEntity(plain)
	view.manager.SelectEntity(camera)

	gotEntity, gotData, ok := view.selectedCameraPreviewBinding()
	if !ok {
		t.Fatal("expected selected camera binding")
	}
	if gotEntity != camera {
		t.Fatalf("selected entity = %#v, want camera", gotEntity)
	}
	if gotData != data {
		t.Fatalf("selected data = %#v, want camera binding", gotData)
	}
}

func TestSelectedCameraPreviewBindingIgnoresNonCameraSelection(t *testing.T) {
	t.Parallel()

	view := StageView{}
	view.manager.SelectEntity(&editor_stage_manager.StageEntity{})

	if _, _, ok := view.selectedCameraPreviewBinding(); ok {
		t.Fatal("did not expect camera preview binding")
	}
}

func TestCameraPreviewDisplaySizePreservesAspectWithinBounds(t *testing.T) {
	t.Parallel()

	w, h := cameraPreviewDisplaySize(1920, 1080)
	if !matrix.ApproxTo(w/h, 1920.0/1080.0, 0.001) {
		t.Fatalf("preview aspect = %f, want 16:9", w/h)
	}
	if w > cameraPreviewMaxWidth || h > cameraPreviewMaxHeight {
		t.Fatalf("preview size = %fx%f exceeds bounds", w, h)
	}

	w, h = cameraPreviewDisplaySize(9, 16)
	if !matrix.ApproxTo(w/h, 9.0/16.0, 0.001) {
		t.Fatalf("portrait preview aspect = %f, want 9:16", w/h)
	}
	if w > cameraPreviewMaxWidth || h > cameraPreviewMaxHeight {
		t.Fatalf("portrait preview size = %fx%f exceeds bounds", w, h)
	}
}

func TestCameraPreviewCameraUsesRuntimeCameraUp(t *testing.T) {
	t.Parallel()

	entity := &editor_stage_manager.StageEntity{}
	entity.Transform.SetupRawTransform()
	entity.Transform.SetRotation(matrix.NewVec3(0, 0, 180))

	camera := cameraPreviewCameraFromBinding(entity, cameraPreviewTestBinding(), 16, 9, 160, 90)

	if !matrix.Vec3ApproxTo(camera.Up(), matrix.Vec3Up(), 0.0001) {
		t.Fatalf("preview camera up = %v, want %v", camera.Up(), matrix.Vec3Up())
	}
}

func cameraPreviewTestBinding() *entity_data_binding.EntityDataEntry {
	data := engine_entity_data_camera.NewCameraDataBinding()
	entry := entity_data_binding.ToDataBinding("", &data)
	entry.Gen.RegisterKey = engine_entity_data_camera.BindingKey()
	return &entry
}
