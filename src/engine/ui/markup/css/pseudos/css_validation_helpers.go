/******************************************************************************/
/* css_validation_helpers.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import "kaijuengine.com/engine/ui/markup/document"

func isValidationControl(elm *document.Element) bool {
	return elm.IsInput() || elm.IsTextArea()
}
