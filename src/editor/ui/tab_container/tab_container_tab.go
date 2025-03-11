package tab_container

import (
	"kaiju/markup/document"
	"weak"
)

type TabContainerTab struct {
	Id      string
	Label   string
	parent  weak.Pointer[TabContainer]
	content TabContent
}

func NewTab(content TabContent) TabContainerTab {
	return TabContainerTab{
		Label:   content.TabTitle(),
		content: content,
	}
}

func (t *TabContainerTab) DragUpdate() {}

func (t *TabContainerTab) Reload(root *document.Element) {
	t.parent.Value().host.CreatingEditorEntities()
	t.content.Reload(root)
	t.parent.Value().host.DoneCreatingEditorEntities()
}

func (t *TabContainerTab) Destroy() { t.content.Destroy() }
func (t *TabContainerTab) Show()    { t.content.Show() }
func (t *TabContainerTab) Hide()    { t.content.Hide() }
