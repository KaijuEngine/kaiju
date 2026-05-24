/******************************************************************************/
/* html_style_interfaces.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package document

import (
	"kaijuengine.com/engine/ui/markup/css/rules"
)

type Stylizer interface {
	ApplyStyles(s rules.StyleSheet, doc *Document)
	// ApplyStylesToElement re-applies CSS rules only to `target` and its
	// descendants. Implementations must clear and re-add rules + event
	// handlers in the subtree without touching elements outside it, so a
	// per-element class/id/parent change does not dirty the whole document.
	ApplyStylesToElement(s rules.StyleSheet, doc *Document, target *Element)
}
