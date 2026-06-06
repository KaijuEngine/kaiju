//go:build masterServer

/******************************************************************************/
/* bootstrap_master_server.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package bootstrap

import (
	"time"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/systems/logging"
	"kaijuengine.com/network/master_server"
)

func bootstrapInternal(*logging.LogStream) {
	updater := engine.NewUpdater()
	_, err := master_server.New(&updater)
	if err != nil {
		panic(err)
	}
	lastTime := time.Now()
	for {
		since := time.Since(lastTime)
		deltaTime := since.Seconds()
		lastTime = time.Now()
		updater.Update(deltaTime)
		time.Sleep(time.Millisecond * 16)
	}
}
