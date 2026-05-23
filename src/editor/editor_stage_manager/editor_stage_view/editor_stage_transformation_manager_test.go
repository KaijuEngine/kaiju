/******************************************************************************/
/* editor_stage_transformation_manager_test.go                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"testing"

	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/platform/hid"
)

func TestTransformationManagerToolHotkeysDefault(t *testing.T) {
	tm := TransformationManager{}
	translate, rotate, scale := tm.toolHotkeys()
	if translate != hid.KeyboardKeyG || rotate != hid.KeyboardKeyR || scale != hid.KeyboardKeyS {
		t.Fatalf("expected G/R/S hotkeys, got %d/%d/%d", translate, rotate, scale)
	}
}

func TestTransformationManagerToolHotkeysWER(t *testing.T) {
	settings := editor_settings.Settings{UseWERTransformHotkeys: true}
	tm := TransformationManager{settings: &settings}
	translate, rotate, scale := tm.toolHotkeys()
	if translate != hid.KeyboardKeyW || rotate != hid.KeyboardKeyE || scale != hid.KeyboardKeyR {
		t.Fatalf("expected W/E/R hotkeys, got %d/%d/%d", translate, rotate, scale)
	}
}
