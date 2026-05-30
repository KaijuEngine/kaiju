/******************************************************************************/
/* editor_action_palette.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"log/slog"

	"kaijuengine.com/editor/editor_overlay/action_palette"
)

func (ed *Editor) showActionPalette() {
	ed.BlurInterface()
	if _, err := action_palette.Show(ed.host, ed.Actions(), ed.FocusInterface); err != nil {
		slog.Error("failed to show action palette", "error", err)
		ed.FocusInterface()
	}
}
