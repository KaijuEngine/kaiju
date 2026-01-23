package supporters

import (
	"encoding/json"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"net/http"
	"slices"
)

type Supporters struct {
	doc     *document.Document
	uiMan   ui.Manager
	onClose func()
}

type Supporter struct {
	Name    string
	GitHub  string
	Website string
	Support int
}

func Show(host *engine.Host, onClose func()) (*Supporters, error) {
	defer tracing.NewRegion("supporters.Show").End()
	o := &Supporters{onClose: onClose}
	o.uiMan.Init(host)
	supporters := readSupporterList()
	slices.SortFunc(supporters, func(a, b Supporter) int {
		return b.Support - a.Support
	})
	var err error
	o.doc, err = markup.DocumentFromHTMLAsset(&o.uiMan, "editor/ui/overlay/supporters_overlay.go.html",
		supporters, map[string]func(*document.Element){
			"missClose": func(*document.Element) { o.Close() },
			"clickLink": clickLink,
		})
	if err != nil {
		return o, err
	}
	return o, err
}

func (o *Supporters) Close() {
	defer tracing.NewRegion("Supporters.Close").End()
	o.doc.Destroy()
	if o.onClose == nil {
		slog.Warn("onClose was not set on the Supporters")
		return
	}
	o.onClose()
}

func clickLink(e *document.Element) {
	defer tracing.NewRegion("supporters.clickSupport").End()
	klib.OpenWebsite(e.Attribute("data-path"))
}

func readSupporterList() []Supporter {
	const url = `https://raw.githubusercontent.com/KaijuEngine/kaiju/refs/heads/master/sponsors.json`
	supporters := []Supporter{}
	response, err := http.Get(url)
	if err != nil {
		slog.Error("could not communicate with GitHub for supporter list")
		return supporters
	}
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&supporters); err != nil {
		slog.Error("could not deserialize list of supporters from GitHub")
		return supporters
	}
	return supporters
}
