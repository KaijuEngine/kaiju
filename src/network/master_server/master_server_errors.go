/******************************************************************************/
/* master_server_errors.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package master_server

type Error = uint8

const (
	ErrorNone = Error(iota)
	ErrorIncorrectPassword
	ErrorServerDoesntExist
)
