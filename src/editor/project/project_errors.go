/******************************************************************************/
/* project_errors.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import "fmt"

type ConfigLoadError struct {
	Err error
}

func (e ConfigLoadError) Error() string {
	return fmt.Sprintf("failed to load the project configuration file: %v", e.Err)
}

type ProjectOpenError struct {
	Path      string
	IsFile    bool
	IsMissing bool
}

func (e ProjectOpenError) Error() string {
	if e.IsFile {
		return fmt.Sprintf("the path specified is a file, not a folder: %s", e.Path)
	} else if e.IsMissing {
		return fmt.Sprintf("the path specified is missing: %s", e.Path)
	} else {
		return fmt.Sprintf("the path specified is not a Kaiju project: %s", e.Path)
	}
}
