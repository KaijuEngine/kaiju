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

func NewHost() (Host, error) {
	win, err := windowing.New("Kaiju")
	if err != nil {
		return Host{}, err
	}
	return Host{
		entities:    make([]*Entity, 0),
		frameTime:   0,
		Closing:     false,
		Updater:     NewUpdater(),
		LateUpdater: NewUpdater(),
		Window:      win,
	}, nil
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
