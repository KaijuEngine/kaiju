package content_previews

import (
	"bytes"
	"image"
	"image/png"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
	"kaiju/engine/cameras"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const (
	sphereRadius   = 1
	sphereSegments = 32
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
	p.cam = cameras.NewStandardCamera(float32(rp.Width()), float32(rp.Height()),
		float32(rp.Width()), float32(rp.Height()), matrix.Vec3Zero())
	p.cam.SetPositionAndLookAt(pos, matrix.Vec3Zero())
	p.sphereTransform.Initialize(p.ed.Host().WorkGroup())
	return nil
}

func (p *ContentPreviewer) GeneratePreviews(ids []string) {
	defer tracing.NewRegion("ContentPreviewer.GeneratePreviews").End()
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.pending = append(p.pending, ids...)
	if p.inProc {
		return
	}
	// goroutine
	go p.nextPreview()
}

func (p *ContentPreviewer) DeletePreviewImage(id string) error {
	defer tracing.NewRegion("ContentPreviewer.DeletePreviewImage").End()
	path := filepath.Join(project_file_system.EditorCacheContentPreviews, id)
	return p.ed.ProjectFileSystem().Remove(path)
}

func (p *ContentPreviewer) LoadPreviewImage(id string) (*rendering.Texture, error) {
	defer tracing.NewRegion("ContentPreviewer.LoadPreviewImage").End()
	path := filepath.Join(project_file_system.EditorCacheContentPreviews, id)
	data, err := p.ed.ProjectFileSystem().ReadFile(path)
	if err != nil {
		return nil, err
	}
	host := p.ed.Host()
	texKey := "preview_" + id
	const filter = rendering.TextureFilterLinear
	tex, err := host.TextureCache().Texture(texKey, filter)
	if tex == nil || err != nil {
		if tex, err = rendering.NewTextureFromImage(texKey, data, filter); err != nil {
			return nil, err
		}
		host.TextureCache().InsertTexture(tex)
	}
	return tex, nil
}

func (p *ContentPreviewer) previewExists(id string) bool {
	defer tracing.NewRegion("ContentPreviewer.previewExists").End()
	path := filepath.Join(project_file_system.EditorCacheContentPreviews, id)
	return p.ed.ProjectFileSystem().Exists(path)
}

func (p *ContentPreviewer) nextPreview() {
	defer tracing.NewRegion("ContentPreviewer.nextPreview").End()
	id := ""
	p.mutex.Lock()
	if p.inProc {
		p.mutex.Unlock()
		return
	}
	if len(p.pending) > 0 {
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
	cc, err := p.ed.Cache().Read(id)
	if err != nil {
		slog.Error("failed to read the cache for content", "id", id, "error", err)
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
	// case content_database.Material{}.TypeName():
	// 	p.renderMaterial(id)
	default:
		p.completeProc()
	}
}

func (p *ContentPreviewer) writePreviewFile(id string, data []byte) error {
	defer tracing.NewRegion("ContentPreviewer.writePreviewFile").End()
	pfs := p.ed.ProjectFileSystem()
	dir := project_file_system.EditorCacheContentPreviews
	pfs.MkdirAll(dir, os.ModePerm)
	return pfs.WriteFile(filepath.Join(dir, id), data, os.ModePerm)
}

func (p *ContentPreviewer) readRenderPass(host *engine.Host, sd rendering.DrawInstance, id string) {
	defer p.completeProc()
	pixels, err := p.mat.RenderPass().Texture(0).ReadAllPixels(host.Window.Renderer)
	sd.Destroy()
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
