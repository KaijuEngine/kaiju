package engine

import (
	"kaiju/assets"
	"kaiju/cameras"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/windowing"
)

type Host struct {
	entities      []*Entity
	Window        *windowing.Window
	Camera        *cameras.StandardCamera
	UICamera      *cameras.StandardCamera
	shaderCache   rendering.ShaderCache
	textureCache  rendering.TextureCache
	meshCache     rendering.MeshCache
	fontCache     rendering.FontCache
	Drawings      rendering.Drawings
	frameTime     float64
	Closing       bool
	Updater       Updater
	LateUpdater   Updater
	assetDatabase assets.Database
}

func NewHost() (Host, error) {
	win, err := windowing.New("Kaiju")
	if err != nil {
		return Host{}, err
	}
	host := Host{
		entities:      make([]*Entity, 0),
		frameTime:     0,
		Closing:       false,
		Updater:       NewUpdater(),
		LateUpdater:   NewUpdater(),
		Window:        win,
		assetDatabase: assets.NewDatabase(),
		Camera:        cameras.NewStandardCamera(float32(win.Width()), float32(win.Height()), matrix.Vec3{0, 0, 1}),
		UICamera:      cameras.NewStandardCameraOrthographic(float32(win.Width()), float32(win.Height()), matrix.Vec3{0, 0, 1}),
		Drawings:      rendering.NewDrawings(),
	}
	uiPos := matrix.Vec3{matrix.Float(host.Window.Width()) * 0.5, matrix.Float(host.Window.Height()) * 0.5, 250}
	host.UICamera.SetPosition(uiPos)
	host.shaderCache = rendering.NewShaderCache(host.Window.Renderer, &host.assetDatabase)
	host.textureCache = rendering.NewTextureCache(host.Window.Renderer, &host.assetDatabase)
	host.meshCache = rendering.NewMeshCache(host.Window.Renderer, &host.assetDatabase)
	host.fontCache = rendering.NewFontCache(host.Window.Renderer, &host.assetDatabase)
	return host, nil
}

func (host *Host) ShaderCache() *rendering.ShaderCache   { return &host.shaderCache }
func (host *Host) TextureCache() *rendering.TextureCache { return &host.textureCache }
func (host *Host) MeshCache() *rendering.MeshCache       { return &host.meshCache }
func (host *Host) FontCache() *rendering.FontCache       { return &host.fontCache }
func (host *Host) AssetDatabase() *assets.Database       { return &host.assetDatabase }

func (host *Host) AddEntity(entity *Entity) {
	host.entities = append(host.entities, entity)
}

func (host *Host) AddEntities(entities ...*Entity) {
	host.entities = append(host.entities, entities...)
}

func (host *Host) Entities() []*Entity { return host.entities }

func (host *Host) NewEntity() *Entity {
	entity := NewEntity()
	host.AddEntity(entity)
	return entity
}

func (host *Host) Update(deltaTime float64) {
	host.Window.Poll()
	host.Updater.Update(deltaTime)
	host.LateUpdater.Update(deltaTime)
	if host.Window.IsClosed() || host.Window.IsCrashed() {
		host.Closing = true
	}
	//gl.ClearScreen()
	//host.Window.SwapBuffers()
	// TODO:  Do end updates on various systems
	host.Window.EndUpdate()
}

func (host *Host) Render() {
	host.shaderCache.CreatePending()
	host.textureCache.CreatePending()
	host.meshCache.CreatePending()
	host.Window.Renderer.ReadyFrame(host.Camera, host.UICamera, float32(host.Runtime()))
	host.Drawings.Render(host.Window.Renderer)
	host.Window.SwapBuffers()
	// TODO:  Thread this or make the dirty on demand, and have a flag for the dirty frame
	for _, e := range host.entities {
		e.Transform.ResetDirty()
	}
}

func (host *Host) Runtime() float64 {
	return host.frameTime
}
