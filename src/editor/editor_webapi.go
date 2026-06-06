/******************************************************************************/
/* editor_webapi.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"context"
	"log/slog"
	"time"

	"kaijuengine.com/editor/webapi"
)

func (ed *Editor) initializeWebAPI() {
	ed.webAPIServer = webapi.New(ed)
	ed.host.OnClose.Add(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := ed.webAPIServer.Close(ctx); err != nil {
			slog.Error("failed to stop editor Web API server", "error", err)
		}
	})
}

func (ed *Editor) updateWebAPI() {
	if ed.webAPIServer == nil {
		ed.initializeWebAPI()
	}
	if err := ed.webAPIServer.Apply(webapi.Config{
		Enabled: ed.settings.WebAPI.Enabled,
		Port:    ed.settings.WebAPI.Port,
		APIKey:  ed.settings.WebAPI.APIKey,
	}); err != nil {
		slog.Error("failed to configure editor Web API server",
			"address", webapi.Address(ed.settings.WebAPI.Port), "error", err)
	}
}
