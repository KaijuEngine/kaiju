package klib

type ErrorList struct {
	Errors []error
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
