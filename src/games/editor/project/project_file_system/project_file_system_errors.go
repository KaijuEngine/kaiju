package project_file_system

import "fmt"

type PathError struct {
	Path string
	Msg  string
	Err  error
}

func (e PathError) Error() string {
	if e.Msg == "" {
		return fmt.Sprintf("(%s): %v", e.Path, e.Err)
	}
	return fmt.Sprintf("%s (%s): %v", e.Msg, e.Path, e.Err)
}
