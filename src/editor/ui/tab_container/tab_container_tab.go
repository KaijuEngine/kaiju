package tab_container

import (
	"kaiju/markup/document"
	"weak"
)

type TabContainerTab struct {
	Id      int
	Label   string
	parent  weak.Pointer[TabContainer]
	content TabContent
}

func NewTab(label string, content TabContent) TabContainerTab {
	return TabContainerTab{
		Label:   label,
		content: content,
	}
}

func (t *TabContainerTab) DragUpdate() {}

func (t *TabContainerTab) loadDocument(root *document.Element) {
	t.content.Reload()
	doc := t.content.Document()
	bodyPanel := root.UI.ToPanel()
	for i := range doc.TopElements {
		bodyPanel.AddChild(doc.TopElements[i].UI)
	}
}

func (t *TabContainerTab) Destroy() { t.content.Destroy() }
func (t *TabContainerTab) Show()    { t.content.Show() }
func (t *TabContainerTab) Hide()    { t.content.Hide() }
