/******************************************************************************/
/* post_window_create_caller.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package assets

type PostWindowCreateHandle interface {
	ReadApplicationAsset(path string) ([]byte, error)
}
