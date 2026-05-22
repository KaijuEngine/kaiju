/******************************************************************************/
/* default_css.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package css

import _ "embed"

//go:embed default.css
var DefaultCSS string

var OverrideCSS string
