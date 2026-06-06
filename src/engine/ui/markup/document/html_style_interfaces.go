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
}
