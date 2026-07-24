/******************************************************************************/
/* content_previewer.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_previews

import (
	"bytes"
	"image"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

const (
	contentPreviewCacheVersion = "v2"
	sphereRadius               = 1
	sphereSegments             = 32
)

type ContentPreviewer struct {
	ed              EditorInterface
	pending         []string
	mat             *rendering.Material
	cam             cameras.Camera
	sphereTransform matrix.Transform
	mutex           sync.Mutex
	inProc          bool
}

func (p *ContentPreviewer) Initialize(ed EditorInterface) error {
	defer tracing.NewRegion("ContentPreviewer.Initialize").End()
	p.ed = ed
	mat, err := p.ed.Host().MaterialCache().Material("ed_thumb_preview_mesh.material")
	if err != nil {
		return err
	}
	p.mat = mat
	rp := p.mat.RenderPass()
	lookAt := matrix.NewVec3(0, 0.67, -matrix.Vec3Forward().Z())
	pos := lookAt.Scale(2)
	p.cam = cameras.NewStandardCamera(matrix.Float(rp.Width()), matrix.Float(rp.Height()),
		matrix.Float(rp.Width()), matrix.Float(rp.Height()), matrix.Vec3Zero())
	p.cam.SetPositionAndLookAt(pos, matrix.Vec3Zero())
	p.sphereTransform.Initialize(p.ed.Host().WorkGroup())
	return nil
}

func (p *ContentPreviewer) GeneratePreviews(ids []string) {
	defer tracing.NewRegion("ContentPreviewer.GeneratePreviews").End()
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.pending = append(p.pending, p.expandPreviewIds(ids)...)
	if p.inProc {
		return
	}
	// goroutine
	go p.nextPreview()
}

func (p *ContentPreviewer) expandPreviewIds(ids []string) []string {
	expanded := make([]string, 0, len(ids))
	for _, id := range ids {
		expanded = append(expanded, id)
		ref := kaiju_mesh.ParseMeshRef(id)
		if ref.Key != "" {
			continue
		}
		cc, err := p.ed.Cache().Read(ref.Asset)
		if err != nil || cc.Config.Type != (content_database.Mesh{}).TypeName() ||
			cc.Config.Mesh == nil || len(cc.Config.Mesh.Submeshes) <= 1 {
			continue
		}
		for i := range cc.Config.Mesh.Submeshes {
			submesh := &cc.Config.Mesh.Submeshes[i]
			if submesh.Key != "" && !submesh.Missing {
				expanded = append(expanded, kaiju_mesh.MeshRefString(ref.Asset, submesh.Key))
			}
		}
	}
	return expanded
}

func (p *ContentPreviewer) DeletePreviewImage(id string) error {
	defer tracing.NewRegion("ContentPreviewer.DeletePreviewImage").End()
	path := p.previewPath(id)
	return p.ed.ProjectFileSystem().Remove(path)
}

func (p *ContentPreviewer) LoadPreviewImage(id string) (*rendering.Texture, error) {
	defer tracing.NewRegion("ContentPreviewer.LoadPreviewImage").End()
	path := p.previewPath(id)
	data, err := p.ed.ProjectFileSystem().ReadFile(path)
	if err != nil {
		return nil, err
	}
	host := p.ed.Host()
	texKey := "preview_" + contentPreviewCacheVersion + "_" + id
	const filter = rendering.TextureFilterLinear
	return cachedPreviewTexture(host.TextureCache(), texKey, data, filter)
}

// useCachedTexture attempts to insert the preview image data into the texture cache and returns the texture if successful.
// This allows for efficient reuse of textures across multiple previews and avoids redundant GPU uploads if the texture already exists in the cache.
func (p *ContentPreviewer) useCachedTexture(texKey string, data []byte, filter rendering.TextureFilter) (*rendering.Texture, error) {
	defer tracing.NewRegion("ContentPreviewer.useCachedTexture").End()
	host := p.ed.Host()
	return cachedPreviewTexture(host.TextureCache(), texKey, data, filter)
}

func cachedPreviewTexture(textureCache *rendering.TextureCache, texKey string, data []byte, filter rendering.TextureFilter) (*rendering.Texture, error) {
	if tex, ok := textureCache.Find(texKey, filter); ok {
		return tex, nil
	}
	return textureCache.InsertImageTextureWithPriority(
		texKey, data, filter, rendering.TextureUploadPriorityHigh)
}

func (p *ContentPreviewer) previewExists(id string) bool {
	defer tracing.NewRegion("ContentPreviewer.previewExists").End()
	return p.ed.ProjectFileSystem().Exists(p.previewPath(id))
}

func (p *ContentPreviewer) nextPreview() {
	defer tracing.NewRegion("ContentPreviewer.nextPreview").End()
	id := ""
	p.mutex.Lock()
	if p.inProc {
		p.mutex.Unlock()
		return
	}
	for len(p.pending) > 0 && id == "" {
		id = p.pending[0]
		p.pending = p.pending[1:]
	}
	p.inProc = id != ""
	p.mutex.Unlock()
	if id != "" {
		p.proc(id)
	}
}

func (p *ContentPreviewer) completeProc() {
	defer tracing.NewRegion("ContentPreviewer.completeProc").End()
	p.mutex.Lock()
	p.inProc = false
	p.mutex.Unlock()
	p.nextPreview()
}

func (p *ContentPreviewer) proc(id string) {
	defer tracing.NewRegion("ContentPreviewer.proc").End()
	ref := kaiju_mesh.ParseMeshRef(id)
	cc, err := p.ed.Cache().Read(ref.Asset)
	if err != nil {
		slog.Error("failed to read the cache for content", "id", id, "error", err)
		p.completeProc()
		return
	}
	if p.previewExists(id) {
		p.ed.Events().OnContentPreviewGenerated.Execute(id)
		p.completeProc()
		return
	}
	switch cc.Config.Type {
	case content_database.Mesh{}.TypeName():
		p.renderMesh(id)
	case content_database.Terrain{}.TypeName():
		p.renderTerrain(id)
	// case content_database.Material{}.TypeName():
	// 	p.renderMaterial(id)
	default:
		p.completeProc()
	}
}

func (p *ContentPreviewer) writePreviewFile(id string, data []byte) error {
	defer tracing.NewRegion("ContentPreviewer.writePreviewFile").End()
	pfs := p.ed.ProjectFileSystem()
	path := p.previewPath(id)
	dir := filepath.Dir(path)
	pfs.MkdirAll(dir, os.ModePerm)
	return pfs.WriteFile(path, data, os.ModePerm)
}

func (p *ContentPreviewer) previewPath(id string) string {
	return filepath.Join(project_file_system.EditorCacheContentPreviews, contentPreviewCacheVersion, previewFileName(id))
}

func previewFileName(id string) string {
	replacer := strings.NewReplacer(
		"<", "_", ">", "_", ":", "_", `"`, "_", "/", "_", "\\", "_",
		"|", "_", "?", "_", "*", "_",
	)
	return replacer.Replace(id)
}

func (p *ContentPreviewer) readRenderPassAfterNextRender(host *engine.Host, id string, shaderData ...rendering.DrawInstance) {
	host.RunAfterRender(func(*rendering.GPUDevice, engine.RenderFrame) {
		p.readRenderPass(host, id, shaderData...)
	})
}

func (p *ContentPreviewer) readRenderPass(host *engine.Host, id string, shaderData ...rendering.DrawInstance) {
	defer p.completeProc()
	defer func() {
		for _, sd := range shaderData {
			if sd != nil {
				sd.Destroy()
			}
		}
	}()
	pixels, err := p.mat.RenderPass().Texture(0).ReadAllPixels(&host.Window.GpuHost)
	if err != nil {
		slog.Error("failed to read the mesh preview image from GPU", "id", id, "error", err)
		return
	} else if len(pixels) == 0 {
		slog.Error("failed to read the mesh preview image from GPU, result was empty", "id", id)
		return
	}
	tex := p.mat.RenderPass().Texture(0)
	w, h := tex.Width, tex.Height
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pixels)
	var buf bytes.Buffer
	if err = png.Encode(&buf, img); err != nil {
		slog.Error("failed to encode the pixel buffer from the GPU for the mesh preview image", "id", id, "error", err)
		return
	}
	if err = p.writePreviewFile(id, buf.Bytes()); err != nil {
		slog.Error("failed to write the mesh preview image cache file", "id", id, "error", err)
		return
	}
	p.ed.Events().OnContentPreviewGenerated.Execute(id)
}
