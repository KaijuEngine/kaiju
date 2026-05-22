/******************************************************************************/
/* input_prompt_overlay.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package input_prompt

import (
	"log/slog"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
)

type InputPrompt struct {
	doc    *document.Document
	uiMan  ui.Manager
	config Config
	input  *document.Element
}

type Config struct {
	Title       string
	Description string
	Placeholder string
	Value       string
	ConfirmText string
	CancelText  string
	OnConfirm   func(string)
	OnCancel    func()
}

func Show(host *engine.Host, config Config) (*InputPrompt, error) {
	defer tracing.NewRegion("input_prompt.Show").End()
	ip := &InputPrompt{
		config: config,
	}
	ip.uiMan.Init(host)
	var err error
	ip.doc, err = markup.DocumentFromHTMLAsset(&ip.uiMan, "editor/ui/overlay/input_prompt.go.html",
		ip.config, map[string]func(*document.Element){
			"confirm": ip.confirm,
			"cancel":  ip.cancel,
		})
	if err != nil {
		return ip, err
	}
	ip.input, _ = ip.doc.GetElementById("input")
	box := ip.input.UI.ToInput()
	box.Focus()
	box.SelectAll()
	return ip, err
}

func (ip *InputPrompt) Close() {
	defer tracing.NewRegion("InputPrompt.Close").End()
	ip.doc.Destroy()
}

func (ip *InputPrompt) confirm(e *document.Element) {
	defer tracing.NewRegion("InputPrompt.confirm").End()
	txt := ip.input.UI.ToInput().Text()
	ip.Close()
	if ip.config.OnConfirm == nil {
		slog.Error("the input prompt didn't have a OnConfirm set, nothing to do")
		return
	}
	ip.config.OnConfirm(txt)
}

func (ip *InputPrompt) cancel(e *document.Element) {
	defer tracing.NewRegion("InputPrompt.cancel").End()
	ip.Close()
	if ip.config.OnCancel != nil {
		ip.config.OnCancel()
	}
}
