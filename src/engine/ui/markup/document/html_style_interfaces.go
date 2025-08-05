package document

import (
	"kaiju/engine"
	"kaiju/engine/ui/markup/css/rules"
)

type Stylizer interface {
	ApplyStyles(s rules.StyleSheet, doc *Document, host *engine.Host)
}
