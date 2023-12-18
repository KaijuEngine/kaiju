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
	ShaderCache   rendering.ShaderCache
	TextureCache  rendering.TextureCache
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
		Drawings:      rendering.NewDrawings(),
	}
	host.ShaderCache = rendering.NewShaderCache(host.Window.Renderer, &host.assetDatabase)
	host.TextureCache = rendering.NewTextureCache(host.Window.Renderer, &host.assetDatabase)
	return host, nil
}

func (host *Host) Update(deltaTime float64) {
	host.Window.Poll()
	host.Updater.Update(deltaTime)
	host.LateUpdater.Update(deltaTime)
	if host.Window.IsClosed() || host.Window.IsCrashed() {
		host.Closing = true
	}
	host.ShaderCache.CreatePending()
	host.TextureCache.CreatePending()
	//gl.ClearScreen()
	//host.Window.SwapBuffers()
	// TODO:  Do end updates on various systems
}

func (host *Host) Render() {
	host.Window.Renderer.ReadyFrame(host.Camera, float32(host.Runtime()))
	host.Drawings.Render(host.Window.Renderer)
	host.Window.SwapBuffers()
}

func (host Host) Runtime() float64 {
	return host.frameTime
}
