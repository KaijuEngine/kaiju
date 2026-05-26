/******************************************************************************/
/* table_of_contents_overlay.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package table_of_contents_overlay

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets/table_of_contents"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
)

type TableOfContentsOverlay struct {
	doc     *document.Document
	uiMan   ui.Manager
	config  Config
	changed bool
}

type Config struct {
	TOC       table_of_contents.TableOfContents
	OnChanged func(toc table_of_contents.TableOfContents)
	OnClose   func()
}

func Show(host *engine.Host, config Config) (*TableOfContentsOverlay, error) {
	defer tracing.NewRegion("table_of_contents_overlay.Show").End()
	o := &TableOfContentsOverlay{
		config: config,
	}
	o.uiMan.Init(host)
	var err error
	o.doc, err = markup.DocumentFromHTMLAsset(&o.uiMan, "editor/ui/overlay/table_of_contents.go.html",
		config.TOC, map[string]func(*document.Element){
			"clickRemove": o.clickRemove,
			"clickMiss":   o.clickMiss,
		})
	if err != nil {
		return o, err
	}
	return o, err
}

func (o *TableOfContentsOverlay) Close() {
	defer tracing.NewRegion("TableOfContentsOverlay.Close").End()
	o.doc.Destroy()
}

func (o *TableOfContentsOverlay) clickRemove(e *document.Element) {
	defer tracing.NewRegion("TableOfContentsOverlay.clickRemove").End()
	o.config.TOC.Remove(e.Attribute("id"))
	o.changed = true
	o.doc.RemoveElement(e.Parent.Value())
}

func (o *TableOfContentsOverlay) clickMiss(*document.Element) {
	defer tracing.NewRegion("TableOfContentsOverlay.clickMiss").End()
	o.Close()
	if o.changed && o.config.OnChanged != nil {
		o.config.OnChanged(o.config.TOC)
	}
	if o.config.OnClose != nil {
		o.config.OnClose()
	}
}
