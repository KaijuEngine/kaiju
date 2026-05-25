/******************************************************************************/
/* render_target_test.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"testing"
)

func TestRenderTargetCreationStoresOptionsAndSize(t *testing.T) {
	manager := NewRenderTargetManager()
	options := RenderTargetOptions{
		Name:        "game",
		Width:       640,
		Height:      360,
		ResizeMode:  RenderTargetResizeModeMatchWindow,
		ColorFormat: GPUFormatB8g8r8a8Unorm,
		Depth:       true,
	}
	target, err := manager.Create(options)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if target.Options() != options {
		t.Fatalf("Options = %+v, want %+v", target.Options(), options)
	}
	if target.Width() != options.Width || target.Height() != options.Height {
		t.Fatalf("size = %dx%d, want %dx%d", target.Width(), target.Height(), options.Width, options.Height)
	}
	if got, ok := manager.Target(options.Name); !ok || got != target {
		t.Fatalf("Target did not return the created render target")
	}
}

func TestRenderTargetResizeNoOpsWhenSizeUnchanged(t *testing.T) {
	target := mustCreateRenderTarget(t, RenderTargetOptions{
		Name:   "viewport",
		Width:  320,
		Height: 200,
	})
	if target.Resize(320, 200) {
		t.Fatalf("Resize returned true for unchanged dimensions")
	}
	if target.ResizeDirty() {
		t.Fatalf("ResizeDirty was set for unchanged dimensions")
	}
}

func TestRenderTargetResizeMarksDirtyOnlyWhenDimensionsChange(t *testing.T) {
	target := mustCreateRenderTarget(t, RenderTargetOptions{
		Name:   "viewport",
		Width:  320,
		Height: 200,
	})
	if target.ResizeDirty() {
		t.Fatalf("new target should not start resize-dirty")
	}
	if !target.Resize(512, 256) {
		t.Fatalf("Resize returned false for changed dimensions")
	}
	if target.Width() != 512 || target.Height() != 256 {
		t.Fatalf("size = %dx%d, want 512x256", target.Width(), target.Height())
	}
	if !target.ResizeDirty() {
		t.Fatalf("ResizeDirty was not set after changed dimensions")
	}
	if target.Resize(512, 256) {
		t.Fatalf("Resize returned true for unchanged dimensions after resize")
	}
}

func TestRenderTargetTextureColorBeforeRealizationReturnsError(t *testing.T) {
	target := mustCreateRenderTarget(t, RenderTargetOptions{
		Name:   "viewport",
		Width:  320,
		Height: 200,
	})
	tex, err := target.Texture(RenderTargetOutputColor)
	if tex != nil {
		t.Fatalf("Texture returned %#v before realization, want nil", tex)
	}
	if !errors.Is(err, ErrRenderTargetNotRealized) {
		t.Fatalf("Texture error = %v, want ErrRenderTargetNotRealized", err)
	}
}

func TestRenderTargetDestroyRemovesAndMarksDestroyedOnPendingProcess(t *testing.T) {
	manager := NewRenderTargetManager()
	target, err := manager.Create(RenderTargetOptions{
		Name:   "viewport",
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := manager.Destroy("viewport"); err != nil {
		t.Fatalf("Destroy returned error: %v", err)
	}
	if _, ok := manager.Target("viewport"); ok {
		t.Fatalf("destroyed target remained in manager lookup")
	}
	if target.Destroyed() {
		t.Fatalf("target should wait for pending render-thread processing before marking destroyed")
	}
	manager.ProcessPending(nil)
	if !target.Destroyed() {
		t.Fatalf("target was not marked destroyed after pending processing")
	}
}

func mustCreateRenderTarget(t *testing.T, options RenderTargetOptions) *RenderTarget {
	t.Helper()
	manager := NewRenderTargetManager()
	target, err := manager.Create(options)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	return target
}
