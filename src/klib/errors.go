/******************************************************************************/
/* errors.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

type ErrorList struct {
	Errors []error
}

func ErrorIs[T error](err error) bool {
	_, ok := err.(T)
	return ok
}

func NewErrorList() ErrorList {
	return ErrorList{}
}

func (e *ErrorList) Any() bool {
	return len(e.Errors) > 0
}

func (e *ErrorList) First() error {
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}
	return nil
}

func (e *ErrorList) AddAny(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

func Check(outError *error, newError error) bool {
	*outError = newError
	return newError != nil
}

func CheckAll(res bool) {}
