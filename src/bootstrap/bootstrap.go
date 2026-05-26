/******************************************************************************/
/* bootstrap.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package bootstrap

import (
	"log/slog"

	"kaijuengine.com/build"
	"kaijuengine.com/engine/systems/logging"
)

func Main(game GameInterface, platformState any) {
	var ops *slog.HandlerOptions = nil
	if !build.Debug {
		ops = &slog.HandlerOptions{
			Level: slog.LevelError,
		}
	}
	logStream := logging.Initialize(ops)
	defer logStream.Close()
	bootstrapInternal(logStream, game, platformState)
}
