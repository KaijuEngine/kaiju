package tab_container

import "kaiju/markup/document"

type TabContent interface {
	Document() *document.Document
	Reload()
	Destroy()
	Show()
	Hide()
}
