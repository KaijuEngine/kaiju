package sponsors

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

type Sponsors struct {
	doc     *document.Document
	uiMan   ui.Manager
	onClose func()
}

type Sponsor struct {
	Name    string
	GitHub  string
	Website string
	Support int
}

func Show(host *engine.Host, onClose func()) (*Sponsors, error) {
	defer tracing.NewRegion("sponsors.Show").End()
	o := &Sponsors{onClose: onClose}
	o.uiMan.Init(host)
	sponsors := readSponsorList()
	slices.SortFunc(sponsors, func(a, b Sponsor) int {
		return b.Support - a.Support
	})
	var err error
	o.doc, err = markup.DocumentFromHTMLAsset(&o.uiMan, "editor/ui/overlay/sponsors_overlay.go.html",
		sponsors, map[string]func(*document.Element){
			"missClose": func(*document.Element) { o.Close() },
			"clickLink": clickLink,
		})
	if err != nil {
		return o, err
	}
	return o, err
}

func (o *Sponsors) Close() {
	defer tracing.NewRegion("Sponsors.Close").End()
	o.doc.Destroy()
	if o.onClose == nil {
		slog.Warn("onClose was not set on the Sponsors")
		return
	}
	o.onClose()
}

func clickLink(e *document.Element) {
	defer tracing.NewRegion("sponsors.clickLink").End()
	klib.OpenWebsite(e.Attribute("data-path"))
}

func readSponsorList() []Sponsor {
	const url = `https://raw.githubusercontent.com/KaijuEngine/kaiju/refs/heads/master/sponsors.json`
	sponsors := []Sponsor{}
	response, err := http.Get(url)
	if err != nil {
		slog.Error("could not communicate with GitHub for sponsor list")
		return sponsors
	}
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&sponsors); err != nil {
		slog.Error("could not deserialize list of sponsors from GitHub")
		return sponsors
	}
	return sponsors
}
