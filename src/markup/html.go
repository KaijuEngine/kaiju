package markup

import (
	"kaiju/engine"
	"kaiju/markup/css"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func DocumentFromHTML(host *engine.Host, htmlPath, cssPath string, withData any, funcMap map[string]func(*document.DocElement)) *document.Document {
	htmlBytes, err := host.AssetDatabase().ReadText(htmlPath)
	if err != nil {
		panic("Failed to read markup file: " + htmlPath)
	}
	if cssPath == "" {
		return DocumentFromHTMLString(host, string(htmlBytes), "", withData, funcMap)
	} else {
		cssBytes, err := host.AssetDatabase().ReadText(cssPath)
		if err != nil {
			panic("Failed to read css file: " + cssPath)
		}
		return DocumentFromHTMLString(host, string(htmlBytes), string(cssBytes), withData, funcMap)
	}
}

func sizeTexts(doc *document.Document, host *engine.Host) {
	for i := range doc.Elements {
		elm := &doc.Elements[i]
		e := elm.HTML
		if e.IsText() {
			label := elm.UI.(*ui.Label)
			parentWidth := float32(-1.0)
			updateSize := func(l *ui.Layout) {
				if p := ui.FirstOnEntity(label.Entity().Parent); p != nil {
					newParentWidth := p.Layout().PixelSize().Width()
					height := l.PixelSize().Height()
					if newParentWidth != parentWidth {
						parentWidth = newParentWidth
						textSize := host.FontCache().MeasureStringWithin(label.FontFace(),
							e.Data(), label.FontSize(), parentWidth)
						height = textSize.Height()
					}
					l.Scale(parentWidth, height)
				}
			}
			updateSize(label.Layout())
			label.Layout().AddFunction(updateSize)
		}
		height := elm.UI.Layout().PixelSize().Y()
		p := elm.HTML.Parent
		for p != nil && p.DocumentElement != nil {
			pPanel := p.DocumentElement.UIPanel
			if pPanel.FittingContent() && pPanel.Layout().PixelSize().Y() < height {
				pPanel.Layout().ScaleHeight(height)
			}
			p = p.Parent
		}
	}
}

func DocumentFromHTMLString(host *engine.Host, html, cssStr string, withData any, funcMap map[string]func(*document.DocElement)) *document.Document {
	doc := document.DocumentFromHTMLString(host, html, withData, funcMap)
	s := rules.NewStyleSheet()
	s.Parse(css.DefaultCSS)
	s.Parse(cssStr)
	for i := range doc.HeadElements {
		if doc.HeadElements[i].Data() == "style" {
			s.Parse(doc.HeadElements[i].Children[0].Data())
		}
	}
	css.Apply(s, doc, host)
	sizeTexts(doc, host)
	return doc
}
