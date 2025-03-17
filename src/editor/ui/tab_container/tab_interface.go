package tab_container

import (
	"kaiju/engine/ui/markup/document"
	"kaiju/engine/ui"
)

type TabContent interface {
	TabTitle() string
	Document() *document.Document
	Reload(uiMan *ui.Manager, root *document.Element)
	Destroy()
}
