//go:build steam

/******************************************************************************/
/* bootstrap_with_steam.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package bootstrap

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/platform/steam"
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
