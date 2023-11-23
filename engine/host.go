package engine

import "kaiju/windowing"

type Host struct {
	entities    []*Entity
	Window      *windowing.Window
	frameTime   float64
	Closing     bool
	Updater     Updater
	LateUpdater Updater
}

func NewHost() Host {
	return Host{
		entities:    make([]*Entity, 0),
		frameTime:   0,
		Closing:     false,
		Updater:     NewUpdater(),
		LateUpdater: NewUpdater(),
		Window:      windowing.New("Kaiju"),
	}
}

func (host *Host) Update(deltaTime float64) {
	host.Window.Poll()
	host.Updater.Update(deltaTime)
	host.LateUpdater.Update(deltaTime)
	if host.Window.IsClosed() || host.Window.IsCrashed() {
		host.Closing = true
	}
	// TODO:  Do end updates on various systems
}
