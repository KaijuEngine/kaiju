/******************************************************************************/
/* confirm_prompt_overlay.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package confirm_prompt

import (
	"log/slog"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
)

type ConfirmPrompt struct {
	doc    *document.Document
	uiMan  ui.Manager
	config Config
}

type Config struct {
	Title       string
	Description string
	ConfirmText string
	CancelText  string
	OnConfirm   func()
	OnCancel    func()
}

func Show(host *engine.Host, config Config) (*ConfirmPrompt, error) {
	defer tracing.NewRegion("confirm_prompt.Show").End()
	ip := &ConfirmPrompt{config: config}
	ip.uiMan.Init(host)
	var err error
	ip.doc, err = markup.DocumentFromHTMLAsset(&ip.uiMan, "editor/ui/overlay/confirm_prompt.go.html",
		ip.config, map[string]func(*document.Element){
			"confirm": ip.confirm,
			"cancel":  ip.cancel,
		})
	if err != nil {
		return ip, err
	}
	return ip, err
}

func (ip *ConfirmPrompt) Close() {
	defer tracing.NewRegion("ConfirmPrompt.Close").End()
	ip.doc.Destroy()
}

func (ip *ConfirmPrompt) confirm(e *document.Element) {
	defer tracing.NewRegion("ConfirmPrompt.confirm").End()
	ip.Close()
	if ip.config.OnConfirm == nil {
		slog.Error("the input prompt didn't have a OnConfirm set, nothing to do")
		return
	}
	ip.config.OnConfirm()
}

func (ip *ConfirmPrompt) cancel(e *document.Element) {
	defer tracing.NewRegion("ConfirmPrompt.cancel").End()
	ip.Close()
	if ip.config.OnCancel != nil {
		ip.config.OnCancel()
	}
}
