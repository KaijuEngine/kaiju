/******************************************************************************/
/* shader_cache_test.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"
	"unsafe"
)

func TestShaderCacheReloadQueuesDestroyForCreatePending(t *testing.T) {
	cache := NewShaderCache(nil, nil)
	shader := NewShader(ShaderDataCompiled{Name: "test"})
	shader.RenderId = ShaderId{
		graphicsPipeline: GPUPipeline{GPUHandle{handle: unsafe.Pointer(&testReadyMeshHandle)}},
	}
	cache.shaders[shader.data.Name] = shader

	cache.ReloadShader(ShaderDataCompiled{Name: "test"})

	if len(cache.pendingDestroy) != 1 {
		t.Fatalf("pending destroy count = %d, want 1", len(cache.pendingDestroy))
	}
	if shader.RenderId.IsValid() {
		t.Fatalf("reload should clear the live render id before re-create")
	}
	if len(cache.pendingShaders) != 1 || cache.pendingShaders[0] != shader {
		t.Fatalf("reload should queue the shader for pending creation")
	}
}
