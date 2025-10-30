package editor_stage_manager

import "fmt"

type StageAlreadyExistsError struct {
	Id string
}

func (e StageAlreadyExistsError) Error() string {
	return fmt.Sprintf("the stage with id '%s' already exists", e.Id)
}
