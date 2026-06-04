/******************************************************************************/
/* combined_target_cache_test.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"

	"kaijuengine.com/engine/assets"
)

type combinedTargetTestCaches struct {
	meshCache *MeshCache
}

func (c combinedTargetTestCaches) ShaderCache() *ShaderCache     { return nil }
func (c combinedTargetTestCaches) TextureCache() *TextureCache   { return nil }
func (c combinedTargetTestCaches) MeshCache() *MeshCache         { return c.meshCache }
func (c combinedTargetTestCaches) FontCache() *FontCache         { return nil }
func (c combinedTargetTestCaches) MaterialCache() *MaterialCache { return nil }
func (c combinedTargetTestCaches) AssetDatabase() assets.Database {
	return nil
}

func TestCombinedTargetDrawCacheKeepsAlternatingSignatures(t *testing.T) {
	meshCache := NewMeshCache(nil, nil)
	device := &GPUDevice{}
	device.Painter.caches = combinedTargetTestCaches{meshCache: &meshCache}
	combineMat := &Material{
		Id:         "combine",
		Instances:  make(map[string]*Material),
		renderPass: &RenderPass{},
	}
	basePosition := &Texture{Key: "position"}
	baseNormal := &Texture{Key: "normal"}
	specA := []combinedTargetSpec{{
		sort:     1,
		color:    &Texture{Key: "color-a"},
		position: basePosition,
		normal:   baseNormal,
	}}
	specB := []combinedTargetSpec{{
		sort:     1,
		color:    &Texture{Key: "color-b"},
		position: basePosition,
		normal:   baseNormal,
	}}

	var cache combinedTargetDrawCache
	entryA, err := cache.Prepare(device, combinedTargetSignature(specA), specA, combineMat, combinedDrawingCuller{})
	if err != nil {
		t.Fatalf("Prepare A returned error: %v", err)
	}
	entryB, err := cache.Prepare(device, combinedTargetSignature(specB), specB, combineMat, combinedDrawingCuller{})
	if err != nil {
		t.Fatalf("Prepare B returned error: %v", err)
	}
	entryAAgain, err := cache.Prepare(device, combinedTargetSignature(specA), specA, combineMat, combinedDrawingCuller{})
	if err != nil {
		t.Fatalf("Prepare A again returned error: %v", err)
	}

	if entryA == nil || entryB == nil {
		t.Fatalf("prepared entries must not be nil")
	}
	if entryA == entryB {
		t.Fatalf("different combined target signatures reused the same entry")
	}
	if entryAAgain != entryA {
		t.Fatalf("alternating back to a prior signature rebuilt the combine drawings")
	}
	if cache.EntryCount() != 2 {
		t.Fatalf("entry count = %d, want 2", cache.EntryCount())
	}
}

func TestCombinedTargetSignatureDistinguishesSameKeyTexturePointers(t *testing.T) {
	sharedPosition := &Texture{Key: "position"}
	sharedNormal := &Texture{Key: "normal"}
	specA := []combinedTargetSpec{{
		sort:     1,
		color:    &Texture{Key: "same-key"},
		position: sharedPosition,
		normal:   sharedNormal,
	}}
	specB := []combinedTargetSpec{{
		sort:     1,
		color:    &Texture{Key: "same-key"},
		position: sharedPosition,
		normal:   sharedNormal,
	}}

	if combinedTargetSignature(specA) == combinedTargetSignature(specB) {
		t.Fatalf("signature should include texture identity, not only texture key")
	}
}

func TestCombinedTargetDrawsIgnoreRenderViewCameraCulling(t *testing.T) {
	meshCache := NewMeshCache(nil, nil)
	device := &GPUDevice{}
	device.Painter.caches = combinedTargetTestCaches{meshCache: &meshCache}
	combineMat := &Material{
		Id:         "combine",
		Instances:  make(map[string]*Material),
		renderPass: &RenderPass{},
	}
	specs := []combinedTargetSpec{{
		sort:     1,
		color:    &Texture{Key: "color"},
		position: &Texture{Key: "position"},
		normal:   &Texture{Key: "normal"},
	}}

	var cache combinedTargetDrawCache
	entry, err := cache.Prepare(device, combinedTargetSignature(specs), specs, combineMat, combinedDrawingCuller{})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	groups := entry.InstanceGroups()
	if len(groups) != 1 {
		t.Fatalf("instance group count = %d, want 1", len(groups))
	}
	if groups[0].EffectiveLayer() != RenderLayerUI {
		t.Fatalf("combine group layer = %v, want UI", groups[0].EffectiveLayer())
	}

	view := newRenderView(RenderViewOptions{
		Name:   "stage",
		Camera: &testViewCuller{inView: false, viewChanged: true},
	}, 0)
	groups[0].CaptureDataForView(LightsForRender{}, newRenderViewFrame(view))
	if groups[0].VisibleCountForView(view) != 1 {
		t.Fatalf("combine quad was culled by the render view camera")
	}
}
