package input_prompt

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"log/slog"
)

type InputPrompt struct {
	doc    *document.Document
	uiMan  ui.Manager
	config InputPromptConfig
	input  *document.Element
}

type InputPromptConfig struct {
	Title       string
	Description string
	Placeholder string
	Value       string
	ConfirmText string
	CancelText  string
	OnConfirm   func(string)
	OnCancel    func()
}

func Show(host *engine.Host, config InputPromptConfig) (*InputPrompt, error) {
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
	return ip, err
}

func (ip *InputPrompt) Close() { ip.doc.Destroy() }

func (ip *InputPrompt) confirm(e *document.Element) {
	txt := ip.input.UI.ToInput().Text()
	ip.Close()
	if ip.config.OnConfirm == nil {
		slog.Error("the input prompt didn't have a OnConfirm set, nothing to do")
		return
	}
	ip.config.OnConfirm(txt)
}

func (ip *InputPrompt) cancel(e *document.Element) {
	ip.Close()
	if ip.config.OnCancel != nil {
		ip.config.OnCancel()
	}
}
