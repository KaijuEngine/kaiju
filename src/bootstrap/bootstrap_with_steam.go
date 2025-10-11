//go:build steam

package bootstrap

import (
	"kaiju/engine"
	"kaiju/platform/steam"
)

func initExternalGameService() {
	steam.Initialize()
}

func initExternalGameServiceRuntime(host *engine.Host) {
	if steam.IsInitialized() {
		sid := host.Updater.AddUpdate(func(f float64) { steam.RunCallbacks() })
		host.OnClose.Add(func() { host.Updater.RemoveUpdate(sid) })
	}
}

func terminateExternalGameService() {
	steam.Shutdown()
}
