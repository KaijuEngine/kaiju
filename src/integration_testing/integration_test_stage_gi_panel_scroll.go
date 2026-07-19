//go:build editor

/******************************************************************************/
/* integration_test_stage_gi_panel_scroll.go                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
)

var stageGIPanelTestUIMan *ui.Manager

func init() {
	tests["stage-gi-panel-scroll"] = IntegrationTestStageGIPanelScroll
}

func IntegrationTestStageGIPanelScroll(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	stageGIPanelTestUIMan = &uiMan
	noop := func(*document.Element) {}
	funcs := map[string]func(*document.Element){
		"giOverrideChanged": noop, "giRuntimeChanged": noop,
		"selectGIProbe": noop, "clearGIProbe": noop,
		"giProbeDrop": noop, "giProbeDragEnter": noop, "giProbeDragExit": noop,
		"giBakeSettingChanged": noop, "bakeGI": noop, "bakeGIAs": noop,
		"cancelGIBake": noop,
	}
	doc, err := markup.DocumentFromHTMLAsset(&uiMan,
		"editor/ui/workspace/stage_workspace_gi.go.html", nil, funcs)
	if err != nil {
		slog.Error("failed to load Stage GI panel", "error", err)
		os.Exit(1)
	}
	area, ok := doc.GetElementById("giArea")
	if !ok {
		slog.Error("Stage GI panel is missing #giArea")
		os.Exit(1)
	}
	panel := area.UI.ToPanel()
	host.RunAfterFrames(12, func() {
		if panel.ScrollDirection()&ui.PanelScrollDirectionVertical == 0 {
			slog.Error("Stage GI panel does not accept vertical scrolling")
			os.Exit(1)
		}
		maxScroll := panel.MaxScroll().Y()
		if maxScroll <= 0 {
			slog.Error("Stage GI panel has no vertical scroll extent", "maxScroll", maxScroll)
			os.Exit(1)
		}
		panel.SetScrollY(maxScroll)
		host.RunAfterFrames(3, func() {
			if panel.ScrollY() <= 0 {
				slog.Error("Stage GI panel did not apply a vertical scroll request",
					"scrollY", panel.ScrollY(), "maxScroll", maxScroll)
				os.Exit(1)
			}
			slog.Info("Stage GI panel scroll verified",
				"scrollY", panel.ScrollY(), "maxScroll", maxScroll)
			os.Exit(0)
		})
	})
}
