/******************************************************************************/
/* editor_stage_manager_errors.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import "fmt"

type StageAlreadyExistsError struct {
	Id string
}

func (e StageAlreadyExistsError) Error() string {
	return fmt.Sprintf("the stage with id '%s' already exists", e.Id)
}
