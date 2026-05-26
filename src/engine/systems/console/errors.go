/******************************************************************************/
/* errors.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package console

import "errors"

var (
	ErrCommandNotFound = errors.New("the command with the given key does not exist")
)
