package tab_container

import "kaiju/markup/document"

type TabContent interface {
	TabTitle() string
	Document() *document.Document
	Reload(root *document.Element)
	Destroy()
	Show()
	Hide()
}
