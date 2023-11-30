package engine

import (
	"kaiju/assets"
	"kaiju/rendering"
	"kaiju/windowing"
)

type Host struct {
	entities      []*Entity
	Window        *windowing.Window
	ShaderCache   rendering.ShaderCache
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
	}
	host.ShaderCache = rendering.NewShaderCache(host.Window.Renderer, &host.assetDatabase)
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
	//gl.ClearScreen()
	//host.Window.SwapBuffers()
	// TODO:  Do end updates on various systems
}
