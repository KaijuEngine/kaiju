package tab_container

import (
	"kaiju/markup/document"
	"kaiju/ui"
)

type TabContent interface {
	TabTitle() string
	Document() *document.Document
	Reload(uiMan *ui.Manager, root *document.Element)
	Destroy()
	Show()
	Hide()
}
